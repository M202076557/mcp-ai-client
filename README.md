# MCP AI Client

专门用于演示 MCP AI Server 中 AI 工具功能的客户端程序。通过 MCP 协议与服务器通信，展示 5 种 AI 增强工具的实际应用。

## 🎯 项目概述

这个客户端专门展示以下 AI 工具：

- **AI 对话** (ai_chat): 基础 AI 聊天功能
- **AI 文件管理** (ai_file_manager): 智能文件系统操作
- **AI 数据处理** (ai_data_processor): JSON/CSV 数据解析和转换
- **AI 网络请求** (ai_api_client): 智能 HTTP API 调用
- **AI 数据库查询** (ai_query_with_analysis): 数据库查询+AI 分析

## 🚀 快速开始

### 1. 启动 MCP AI Server

```bash
cd /path/to/mcp-ai-server
./bin/mcp-server -mode=websocket -port=8080
```

### 2. 构建并运行客户端

```bash
# 构建AI客户端
make ai-client

# 运行所有AI工具演示
make demo
```

## �️ 使用方法

### 完整演示

```bash
# 运行所有AI工具演示
./bin/ai-client demo
```

### 单独测试各工具

```bash
# AI对话演示
./bin/ai-client chat

# AI文件管理演示
./bin/ai-client file

# AI数据处理演示
./bin/ai-client data

# AI网络请求演示
./bin/ai-client api

# AI数据库查询演示
./bin/ai-client db
```

## 📁 项目结构

```
mcp-ai-client/
├── cmd/
│   └── ai-client/          # AI工具演示客户端
│       └── main.go
├── internal/
│   └── mcp/
│       └── ai_client.go    # AI专用MCP客户端
├── configs/
│   └── ai-config.yaml     # 客户端配置
├── test/
│   └── docs/
│       └── AI_TOOLS_TEST_GUIDE.md  # 完整测试指南
├── Makefile               # 构建和演示命令
└── README.md             # 本文档
```

## ⚙️ 配置

编辑 `configs/ai-config.yaml`：

```yaml
# MCP服务器配置
mcp:
  server_url: "ws://localhost:8080/ws"
  timeout: 30s

# AI工具配置
ai:
  response_language: "zh-CN"
  default_provider: "ollama"
  default_model: "llama2:7b"
```

## 🧪 测试

详细测试指南请参考：[test/docs/AI_TOOLS_TEST_GUIDE.md](test/docs/AI_TOOLS_TEST_GUIDE.md)

### 快速测试

```bash
# 检查构建是否成功
make ai-client

# 运行完整演示（需要mcp-ai-server运行）
make demo

# 运行自动化测试
cd test && bash auto_test.sh
```

## 🔧 构建命令

```bash
make help          # 显示帮助
make ai-client      # 构建AI客户端
make demo           # 运行演示
make clean          # 清理构建文件
make deps           # 安装依赖
```

## � 要求

- Go 1.21+
- MCP AI Server 运行中
- 网络连接（部分 AI 工具需要）

## 🚨 注意事项

1. 确保 `mcp-ai-server` 在运行并监听端口 8080
2. 某些 AI 工具需要配置相应的 AI 提供商
3. 网络相关的 AI 工具需要互联网连接
4. 数据库相关的 AI 工具需要预先准备数据

## 📖 相关文档

- [完整测试指南](test/docs/AI_TOOLS_TEST_GUIDE.md)
- [MCP AI Server](../mcp-ai-server/README.md)
  make test # 测试
  make clean # 清理

```

## 🔗 依赖项目

需要启动 `mcp-ai-server` 服务端（端口 8081）作为 MCP 协议服务提供者。

## 🔍 故障排除

1. **MCP 连接失败**: 检查 mcp-ai-server 服务是否运行在 ws://localhost:8081
2. **MySQL 连接失败**: 检查数据库服务和连接配置
3. **AI 工具错误**: 按顺序测试，从基础工具开始调试

## 📄 许可证

MIT License
```
