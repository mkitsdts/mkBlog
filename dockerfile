# ʹ���ض��汾�ľ����������
FROM alpine:3.21.3@sha256:a8560b36e8b8210634f77d9f7f9efd7ffa463e380b75e2e74aff4511df3ef88c AS prod

# ��װ�����İ�ȫ����
RUN apk --no-cache add ca-certificates && \
    apk --no-cache upgrade

# ������root�û�
RUN adduser -D -u 10001 appuser

# ���Ʊ���Ķ������ļ�������Ȩ��
COPY --from=builder --chown=appuser:appuser /app/main /app/main
COPY --chown=appuser:appuser ./resource /app/resource

# ʹ�÷�root�û�
USER appuser
WORKDIR /app

# �������
HEALTHCHECK --interval=30s --timeout=3s CMD wget -q -O - http://localhost:8080/health || exit 1

EXPOSE 8080
CMD ["/app/main"]