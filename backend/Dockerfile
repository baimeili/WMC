# 多阶段构建：编译静态二进制 -> 最小运行时镜像
FROM golang:1.26-alpine AS builder
WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/wmc-api ./cmd/api \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/wmc-worker ./cmd/worker \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/wmc-seed ./cmd/seed

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 wmc
WORKDIR /app
COPY --from=builder /out/wmc-api /out/wmc-worker /out/wmc-seed /app/
USER wmc
EXPOSE 8080
ENTRYPOINT ["/app/wmc-api"]
