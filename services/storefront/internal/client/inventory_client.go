package client

import (
	"context"
	"fmt"
	"io"
	"net/http"

	httppb "github.com/axmz/go-saga-microservices/pkg/proto/http"
	"google.golang.org/protobuf/proto"
)

type InventoryClient interface {
	GetProducts(ctx context.Context) (*httppb.GetProductsResponse, error)
	ResetAll(ctx context.Context) error
}

type HTTPInventoryClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPInventoryClient(baseURL string) *HTTPInventoryClient {
	return &HTTPInventoryClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *HTTPInventoryClient) GetProducts(ctx context.Context) (*httppb.GetProductsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/products", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("inventory service returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var protoResp httppb.GetProductsResponse
	if err := proto.Unmarshal(body, &protoResp); err != nil {
		return nil, err
	}

	return &protoResp, nil
}

func (c *HTTPInventoryClient) ResetAll(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/products/reset", nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("inventory service returned status: %d", resp.StatusCode)
	}
	return nil
}
