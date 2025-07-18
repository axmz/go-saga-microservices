FROM golang:1.24.4-alpine AS builder
ARG SERVICE
ARG MAIN=main.go
WORKDIR /app/services/${SERVICE}
COPY services/${SERVICE}/go.mod services/${SERVICE}/go.sum ./
# TODO: deal with this somehow
COPY ./lib /app/lib
COPY ./pkg /app/pkg
RUN go mod download
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /service ./cmd/${MAIN}

FROM alpine:latest
ARG SERVICE
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /service .
EXPOSE 8080
CMD ["./service"]