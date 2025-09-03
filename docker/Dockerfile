#
FROM node:22.19.0-alpine AS frontend
WORKDIR /app/frontend

COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci

COPY frontend/ ./

RUN npm run build

# 
FROM golang:latest AS backend
WORKDIR /app

COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://goproxy.io,direct
RUN go mod tidy

COPY . .

COPY --from=frontend /app/frontend/dist ./static

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main .

#
FROM alpine:latest
WORKDIR /app

COPY --from=backend /app/main .

COPY --from=backend /app/static ./static
COPY --from=backend /app/config.json .

EXPOSE 8080

CMD ["./main"]
