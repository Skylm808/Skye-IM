FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

COPY . .

WORKDIR /app/app/upload/api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o upload upload.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/app/upload/api/upload .
COPY --from=builder /app/app/upload/api/etc ./etc

EXPOSE 10600

CMD ["./upload", "-f", "etc/upload-api-docker.yaml"]

