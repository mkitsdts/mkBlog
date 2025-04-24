# 开发阶段
FROM golang:1.24@sha256:d9db32125db0c3a680cfb7a1afcaefb89c898a075ec148fdc2f0f646cc2ed509 AS dev

# 安装必要工具但清理缓存
RUN apk add --no-cache curl && \
    curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin && \
    apk del curl

# 设置工作目录
WORKDIR /app

# 创建非root用户
RUN adduser -D -u 10001 appuser
USER appuser

# 复制依赖文件
COPY --chown=appuser:appuser go.mod go.sum ./
RUN go mod tidy

# 启动命令
CMD ["air"]

# 生产构建阶段
FROM golang:1.24@sha256:d9db32125db0c3a680cfb7a1afcaefb89c898a075ec148fdc2f0f646cc2ed509 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/main ./cmd/main.go

# 使用特定版本的精简基础镜像
FROM alpine:3.21.3@sha256:a8560b36e8b8210634f77d9f7f9efd7ffa463e380b75e2e74aff4511df3ef88c AS prod

# 安装基本的安全更新
RUN apk --no-cache add ca-certificates && \
    apk --no-cache upgrade

# 创建非root用户
RUN adduser -D -u 10001 appuser

# 复制编译的二进制文件并设置权限
COPY --from=builder --chown=appuser:appuser /app/main /app/main
COPY --chown=appuser:appuser ./resource /app/resource

# 使用非root用户
USER appuser
WORKDIR /app

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s CMD wget -q -O - http://localhost:8080/health || exit 1

EXPOSE 8080
CMD ["/app/main"]