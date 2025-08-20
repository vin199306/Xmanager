# 使用多阶段构建
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的构建工具
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags "-s -w -X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -trimpath -o program-manager .

# 运行时阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -g 1000 appgroup \
    && adduser -D -s /bin/sh -u 1000 -G appgroup appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/program-manager /app/program-manager

# 复制 Web 界面文件
COPY --from=builder /app/web /app/web

# 创建必要的目录
RUN mkdir -p /app/data /app/logs \
    && chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# 设置环境变量
ENV PORT=8080
ENV DATA_DIR=/app/data
ENV LOG_DIR=/app/logs

# 启动命令
CMD ["/app/program-manager"]