FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/app/ws
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ws ws.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/app/ws/ws .
COPY --from=builder /app/app/ws/etc ./etc

EXPOSE 10300

CMD ["./ws", "-f", "etc/ws.yaml"]
