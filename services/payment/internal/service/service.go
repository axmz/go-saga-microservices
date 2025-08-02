package service

import (
	"context"

	"github.com/axmz/go-saga-microservices/payment-service/internal/publisher"
)

type Service struct {
	Kafka *publisher.Publisher
}

func New(kafka *publisher.Publisher) *Service {
	return &Service{
		Kafka: kafka,
	}
}

func (s *Service) PaymentSuccess(ctx context.Context, orderID string) error {
	return s.Kafka.PublishPaymentSucceededEvent(orderID)
}

func (s *Service) PaymentFail(ctx context.Context, orderID string) error {
	return s.Kafka.PublishPaymentFailedEvent(orderID)
}
