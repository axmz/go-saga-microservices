package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/axmz/go-saga-microservices/services/storefront/internal/renderer"
)

type Product struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	SKU    string  `json:"sku"`
	Status string  `json:"status"`
	Price  float64 `json:"price"`
}

type OrderItem struct {
	ProductID string `json:"product_id"`
}

type Order struct {
	ID          string      `json:"id"`
	CustomerID  string      `json:"customer_id"`
	Items       []OrderItem `json:"items"`
	TotalAmount float64     `json:"total_amount"`
	Status      string      `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type CreateOrderRequest struct {
	Items []OrderItem `json:"items"`
}

type PaymentInfo struct {
	CardNumber string `json:"card_number"`
	ExpiryDate string `json:"expiry_date"`
	CVV        string `json:"cvv"`
	Name       string `json:"name"`
}

// Service URLs
const (
	InventoryServiceURL = "http://inventory-service:8080"
	OrderServiceURL     = "http://order-service:8080"
)

func main() {
	fmt.Println("Storefront service starting on :8080")
	renderer := renderer.NewTemplateRenderer()

	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Storefront routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		homeHandler(w, r, renderer)
	})
	mux.HandleFunc("/payment", func(w http.ResponseWriter, r *http.Request) {
		paymentHandler(w, r, renderer)
	})
	mux.HandleFunc("/confirmation", func(w http.ResponseWriter, r *http.Request) {
		confirmationHandler(w, r, renderer)
	})
	mux.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		orderHandler(w, r, renderer)
	})
	mux.HandleFunc("/api/products", apiProductsHandler)
	mux.HandleFunc("POST /api/orders", apiCreateOrderHandler)

	log.Println("Storefront service starting on :8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request, renderer *renderer.TemplateRenderer) {
	products, err := getAllProducts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Products": products,
		"Title":    "Saga Microservices Storefront",
	}

	err = renderer.Render(w, "home.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func paymentHandler(w http.ResponseWriter, r *http.Request, renderer *renderer.TemplateRenderer) {
	if r.Method == "POST" {
		orderID := r.FormValue("orderId")
		fail := r.FormValue("fail")
		if orderID == "" {
			http.Error(w, "Missing orderId", http.StatusBadRequest)
			return
		}
		// Call payment-service
		paymentURL := fmt.Sprintf("http://payment-service:8080/payment?orderId=%s&fail=%s", orderID, fail)
		resp, err := http.Post(paymentURL, "application/json", nil)
		if err != nil {
			http.Error(w, "Payment service error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		// Always redirect to order page after payment attempt
		http.Redirect(w, r, "/order?orderId="+orderID, http.StatusSeeOther)
		return
	}
	// Show payment page
	err := renderer.Render(w, "payment.html", map[string]interface{}{
		"Title": "Payment",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func confirmationHandler(w http.ResponseWriter, r *http.Request, renderer *renderer.TemplateRenderer) {
	orderID := r.URL.Query().Get("order_id")

	data := map[string]interface{}{
		"Title":   "Order Confirmation",
		"OrderID": orderID,
	}

	err := renderer.Render(w, "confirmation.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func orderHandler(w http.ResponseWriter, r *http.Request, renderer *renderer.TemplateRenderer) {
	orderID := r.URL.Query().Get("orderId")
	if orderID == "" {
		http.Error(w, "Missing orderId", http.StatusBadRequest)
		return
	}
	order, err := getOrder(orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{
		"Order": order,
		"Title": "Order Status",
	}
	err = renderer.Render(w, "order.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getOrder(orderID string) (*Order, error) {
	resp, err := http.Get(OrderServiceURL + "/orders?orderId=" + orderID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("order service returned status: %d", resp.StatusCode)
	}
	var order Order
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return nil, err
	}
	return &order, nil
}

func apiProductsHandler(w http.ResponseWriter, r *http.Request) {
	products, err := getAllProducts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func apiCreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := createOrder(req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getAllProducts() ([]Product, error) {
	// Call inventory service to get available products
	resp, err := http.Get(InventoryServiceURL + "/products")
	if err != nil {
		return nil, fmt.Errorf("failed to call inventory service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("inventory service returned status: %d", resp.StatusCode)
	}

	var products []Product
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		return nil, fmt.Errorf("failed to decode products: %v", err)
	}

	return products, nil
}

func createOrder(orderReq CreateOrderRequest) error {
	orderData := CreateOrderRequest{
		Items: orderReq.Items,
	}

	jsonData, err := json.Marshal(orderData)
	if err != nil {
		return err
	}

	// Send to order service
	resp, err := http.Post(OrderServiceURL+"/orders", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("order service returned status: %d", resp.StatusCode)
	}

	return nil
}
