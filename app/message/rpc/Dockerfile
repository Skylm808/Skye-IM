FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/app/message/rpc
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o message message.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/app/message/rpc/message .
COPY --from=builder /app/app/message/rpc/etc ./etc

EXPOSE 9300

CMD ["./message", "-f", "etc/message-docker.yaml"]

