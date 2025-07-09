FROM golang:1.24.4-alpine


WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download


COPY . .


RUN go build -o bot ./cmd


COPY .env .env


CMD ["./bot"]