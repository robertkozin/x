FROM golang:1.22-alpine

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/gtd/*.go ./cmd/gtd/

CMD go run ./cmd/gtd/main.go
