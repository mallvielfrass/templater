# Используем официальный образ Go
FROM golang:1.23 AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum в рабочую директорию
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем остальной исходный код
COPY . .

# Собираем исполняемый файл
RUN go build -o main ./cmd/app 

# Второй этап: минимальный образ для запуска
FROM ubuntu:22.04
#-slim

# Устанавливаем переменные окружения
ENV GO_ENV=production \
    PORT=8080

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем исполняемый файл из предыдущего этапа
COPY --from=builder /app/main .
#install ca-certificates
RUN apt-get update && apt-get install -y ca-certificates git  
# Указываем порт, который используется приложением
EXPOSE 3053

# Указываем команду для запуска
CMD ["./main"]