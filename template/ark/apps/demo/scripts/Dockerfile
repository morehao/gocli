FROM golang:1.23-alpine AS builder

# 设置环境变量
ENV GOPROXY=https://goproxy.cn,direct

# 设置工作目录
WORKDIR /build

# 复制整个项目代码
COPY . .
# 下载依赖
RUN go mod download



# 构建应用程序（入口文件在 apps/demo/cmd）
RUN go build -ldflags="-s -w" -o /app/demo ./apps/demo/cmd

# 运行阶段
FROM alpine:3.22

# 从构建阶段复制二进制文件
COPY --from=builder /app/demo /app/demo

# 从构建阶段复制生产环境配置文件
COPY --from=builder /build/apps/demo/config/config.prod.yaml /app/config.yaml

# 设置工作目录
WORKDIR /app

# 暴露端口
EXPOSE 8099

# 启动应用程序
ENTRYPOINT ["/app/demo"]
