# ============================================================
# 第一阶段：构建 (Build Stage)
# ============================================================
FROM golang:1.23.4-alpine AS builder

WORKDIR /app

# 复制项目文件
COPY . .

# 配置 Go 代理（加速下载）
ENV GOPROXY=https://goproxy.cn,direct

# 下载依赖并编译
RUN go mod tidy && \
    go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# ============================================================
# 第二阶段：运行 (Run Stage)
# ============================================================
FROM alpine:latest

# 安装必要工具：CA 证书、时区、cron
RUN apk --no-cache add ca-certificates tzdata cronie

# 设置时区为中国时区
ENV TZ=Asia/Shanghai

# 设置工作目录
WORKDIR /root/

# 从第一阶段复制编译好的程序
COPY --from=builder /app/main .

# 复制配置文件
COPY --from=builder /app/config.json .

# 复制清理脚本
COPY scripts/clean_notes.sh /usr/local/bin/clean_notes.sh
RUN chmod +x /usr/local/bin/clean_notes.sh

# 复制启动脚本
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# 配置 cron 定时任务（每 6 小时执行一次清理）
RUN echo "0 */6 * * * /usr/local/bin/clean_notes.sh" >> /etc/crontabs/root

# 创建 data 目录（存放二维码、会话文件、笔记）
RUN mkdir -p /root/data

# 暴露 HTTP 服务端口（二维码页面）
EXPOSE 8080

# 挂载数据卷
VOLUME ["/root/data"]

# 设置工作目录
WORKDIR /root

# 使用启动脚本（同时运行 cron 和主程序）
CMD ["/usr/local/bin/docker-entrypoint.sh"]