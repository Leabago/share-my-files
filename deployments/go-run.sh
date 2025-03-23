#!/bin/bash

export APP_PORT=:8080
export REDIS_ADDR=localhost:6379
# MAX_FILE_SIZE size in bytes, 100 megabytes = = 104857600 bytes
export MAX_FILE_SIZE=3145728
export ALLOWED_HOST=localhost:8080
 
go run ./cmd/web/