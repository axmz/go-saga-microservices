FROM --platform=$BUILDPLATFORM golang:1.24.4-alpine AS builder
ARG SERVICE
ARG MAIN=main.go
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app/services/${SERVICE}
COPY services/${SERVICE}/go.mod services/${SERVICE}/go.sum ./
# TODO: deal with this somehow
COPY ./pkg /app/pkg
RUN go mod download
COPY . /app
RUN \
	echo "Building for ${TARGETOS:-linux}/${TARGETARCH:-amd64}" && \
	CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
	go build -a -installsuffix cgo -o /service ./cmd/${MAIN}

FROM alpine:latest
ARG SERVICE
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /service .
EXPOSE 8080
CMD ["./service"]