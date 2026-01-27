# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

WORKDIR /app/app/user/rpc
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o user user.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/app/user/rpc/user .
COPY --from=builder /app/app/user/rpc/etc ./etc

EXPOSE 9100

CMD ["./user", "-f", "etc/user-docker.yaml"]

