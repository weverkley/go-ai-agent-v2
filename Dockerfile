# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /go-ai-agent

# Final stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /go-ai-agent .
COPY .goaiagent ./.goaiagent

ENV GOAIAGENT_SESSIONSTORE_TYPE=redis
ENV GOAIAGENT_SESSIONSTORE_REDIS_ADDRESS=redis:6379

ENTRYPOINT ["/app/go-ai-agent"]
