# Используем образ Golang для сборки
FROM golang:1.23-alpine AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum файлы (если есть)
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем все файлы проекта в контейнер
COPY . .

# Собираем приложение
RUN go build -o main ./cmd/main/main.go

# Минимальный образ для запуска
FROM alpine:latest

# Устанавливаем сертификаты для HTTPS-запросов
RUN apk --no-cache add ca-certificates

# Устанавливаем рабочую директорию
WORKDIR /new/

# Копируем скомпилированный бинарник
COPY --from=builder /app/main .

# Открываем порт для приложения
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]
