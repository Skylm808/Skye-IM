FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

COPY . .

WORKDIR /app/app/gateway
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway gateway.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/app/gateway/gateway .
COPY --from=builder /app/app/gateway/etc ./etc

EXPOSE 8080

CMD ["./gateway", "-f", "etc/gateway-docker.yaml"]

