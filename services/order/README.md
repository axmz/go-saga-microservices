# Order Service

A microservice for managing orders in the saga pattern implementation.

## Features

- Create new orders
- Retrieve order details
- Update order status
- Kafka event publishing for saga orchestration
- PostgreSQL data persistence

## API Endpoints

### Health Check
- `GET /healthz` - Service health check

### Orders
- `POST /orders` - Create a new order
- `GET /orders` - Get all orders
- `GET /orders/{id}` - Get order by ID
- `PUT /orders/{id}/status` - Update order status

## Environment Variables

- `DB_HOST` - PostgreSQL host (default: localhost)
- `DB_PORT` - PostgreSQL port (default: 5432)
- `DB_USER` - PostgreSQL user (default: order)
- `DB_PASSWORD` - PostgreSQL password (default: orderpass)
- `DB_NAME` - PostgreSQL database name (default: order)
- `KAFKA_BROKER` - Kafka broker address (default: localhost:9092)

## Database Schema

### Orders Table
- `id` - Order unique identifier (UUID)
- `customer_id` - Customer identifier
- `total_amount` - Order total amount
- `status` - Order status (pending, confirmed, cancelled, etc.)
- `created_at` - Order creation timestamp
- `updated_at` - Order last update timestamp

### Order Items Table
- `id` - Item unique identifier
- `order_id` - Reference to order
- `product_id` - Product identifier
- `product_name` - Product name
- `quantity` - Item quantity
- `unit_price` - Unit price
- `total_price` - Total price for this item
- `created_at` - Item creation timestamp

## Kafka Events

The service publishes the following events to the `order-events` topic:

### Order Created Event
```json
{
  "event_type": "order_created",
  "order_id": "uuid",
  "customer_id": "customer123",
  "status": "pending",
  "total_amount": 99.99,
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Order Status Updated Event
```json
{
  "event_type": "order_status_updated",
  "order_id": "uuid",
  "customer_id": "customer123",
  "status": "confirmed",
  "total_amount": 99.99,
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Running the Service

### Local Development
```bash
go mod tidy
go run cmd/main.go
```

### Docker
```bash
docker build -t order-service .
docker run -p 8080:8080 order-service
```

## Example Usage

### Create Order
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "customer123",
    "items": [
      {
        "product_id": "WIDGET-A",
        "product_name": "Widget A",
        "quantity": 2,
        "unit_price": 19.99
      }
    ]
  }'
```

### Get Orders
```bash
curl http://localhost:8080/orders
```

### Update Order Status
```bash
curl -X PUT http://localhost:8080/orders/{order_id}/status \
  -H "Content-Type: application/json" \
  -d '{"status": "confirmed"}'
``` 