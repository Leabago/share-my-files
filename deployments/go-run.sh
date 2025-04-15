#!/bin/bash

export APP_PORT=:8080
export REDIS_ADDR=localhost:6379
export MAX_FILE_SIZE=3145728  #size in bytes, 100 megabytes = = 104857600 bytes
export ALLOWED_HOST=localhost:8080
export DNS_VALIDATION=false 

go run ./cmd/web/