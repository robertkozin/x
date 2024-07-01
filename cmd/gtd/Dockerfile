FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/gtd/*.go ./cmd/gtd/
COPY ./cmd/htmgo/*.go ./cmd/htmgo/
COPY ./htmgo/*.go ./htmgo/

RUN CGO_ENABLED=1 go install ./cmd/gtd

WORKDIR /data

ENTRYPOINT ["gtd"]
