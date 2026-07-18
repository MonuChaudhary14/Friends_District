FROM golang:alpine AS builder

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/server .

# Expose port
EXPOSE 8080

# Command to run the application
CMD ["./server"]
