# Build stage
FROM golang:1.25-alpine AS build

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bunjgames-server .

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from build stage
COPY --from=build /app/bunjgames-server .

# Create media directory
RUN mkdir -p media

# Set environment variable for port (can be overridden)
ENV PORT=80

# Run the application
CMD ["./bunjgames-server"]
