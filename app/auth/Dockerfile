# 多阶段构建：先编译，再运�?
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod �?go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制整个项目
COPY . .

# 编译 Auth 服务
WORKDIR /app/app/auth
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auth auth.go

# 运行阶段：使用更小的镜像
FROM alpine:latest

# 安装 ca-certificates（HTTPS 需要）
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为上�?
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/app/auth/auth .
COPY --from=builder /app/app/auth/etc ./etc

# 暴露端口
EXPOSE 10001

# 运行服务
CMD ["./auth", "-f", "etc/auth-api-docker.yaml"]

