# 
FROM golang:tip-alpine AS backend
WORKDIR /app

RUN sed -i 's#https\?://dl-cdn.alpinelinux.org/alpine#https://mirrors.tuna.tsinghua.edu.cn/alpine#g' /etc/apk/repositories
RUN apk update
RUN apk add gcc g++ git

COPY . .
RUN go env -w GOPROXY=https://goproxy.io,direct
RUN go mod tidy

RUN CGO_ENABLED=1 GOOS=linux go build -o /app/main .

#
FROM alpine:3.22
WORKDIR /app

COPY --from=backend /app/main .

COPY --from=backend /app/static ./static
# 复制配置文件
COPY config.yaml ./config.yaml

EXPOSE 8080 6060

CMD ["./main"]
