#
FROM node:lts-alpine3.22 AS frontend
WORKDIR /app/frontend

COPY frontend/package.json frontend/package-lock.json* ./
RUN npm install

COPY frontend/ ./

RUN npm run build

# 
FROM golang:tip-alpine AS backend
WORKDIR /app

RUN sed -i 's#https\?://dl-cdn.alpinelinux.org/alpine#https://mirrors.tuna.tsinghua.edu.cn/alpine#g' /etc/apk/repositories
RUN apk update
RUN apk add gcc g++ git

COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://goproxy.io,direct
RUN go mod tidy

COPY . .

COPY --from=frontend /app/frontend/dist ./static

RUN CGO_ENABLED=1 GOOS=linux go build -o /app/main .

#
FROM alpine:3.22
WORKDIR /app

COPY --from=backend /app/main .

COPY --from=backend /app/static ./static
# 复制配置文件供前端运行时获取
COPY config.yaml ./config.yaml

EXPOSE 8080 6060

CMD ["./main"]
