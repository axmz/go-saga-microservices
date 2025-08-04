package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/axmz/go-saga-microservices/services/order/internal/domain"
	"github.com/axmz/go-saga-microservices/services/order/internal/publisher"
	"github.com/axmz/go-saga-microservices/services/order/internal/repository"
	"github.com/axmz/go-saga-microservices/services/order/internal/sync"
)

type Service struct {
	Repo  *repository.Repository
	Kafka *publisher.Publisher
	Sync  *sync.Sync
}

func New(repo *repository.Repository, kafka *publisher.Publisher, s *sync.Sync) *Service {
	return &Service{
		Repo:  repo,
		Kafka: kafka,
		Sync:  s,
	}
}

func (s *Service) CreateOrder(ctx context.Context, items []domain.Item) (*domain.Order, error) {
	order := domain.NewOrder(items)
	ch := s.Sync.Push(order.ID)
	defer func() {
		s.Sync.Remove(order.ID)
	}()

	err := s.Repo.CreateOrder(ctx, order)
	if err != nil {
		return nil, err
	}
	if err := s.Kafka.PublishOrderCreated(ctx, order); err != nil {
		return nil, fmt.Errorf("order created but failed to emit event: %w", err)
	}
	slog.Info("[OrderService] Created order: %s, status: %s", order.ID, order.Status)
	// TODO: handle ctx cancellation
	status := <-ch
	if status != sync.OK {
		return order, fmt.Errorf("order created but failed to reserve items: %s", order.ID)
	}
	// By this point the order is in AwaitingPayment status, update by the consumer handlers.
	// We cut corners here.
	order.Status = domain.StatusAwaitingPayment
	slog.Info("[OrderService] Order items reserved successfully:", "orderID", order.ID)
	return order, nil
}

func (s *Service) GetOrder(ctx context.Context, orderId string) (*domain.Order, error) {
	return s.Repo.GetOrder(ctx, orderId)
}

func (s *Service) UpdateOrder(ctx context.Context, orderID string, status domain.Status) error {
	// TODO: do better
	o := domain.NewOrder(nil)
	o.ID = orderID
	o.Status = status

	err := s.Repo.UpdateOrder(ctx, o)
	slog.Info("Updating order status:", "orderID", orderID, "status", status)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) UpdateOrderAwaitingPayment(ctx context.Context, orderID string) {
	err := s.UpdateOrder(ctx, orderID, domain.StatusAwaitingPayment)
	if err != nil {
		// TODO: handle errors better, ex:
		// - cron to deal with PENDINGs for long time
		// - alerting
		// - manual intervention list
		// - fail them
	}

	ch, err := s.Sync.Pull(orderID)
	if err != nil {
		slog.Error("Failed to pull sync channel for order", "orderID", orderID, "err", err)
		return
	}

	// Wait for the channel to receive a status
	ch <- sync.OK

	// push ch to map with SUCCESS
	slog.Info("InventoryReservationSucceeded:", "orderID", orderID)
}

func (s *Service) UpdateOrderPaid(ctx context.Context, orderID string) {
	time.Sleep(time.Second * 5) // Simulate some processing delay
	if err := s.UpdateOrder(ctx, orderID, domain.StatusPaid); err != nil {
		// TODO: handle errors better
	}
	slog.Info("Order paid:", "orderID", orderID)
}

func (s *Service) UpdateOrderFailed(ctx context.Context, orderID string) {
	if err := s.UpdateOrder(ctx, orderID, domain.StatusFailed); err != nil {
		// TODO: handle errors better
	}
	slog.Info("Order failed:", "orderID", orderID)
}
