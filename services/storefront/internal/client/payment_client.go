package client

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	httppb "github.com/axmz/go-saga-microservices/pkg/proto/http"
	"google.golang.org/protobuf/proto"
)

type PaymentClient interface {
	PaymentSuccess(ctx context.Context, req *httppb.PaymentSuccessRequest) error
	PaymentFail(ctx context.Context, req *httppb.PaymentFailRequest) error
}

type HTTPPaymentClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPPaymentClient(baseURL string) *HTTPPaymentClient {
	return &HTTPPaymentClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *HTTPPaymentClient) PaymentSuccess(ctx context.Context, req *httppb.PaymentSuccessRequest) error {
	protoData, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := c.client.Post(c.baseURL+"/payment-success", "application/x-protobuf", bytes.NewBuffer(protoData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("payment service returned status: %d", resp.StatusCode)
	}

	return nil
}

func (c *HTTPPaymentClient) PaymentFail(ctx context.Context, req *httppb.PaymentFailRequest) error {
	protoData, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := c.client.Post(c.baseURL+"/payment-fail", "application/x-protobuf", bytes.NewBuffer(protoData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("payment service returned status: %d", resp.StatusCode)
	}

	return nil
}
