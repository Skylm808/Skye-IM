FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/app/friend/rpc
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o friend friend.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/app/friend/rpc/friend .
COPY --from=builder /app/app/friend/rpc/etc ./etc

EXPOSE 9200

CMD ["./friend", "-f", "etc/friend-docker.yaml"]

