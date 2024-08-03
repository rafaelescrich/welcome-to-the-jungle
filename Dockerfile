# Build stage
FROM golang:1.22.2 AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code excluding dataset.csv
COPY main.go .
COPY pkg/ ./pkg/
COPY go.mod .
COPY go.sum .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM gcr.io/distroless/base-debian10

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/main /app/main

# Command to run the binary
CMD ["./main"]
