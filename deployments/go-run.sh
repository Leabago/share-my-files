#!/bin/bash

export APP_PORT=:8080
export REDIS_ADDR=localhost:6379
export DDNS_ADDRESS=http://localhost:8080

go run ./cmd/web/