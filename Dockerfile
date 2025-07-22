# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copiar los archivos del módulo y descargar dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar el código fuente
COPY . .

# Construir la aplicación
RUN CGO_ENABLED=0 GOOS=linux go build -o stock-tracker

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Instalar solo lo necesario para ejecutar
RUN apk --no-cache add ca-certificates

# Copiar el binario desde el builder
COPY --from=builder /app/stock-tracker .
COPY --from=builder /app/.env .

# Puerto expuesto
EXPOSE 8080

# Comando para ejecutar la aplicación
CMD ["./stock-tracker"]
