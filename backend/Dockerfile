# ==========================================
# Build Stage
# ==========================================
FROM golang:1.25-alpine AS builder

# Instala certificados CA e tzdata para o fuso horário
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Otimização de cache de dependências
COPY go.mod go.sum ./
RUN go mod download

# Copia o código-fonte
COPY . .

# Compilação estática do binário Go sem CGO e com strip de símbolos para reduzir tamanho
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/api ./cmd/api

# ==========================================
# Production Runtime Stage
# ==========================================
FROM alpine:3.21

# Instala certificados raiz (essenciais para TLS/SSL com bancos na nuvem) e dados de timezone
RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S appgroup -g 10001 \
    && adduser -S appuser -u 10001 -G appgroup

# Define timezone padrão
ENV TZ=America/Recife

WORKDIR /app

# Copia o binário compilado do estágio builder
COPY --from=builder /app/api /usr/local/bin/api

# Executa com usuário não-privilegiado (Hardening de Segurança)
USER appuser:appgroup

# Porta padrão exposta
EXPOSE 8080

# Health check local do container
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/api"]
