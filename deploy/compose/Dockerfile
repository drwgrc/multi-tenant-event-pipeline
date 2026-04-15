FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY migrations ./migrations

CMD ["go", "run", "./cmd/api"]
