# Makefile for Program Manager - Linux专用构建

# 版本信息
VERSION := $(shell git describe --tags --always 2>/dev/null || echo "v1.0.0")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建参数
BINARY_NAME := program-manager
LDFLAGS := -w -s -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitHash=$(GIT_HASH)

# 默认目标
.PHONY: all linux clean test help

all: linux

# Linux AMD64构建
linux:
	@echo "=== 构建Linux AMD64版本 ==="
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
		-ldflags "$(LDFLAGS)" \
		-o dist/$(BINARY_NAME)-linux-amd64 .
	@echo "✅ Linux AMD64构建完成: dist/$(BINARY_NAME)-linux-amd64"

# 交叉编译其他架构 (可选)
linux-arm64:
	@echo "=== 构建Linux ARM64版本 ==="
	@mkdir -p dist
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build \
		-ldflags "$(LDFLAGS)" \
		-o dist/$(BINARY_NAME)-linux-arm64 .
	@echo "✅ Linux ARM64构建完成: dist/$(BINARY_NAME)-linux-arm64"

# 清理构建文件
clean:
	@echo "=== 清理构建文件 ==="
	@rm -rf dist/
	@echo "✅ 清理完成"

# 运行测试
test:
	@echo "=== 运行测试 ==="
	@go test ./...

# 静态检查
lint:
	@echo "=== 运行静态检查 ==="
	@golangci-lint run

# 帮助信息
help:
	@echo "可用的构建目标:"
	@echo "  linux       - 构建Linux AMD64版本"
	@echo "  linux-arm64 - 构建Linux ARM64版本"
	@echo "  clean       - 清理构建文件"
	@echo "  test        - 运行测试"
	@echo "  lint        - 运行静态检查"
	@echo ""
	@echo "示例:"
	@echo "  make linux  # 构建Linux AMD64版本"
	@echo "  make clean  # 清理构建文件"