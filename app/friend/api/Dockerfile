FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

COPY . .

WORKDIR /app/app/friend/api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o friend friend.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/app/friend/api/friend .
COPY --from=builder /app/app/friend/api/etc ./etc

EXPOSE 10200

CMD ["./friend", "-f", "etc/friend-api-docker.yaml"]

