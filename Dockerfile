# 阶段一：编译阶段
FROM golang:1.25.5-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制依赖管理文件
COPY go.mod go.sum ./
RUN go mod download

# 复制项目源码
COPY . .

# 编译二进制文件，关闭CGO并静态链接，以适配Alpine环境
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-chat ./cmd/chat/main.go

# 阶段二：运行阶段
FROM alpine:latest

# 安装 ca-certificates 用于 HTTPS 请求（如调用OSS）
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/go-chat .

# 复制环境变量示例文件（后续通过挂载覆盖）
COPY .env.example .env

# 创建上传目录
RUN mkdir -p ./uploads

# 暴露服务端口
EXPOSE 8080

# 启动应用
CMD ["./go-chat"]