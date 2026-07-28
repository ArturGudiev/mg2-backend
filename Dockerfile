# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary (-buildvcs=false: no git in build context)
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /memoryguard .

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /memoryguard .

# Prefer runtime env from docker-compose / --env-file
EXPOSE 8080

ENTRYPOINT ["./memoryguard"]
