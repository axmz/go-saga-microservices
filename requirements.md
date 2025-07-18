# go-saga-microservices

> Proof of concept for a Saga-based e-commerce app using Go, Kafka, PostgreSQL, and Docker Compose.

## 💡 Purpose

To demonstrate a distributed transaction pattern (Saga) across multiple microservices — Order, Inventory, Payment — communicating via Kafka and using PostgreSQL as backing storage.

---

## 🏗️ Architecture Overview

- **Frontend (Storefront)**: Simple HTTP UI (can be static or minimal Go/React app)
- **Order Service**: Handles order placement and status tracking
- **Inventory Service**: Manages stock and reservations
- **Payment Service**: Simulates payment success/failure
- **Kafka**: Backbone for event-driven communication
- **PostgreSQL**: Local DB per service
- **Docker Compose**: Used for local orchestration

---

## 🧱 Services & Responsibilities

### Storefront

- Displays products
- Sends requests to create orders
- Handles payment redirects
- Websocket updates for order status

Pages/Requests:

- /
    - getAllProducts() 
        - GET /products -> request to inventory-service/products
        - Expects []Product
    - placeOrder([]productId)
        - Ajax POST /order -> request to order-service/order
            - Sends []Item
            - Expects:
                - Success 200 -> redirect to /payment
                - OOS 200 -> display error message OR redirect to /order?orderId (status) failed

- /payment&orderId
    - getOrder()
        - GET /order?orderId
        - Expects Order{id, status = "AwaitingPayment"}
        - Status of the order should be AwaitingPayment
    - paymentSuccess()
        - POST /payment-success&orderId -> redirect to order status page order?orderId
    - paymentFail()
        - POST/payment-fail&orderId -> redirect to order status page order?orderId with error message Payment Failed

- /order&orderId
    - getOrder()
        - GET /order?orderId
        - Status:
            - PaymentFailed
            - BeingProcessed
    - update status with ws://order-service/order?orderId

### Order Service

- Manages order state (Pending, AwaitingPayment, Paid, Failed)
- Emits `orderCreatedEvent`
- Listens for:
  - `reserveProductsEvent`
  - `paymentSuccessEvent`
  - `paymentFailedEvent`

DB: Postgres
Requests:
 - POST /orders
        - Expects: []Item
        - Runs: createOrder([]Item) {
            creates UUID attaches it to Order
            sets status Pending
            saves it to DB
            emits event orderCreateEvent(Order)
        }

- GET /orders?orderId
    - Runs: getOrder(orderId) {
        queries db
        returns Order
    }

- WS /orders/ws?orderId

Listens to Events:
- reserveProductsEvent
Expects orderId and message(fail or success)
if fail
marks status in the db as order failed
returns to storefront order failed
if success
marks awaiting payment
redirects to /payment?orderId and Order
- paymentSuccessEvent mark order status paid
- paymentFailedEvent mark order status failed


### Inventory Service
- Exposes `GET /products`
- Listens for:
  - `orderCreatedEvent`
  - `paymentFailedEvent` (to release inventory)
- Emits:
  - `reserveProductsEvent` (Success or OOS)

DB: Postgres
Requests:
- GET /products
Queries DB select * from inventory 
Returns []Product

Listens to Events:
- orderCreatedEvent
Expects: Order
Runs: reserveProducts(Order) {
checks if Order.Items are availalbe
if not returns ErrorOOS (some items are OOS, order can't be processed)
on success 
reserved the item in the DB 
sends event: reserveProductsEvent(orderId) with orderId and Success
}
- paymentFailedEvent(orderId)
Runs: releaseProducts() {- change status in db}

### Payment Service

- Simulates payment outcome
- Emits:
  - `paymentSuccessEvent`
  - `paymentFailedEvent`

- POST /payment&orderId&fail (fail will be used for testing)
    - if fail
        - emit paymentFailedEvent(orderId)
    - else
        - emit paymentSuccessEvent(orderId)
    - redirect to /order&orderId


---

## 📡 API Endpoints

### Storefront
| Method | Endpoint                                   | Description              |
| ------ | ------------------------------------------ | ------------------------ |
| GET    | `/products`                                | List all products        |
| POST   | `/order`                                   | Place an order (AJAX)    |
| GET    | `/order?orderId=...`                       | Show order status        |
| POST   | `/payment-success`                         | Finalize payment         |
| POST   | `/payment-fail`                            | Simulate payment failure |
| WS     | `ws://order-service/orders/ws?orderId=...` | Realtime status          |

---

## 📬 Events & Topics

| Event Name             | Kafka Topic          | Emitted By        | Consumed By       | Payload                       |
| ---------------------- | -------------------- | ----------------- | ----------------- | ----------------------------- |
| `orderCreatedEvent`    | `orders.created`     | Order Service     | Inventory Service | `{ orderId, items }`          |
| `reserveProductsEvent` | `inventory.reserved` | Inventory Service | Order Service     | `{ orderId, status: "success" | "fail" }` |
| `paymentSuccessEvent`  | `payments.success`   | Payment Service   | Order Service     | `{ orderId }`                 |
| `paymentFailedEvent`   | `payments.failed`    | Payment Service   | Order & Inventory | `{ orderId }`                 |

---

## 🔄 Order State Machine

```text
[Pending] -> Inventory confirms -> [AwaitingPayment]
[Pending] -> Inventory OOS      -> [Failed]
[AwaitingPayment] -> Success    -> [Paid]
[AwaitingPayment] -> Fail       -> [Failed]

