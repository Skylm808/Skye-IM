FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

COPY . .

WORKDIR /app/app/message/api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o message message.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/app/message/api/message .
COPY --from=builder /app/app/message/api/etc ./etc

EXPOSE 10400

CMD ["./message", "-f", "etc/message-api-docker.yaml"]

