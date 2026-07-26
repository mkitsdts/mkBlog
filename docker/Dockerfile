FROM node:22-alpine AS frontend
WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm install

COPY frontend/ ./
RUN npm run build

FROM golang:1.24-alpine AS backend
WORKDIR /app

RUN sed -i 's#https\?://dl-cdn.alpinelinux.org/alpine#https://mirrors.tuna.tsinghua.edu.cn/alpine#g' /etc/apk/repositories
RUN apk update
RUN apk add gcc g++ git

COPY . .
RUN go env -w GOPROXY=https://goproxy.io,direct
RUN go mod tidy

RUN CGO_ENABLED=1 \
    GOOS=linux \
    GOGC=20 \
    GOMAXPROCS=2 \
    go build -ldflags="-s -w" -o /app/main .

FROM alpine:3.22
WORKDIR /app

COPY --from=backend /app/main .
COPY --from=frontend /app/frontend/dist/. ./static/

EXPOSE 4801 8080
EXPOSE 4802 6060

CMD ["./main", "-data_path", "/etc/mkblog"]
