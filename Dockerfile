FROM golang:1.23.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o tasks_bot ./cmd/*.go

# Use a minimal base image for the final stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/tasks_bot /app/tasks_bot
COPY --from=builder /app/.env /app/.env

# Run
CMD ["./tasks_bot"]
