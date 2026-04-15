FROM golang:1.24

WORKDIR /app

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

CMD ["go", "run", "./cmd/api"]
