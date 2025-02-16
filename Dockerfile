FROM golang:1.23-alpine

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

# Устанавливаем bash
RUN apk add --no-cache bash

# Копируем скрипт ожидания (убедитесь, что wait-for-it.sh находится в корне проекта)
COPY wait-for-it.sh /usr/local/bin/wait-for-it.sh
RUN chmod +x /usr/local/bin/wait-for-it.sh

COPY . .
RUN go build -o main ./cmd/app

EXPOSE 8080
CMD ["sh", "-c", "wait-for-it.sh db:5432 -- ./main"]
