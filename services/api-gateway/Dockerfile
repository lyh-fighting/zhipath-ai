# ZhiPath API Gateway (Go Hertz)
# 多阶段构建；Task 5 提供 go.mod 与 cmd/server 后可正常 build
FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/server /server
COPY migrations ./migrations
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
