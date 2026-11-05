# ---- Stage 1: Build ----
FROM golang:1.22-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/alat-gudang ./cmd/server

# ---- Stage 2: Runtime ----
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /out/alat-gudang /app/alat-gudang

ENV TZ=Asia/Jakarta

EXPOSE 8080

ENTRYPOINT ["/app/alat-gudang"]
