#!/bin/bash

# Migration script for inventory service
# Requires golang-migrate to be installed

echo "Running inventory service migrations..."

# Run migrations
migrate -path migrations -database "postgres://inventory:inventorypass@localhost:5433/inventory?sslmode=disable" up

echo "Migrations completed!" 