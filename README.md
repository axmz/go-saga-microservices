# Saga Microservices E-commerce Store

This is a demo e-commerce store built with Go and microservices.
Microservices communicate via REST and Kafka exchanging protobuf messages.
Saga pattern is used for distributed transactions.
Debezium is used as automated Outbox pattern.
The project is deployed to GCP (ephemeral IP)

## TODO:

- auto release reserved after timeout
- next.js
- apply configs from adapters
- use interfaces
- separate http server from grpc
- vscode debug describe better