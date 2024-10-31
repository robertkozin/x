FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/gsites-comments/*.go ./cmd/gsites-comments/
COPY ./internal/. ./internal/

COPY ./cmd/gsites-comments/*.gohtml /data/cmd/gsites-comments/

RUN CGO_ENABLED=0 go install ./cmd/gsites-comments

WORKDIR /data

ENTRYPOINT ["gsites-comments"]
