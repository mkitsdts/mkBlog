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