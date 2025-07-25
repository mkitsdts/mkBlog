FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go env -w  GOPROXY=https://goproxy.io,direct && go mod tidy

# 复制所有源代码
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/main ./main.go

FROM alpine:latest AS prod

RUN addgroup -S appgroup && \
    adduser -S appuser -G appgroup

# builder 阶段复制编译好的二进制文件
COPY --from=builder --chown=appuser:appgroup /app/main /app/main

# 复制 resource 目录 (如果存在且需要)
# 假设 resource 目录与 Dockerfile 的同级目录下
COPY --chown=appuser:appgroup ./resource /app/resource
COPY --chown=appuser:appgroup ./config.json /app

WORKDIR /app

# 使用非 root 用户运行
USER appuser

EXPOSE 8080

# 运行应用程序
CMD ["/app/main"]