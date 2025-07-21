package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

type Product struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	SKU    string  `json:"sku"`
	Status string  `json:"status"`
	Price  float64 `json:"price"`
}

type OrderCreatedEvent struct {
	EventType   string      `json:"event_type"`
	OrderID     string      `json:"order_id"`
	CustomerID  string      `json:"customer_id"`
	Status      string      `json:"status"`
	TotalAmount float64     `json:"total_amount"`
	Items       []OrderItem `json:"items"`
	Timestamp   time.Time   `json:"timestamp"`
}

type OrderItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
}

type InventoryEvent struct {
	EventType string    `json:"event_type"`
	OrderID   string    `json:"order_id"`
	ProductID string    `json:"product_id"`
	SKU       string    `json:"sku"`
	Quantity  int       `json:"quantity"`
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

var db *sql.DB
var kafkaWriter *kafka.Writer
var kafkaReader *kafka.Reader

func main() {
	// Initialize database connection
	initDB()
	defer db.Close()

	// Initialize Kafka writer and reader
	initKafka()

	// Start Kafka consumer in a goroutine
	go consumeOrderEvents()

	// Initialize router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Inventory endpoints
	mux.HandleFunc("GET /products", getProducts)
	mux.HandleFunc("POST /products", createProduct)
	mux.HandleFunc("GET /products/{id}", getProduct) // Note: path param parsing must be handled manually
	mux.HandleFunc("PUT /products/{id}/status", updateStatus)

	// Use env vars for host/port
	host := getEnv("INVENTORY_SERVICE_HOST", "localhost")
	port := getEnv("INVENTORY_SERVICE_PORT", "8081")
	addr := fmt.Sprintf("%s:%s", host, port)

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("Inventory service starting on %s...", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func initDB() {
	host := getEnv("DB_HOST", "inventory-db")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "inventory")
	password := getEnv("DB_PASSWORD", "inventorypass")
	dbname := getEnv("DB_NAME", "inventory")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully connected to database")
}

func initKafka() {
	kafkaBroker := getEnv("KAFKA_BROKER", "kafka:9092")

	kafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP(kafkaBroker),
		Topic:    "inventory.reserved",
		Balancer: &kafka.LeastBytes{},
	}

	kafkaReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaBroker},
		Topic:   "orders.created, payments.failed",
		GroupID: "inventory-service",
	})

	log.Println("Successfully connected to Kafka")
}

func consumeOrderEvents() {
	for {
		ctx := context.Background()
		message, err := kafkaReader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}
		var event map[string]interface{}
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("Error unmarshaling event: %v", err)
			continue
		}
		eventType, _ := event["event_type"].(string)
		switch eventType {
		case "orderCreatedEvent":
			handleOrderCreated(event)
		case "paymentFailedEvent":
			handlePaymentFailed(event)
		}
	}
}

func handleOrderCreated(event map[string]interface{}) {
	orderID, _ := event["orderId"].(string)
	items, _ := event["items"].([]interface{})

	success := true
	for _, item := range items {
		itemMap, _ := item.(map[string]interface{})
		productID, _ := itemMap["product_id"].(string)
		if !reserveItem(orderID, productID, 1) {
			success = false
			break
		}
	}
	status := "success"
	if !success {
		status = "fail"
	}
	publishReserveProductsEvent(orderID, status)
	log.Printf("[InventoryService] Reservation result for order: %s, status: %s", orderID, status)
}

func handlePaymentFailed(event map[string]interface{}) {
	// Release reserved items
	// This logic needs to be updated to handle the new event payload
	// For now, we'll just log the event type
	log.Printf("Received payment failed event: %v", event)
}

func reserveItem(orderID, productID string, _ int) bool {
	// Check if product is available for reservation
	var status string
	err := db.QueryRow(
		"SELECT status FROM products WHERE sku = $1",
		productID,
	).Scan(&status)
	if err != nil {
		log.Printf("Error checking status for product %s: %v", productID, err)
		return false
	}

	// Check if product is available for reservation
	if status != "available" {
		log.Printf("Product %s is not available for reservation. Status: %s", productID, status)
		return false
	}

	// Reserve the item and update status
	_, err = db.Exec(
		"UPDATE products SET status = 'reserved', updated_at = CURRENT_TIMESTAMP WHERE sku = $1",
		productID,
	)
	if err != nil {
		log.Printf("Error reserving product %s: %v", productID, err)
		return false
	}

	log.Printf("Reserved product %s for order %s", productID, orderID)
	return true
}

func releaseItem(orderID, productID string, _ int) {
	_, err := db.Exec(
		"UPDATE products SET status = 'available', updated_at = CURRENT_TIMESTAMP WHERE sku = $1",
		productID,
	)
	if err != nil {
		log.Printf("Error releasing product %s: %v", productID, err)
	} else {
		log.Printf("Released product %s for order %s", productID, orderID)
	}
}

func getProducts(w http.ResponseWriter, r *http.Request) {
	products, err := loadProducts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func getProduct(w http.ResponseWriter, r *http.Request) {
	id := getPathParam(r.URL.Path, "/products/")
	if id == "" {
		http.Error(w, "Missing product id", http.StatusBadRequest)
		return
	}
	product, err := loadProduct(id)
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

func createProduct(w http.ResponseWriter, r *http.Request) {
	var product Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := saveProduct(product); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

func updateStatus(w http.ResponseWriter, r *http.Request) {
	id := getPathParam(r.URL.Path, "/products/")
	if id == "" {
		http.Error(w, "Missing product id", http.StatusBadRequest)
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := updateProductStatus(id, req.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	publishInventoryEvent("status_updated", "", id, 0, true, "Status updated")
	w.WriteHeader(http.StatusOK)
}

func loadProducts() ([]Product, error) {
	query := `SELECT id, name, sku, status, price 
			  FROM products 
			  ORDER BY name`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		err := rows.Scan(&product.ID, &product.Name, &product.SKU, &product.Status, &product.Price)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}

func loadProduct(productID string) (*Product, error) {
	query := `SELECT id, name, sku, status, price FROM products WHERE id = $1`
	var product Product
	err := db.QueryRow(query, productID).Scan(&product.ID, &product.Name, &product.SKU, &product.Status, &product.Price)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func saveProduct(product Product) error {
	query := `INSERT INTO products (name, sku, status, price) VALUES ($1, $2, $3, $4) RETURNING id`
	return db.QueryRow(query, product.Name, product.SKU, product.Status, product.Price).Scan(&product.ID)
}

func updateProductStatus(productID string, status string) error {
	query := `UPDATE products SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := db.Exec(query, status, productID)
	return err
}

func publishInventoryEvent(eventType, orderID, productID string, quantity int, success bool, message string) {
	event := InventoryEvent{
		EventType: eventType,
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  quantity,
		Success:   success,
		Message:   message,
		Timestamp: time.Now(),
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}

	err = kafkaWriter.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(orderID),
		Value: eventJSON,
	})
	if err != nil {
		log.Printf("Error publishing event: %v", err)
	}
}

func publishReserveProductsEvent(orderID, status string) {
	log.Printf("[InventoryService] Publishing reserveProductsEvent for order: %s, status: %s", orderID, status)
	event := map[string]interface{}{
		"orderId": orderID,
		"status":  status,
	}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling reserveProductsEvent: %v", err)
		return
	}
	err = kafkaWriter.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(orderID),
		Value: eventJSON,
	})
	if err != nil {
		log.Printf("Error publishing reserveProductsEvent: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getPathParam(path, prefix string) string {
	if len(path) <= len(prefix) {
		return ""
	}
	return path[len(prefix):]
}
