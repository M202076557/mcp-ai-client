.PHONY: build run clean test help demo

# 默认目标
.DEFAULT_GOAL := help

# 项目名称
PROJECT_NAME := mcp-ai-client
BINARY_NAME := mcp-ai-client

# Go相关变量
GO := go
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# 构建目录
BUILD_DIR := bin
BUILD_PATH := $(BUILD_DIR)/$(BINARY_NAME)

# 版本信息
VERSION := 1.0.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# 帮助信息
help: ## 显示帮助信息
	@echo "$(PROJECT_NAME) - 统一的MCP客户端"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 构建项目
build: ## 构建项目
	@echo "构建 $(PROJECT_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(LDFLAGS) -o $(BUILD_PATH) ./cmd/server
	@echo "构建完成: $(BUILD_PATH)"

# 运行项目 (HTTP服务器模式)
run: build ## 构建并运行HTTP服务器
	@echo "启动 $(PROJECT_NAME) HTTP服务器..."
	@$(BUILD_PATH)

# 运行AI演示 (所有AI工具)
demo: build ## 运行AI工具演示
	@echo "运行AI工具演示..."
	@$(BUILD_PATH) demo

# AI工具单独演示
chat: build ## AI对话演示
	@$(BUILD_PATH) chat

file: build ## AI文件管理演示
	@$(BUILD_PATH) file

data: build ## AI数据处理演示
	@$(BUILD_PATH) data

api: build ## AI网络请求演示
	@$(BUILD_PATH) api

db: build ## AI数据库查询演示
	@$(BUILD_PATH) db

# 开发模式运行
dev: ## 开发模式运行（不构建）
	@echo "开发模式启动 $(PROJECT_NAME)..."
	@$(GO) run ./cmd/server

# 清理构建文件
clean: ## 清理构建文件
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@echo "清理完成"

# 安装依赖
deps: ## 安装Go依赖
	@echo "安装Go依赖..."
	@$(GO) mod download
	@$(GO) mod tidy
	@echo "依赖安装完成"

# 测试
test: ## 运行测试
	@echo "运行测试..."
	@$(GO) test -v ./...

# 代码检查
lint: ## 运行代码检查
	@echo "运行代码检查..."
	@$(GO) vet ./...
	@echo "代码检查完成"

# 格式化代码
fmt: ## 格式化代码
	@echo "格式化代码..."
	@$(GO) fmt ./...
	@echo "代码格式化完成"

# 显示项目信息
info: ## 显示项目信息
	@echo "项目信息:"
	@echo "  名称: $(PROJECT_NAME)"
	@echo "  版本: $(VERSION)"
	@echo "  构建时间: $(BUILD_TIME)"
	@echo "  Git提交: $(GIT_COMMIT)"
	@echo "  操作系统: $(GOOS)"
	@echo "  架构: $(GOARCH)"
	@echo "  二进制文件: $(BUILD_PATH)"

# 创建发布版本
release: clean build ## 创建发布版本
	@echo "创建发布版本..."
	@tar -czf $(PROJECT_NAME)-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz $(BUILD_DIR)
	@echo "发布版本创建完成: $(PROJECT_NAME)-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz"

# 安装到系统
install: build ## 安装到系统
	@echo "安装 $(BINARY_NAME) 到系统..."
	@sudo cp $(BUILD_PATH) /usr/local/bin/
	@echo "安装完成"

# 卸载
uninstall: ## 从系统卸载
	@echo "卸载 $(BINARY_NAME)..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "卸载完成"

