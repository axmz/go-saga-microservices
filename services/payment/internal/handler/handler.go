package handler

import (
	"io"
	"log/slog"
	"net/http"

	httputils "github.com/axmz/go-saga-microservices/lib/adapter/http"
	"github.com/axmz/go-saga-microservices/payment-service/internal/service"
	httppb "github.com/axmz/go-saga-microservices/pkg/proto/http"
	"google.golang.org/protobuf/proto"
)

type Handler struct {
	Service *service.Service
}

func New(service *service.Service) *Handler {
	return &Handler{
		Service: service,
	}
}

func (h *Handler) PaymentSuccess(w http.ResponseWriter, r *http.Request) {
	slog.Info("PaymentSuccess request", "method", r.Method, "path", r.URL.Path, "remote", r.RemoteAddr)
	req, err := h.processPaymentSuccessRequest(r)
	if err != nil {
		slog.Warn("PaymentSuccess bad request", "err", err)
		httputils.ErrorBadRequest(w, err)
		return
	}

	if err := h.Service.PaymentSuccess(r.Context(), req.OrderId); err != nil {
		slog.Error("PaymentSuccess service error", "orderId", req.OrderId, "err", err)
		httputils.ErrorInternal(w, err)
		return
	}

	slog.Info("PaymentSuccess processed", "orderId", req.OrderId)
	h.respondWithPaymentSuccess(w)
}

func (h *Handler) PaymentFail(w http.ResponseWriter, r *http.Request) {
	slog.Info("PaymentFail request", "method", r.Method, "path", r.URL.Path, "remote", r.RemoteAddr)
	req, err := h.processPaymentFailRequest(r)
	if err != nil {
		slog.Warn("PaymentFail bad request", "err", err)
		httputils.ErrorBadRequest(w, err)
		return
	}

	if err := h.Service.PaymentFail(r.Context(), req.OrderId); err != nil {
		slog.Error("PaymentFail service error", "orderId", req.OrderId, "err", err)
		httputils.ErrorInternal(w, err)
		return
	}

	slog.Info("PaymentFail processed", "orderId", req.OrderId)
	h.respondWithPaymentFail(w)
}

// REQ PROCESSING
func (h *Handler) parseProtoBody(r *http.Request, msg proto.Message) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(body, msg); err != nil {
		return err
	}
	return nil
}

func (h *Handler) processPaymentSuccessRequest(r *http.Request) (*httppb.PaymentSuccessRequest, error) {
	req := new(httppb.PaymentSuccessRequest)
	if err := h.parseProtoBody(r, req); err != nil {
		return nil, err
	}
	return req, nil
}

func (h *Handler) processPaymentFailRequest(r *http.Request) (*httppb.PaymentFailRequest, error) {
	req := new(httppb.PaymentFailRequest)
	if err := h.parseProtoBody(r, req); err != nil {
		return nil, err
	}
	return req, nil
}

// RESPONSES
func (h *Handler) respondWithPaymentSuccess(w http.ResponseWriter) {
	resp := &httppb.PaymentSuccessResponse{Success: true}
	httputils.RespondProto(w, resp, http.StatusOK)
}

func (h *Handler) respondWithPaymentFail(w http.ResponseWriter) {
	resp := &httppb.PaymentFailResponse{Success: true}
	httputils.RespondProto(w, resp, http.StatusOK)
}
