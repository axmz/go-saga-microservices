FROM golang:1.24.4-alpine as builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o payment-service ./main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/payment-service .
EXPOSE 8080
CMD ["./payment-service"] 