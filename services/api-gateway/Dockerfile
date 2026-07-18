# ZhiPath API Gateway (Go Hertz)
# 多阶段构建；Task 5 提供 go.mod 与 cmd/server 后可正常 build
FROM golang:1.25.3 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

FROM debian:12-slim
RUN groupadd -g 65532 nonroot && useradd -u 65532 -g nonroot nonroot
WORKDIR /app
COPY --from=builder /out/server /server
COPY migrations ./migrations
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
