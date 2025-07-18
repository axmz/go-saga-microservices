package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

type PaymentRequest struct {
	OrderID string `json:"orderId"`
	Fail    bool   `json:"fail"`
}

type PaymentEvent struct {
	OrderID string `json:"orderId"`
}

var (
	kafkaWriterSuccess *kafka.Writer
	kafkaWriterFail    *kafka.Writer
)

func main() {
	kafkaBroker := getEnv("KAFKA_BROKER", "kafka:9092")
	kafkaWriterSuccess = &kafka.Writer{
		Addr:  kafka.TCP(kafkaBroker),
		Topic: "payments.success",
	}
	kafkaWriterFail = &kafka.Writer{
		Addr:  kafka.TCP(kafkaBroker),
		Topic: "payments.failed",
	}

	http.HandleFunc("/payment", paymentHandler)
	log.Println("Payment service running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func paymentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	orderId := r.URL.Query().Get("orderId")
	fail := r.URL.Query().Get("fail") == "true"
	if orderId == "" {
		http.Error(w, "Missing orderId", http.StatusBadRequest)
		return
	}

	log.Printf("[PaymentService] Payment request for order: %s, fail: %v", orderId, fail)
	event := PaymentEvent{OrderID: orderId}
	var err error
	if fail {
		log.Printf("[PaymentService] Publishing paymentFailedEvent for order: %s", orderId)
		err = emitEvent(kafkaWriterFail, event)
		redirectURL := fmt.Sprintf("/order?orderId=%s&error=Payment+Failed", orderId)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	} else {
		log.Printf("[PaymentService] Publishing paymentSuccessEvent for order: %s", orderId)
		err = emitEvent(kafkaWriterSuccess, event)
		redirectURL := fmt.Sprintf("/order?orderId=%s", orderId)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	}
	if err != nil {
		log.Printf("Failed to emit payment event: %v", err)
	}
}

func emitEvent(writer *kafka.Writer, event PaymentEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	value, _ := json.Marshal(event)
	return writer.WriteMessages(ctx, kafka.Message{Value: value})
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
