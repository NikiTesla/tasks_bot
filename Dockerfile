FROM golang:1.23.1

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o tasks_bot  ./cmd/*.go

# Run
CMD ["./tasks_bot"]
