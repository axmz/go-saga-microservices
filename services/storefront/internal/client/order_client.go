package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	httppb "github.com/axmz/go-saga-microservices/pkg/proto/http"
	"google.golang.org/protobuf/proto"
)

type OrderClient interface {
	CreateOrder(ctx context.Context, req *httppb.CreateOrderRequest) (*httppb.CreateOrderResponse, error)
	GetOrder(ctx context.Context, orderID string) (*httppb.GetOrderResponse, error)
}

type HTTPOrderClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPOrderClient(baseURL string) *HTTPOrderClient {
	return &HTTPOrderClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *HTTPOrderClient) CreateOrder(ctx context.Context, req *httppb.CreateOrderRequest) (*httppb.CreateOrderResponse, error) {
	protoData, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(c.baseURL+"/orders", "application/x-protobuf", bytes.NewBuffer(protoData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("order service returned status: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var protoResp httppb.CreateOrderResponse
	if err := proto.Unmarshal(respBody, &protoResp); err != nil {
		return nil, err
	}

	return &protoResp, nil
}

func (c *HTTPOrderClient) GetOrder(ctx context.Context, orderID string) (*httppb.GetOrderResponse, error) {
	resp, err := c.client.Get(c.baseURL + "/orders/" + orderID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("order service returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var protoResp httppb.GetOrderResponse
	if err := proto.Unmarshal(body, &protoResp); err != nil {
		return nil, err
	}

	return &protoResp, nil
}
