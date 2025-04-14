# Step 1: Build the Go application
FROM golang:latest AS builder

# Set the working directory
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
COPY /ui ./ui
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application (optimized with -ldflags)
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/web/

# Step 2: Create a minimal runtime image
FROM alpine:latest

# Set a working directory
WORKDIR /root/

# Copy the UI folder (HTML, CSS, JS) into the container
COPY ui /root/ui

# Copy the built binary from the builder stage
COPY --from=builder /app/app .

# Expose the application port (change as needed)
EXPOSE 8080

# Run the Go application
CMD ["./app"]
