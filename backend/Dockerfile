# =========================
# Build Stage
# =========================
FROM golang:1.26 AS builder

WORKDIR /app

# 先 copy mod 提高 cache 效率
COPY go.mod go.sum ./

RUN go mod download

# copy source
COPY . .

# codegen
RUN go generate ./...

# build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd

# =========================
# Runtime Stage
# =========================
FROM debian:12-slim

WORKDIR /app

COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]