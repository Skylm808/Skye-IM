FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

COPY . .

WORKDIR /app/app/user/api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o user user.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/app/user/api/user .
COPY --from=builder /app/app/user/api/etc ./etc

EXPOSE 10100

CMD ["./user", "-f", "etc/user-api-docker.yaml"]

