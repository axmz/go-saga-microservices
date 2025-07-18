module github.com/axmz/go-saga-microservices/services/order

go 1.24.4

require (
	github.com/axmz/go-saga-microservices/lib/adapter/db v0.0.0-00010101000000-000000000000
	github.com/axmz/go-saga-microservices/lib/adapter/http v0.0.0-00010101000000-000000000000
	github.com/axmz/go-saga-microservices/lib/adapter/kafka v0.0.0-00010101000000-000000000000
	github.com/axmz/go-saga-microservices/lib/logger v0.0.0-00010101000000-000000000000
	github.com/axmz/go-saga-microservices/pkg/graceful v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
	github.com/segmentio/kafka-go v0.4.48
	golang.org/x/sync v0.15.0
)

require (
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	golang.org/x/net v0.41.0 // indirect
)

replace github.com/axmz/go-saga-microservices/lib/adapter/db => ../../lib/adapter/db

replace github.com/axmz/go-saga-microservices/lib/adapter/kafka => ../../lib/adapter/kafka

replace github.com/axmz/go-saga-microservices/lib/adapter/http => ../../lib/adapter/http

replace github.com/axmz/go-saga-microservices/lib/logger => ../../lib/logger

replace github.com/axmz/go-saga-microservices/pkg/graceful => ../../pkg/graceful
