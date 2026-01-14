FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/app/group/api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o group group.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/app/group/api/group .
COPY --from=builder /app/app/group/api/etc ./etc

EXPOSE 10500

CMD ["./group", "-f", "etc/group-api.yaml"]
