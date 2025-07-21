module github.com/axmz/go-saga-microservices/services/storefront

go 1.24.4

require (
	github.com/axmz/go-saga-microservices/config v0.0.0-00010101000000-000000000000
	github.com/axmz/go-saga-microservices/lib/adapter/http v0.0.0-00010101000000-000000000000
	github.com/axmz/go-saga-microservices/lib/logger v0.0.0-00010101000000-000000000000
	github.com/axmz/go-saga-microservices/pkg/graceful v0.0.0-00010101000000-000000000000
)

replace github.com/axmz/go-saga-microservices/lib/adapter/http => ../../lib/adapter/http

replace github.com/axmz/go-saga-microservices/lib/logger => ../../lib/logger

replace github.com/axmz/go-saga-microservices/pkg/graceful => ../../pkg/graceful

replace github.com/axmz/go-saga-microservices/config => ../../config
