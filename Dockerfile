FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

RUN go build -o click-counter .

EXPOSE 8080

CMD ["./click-counter"]