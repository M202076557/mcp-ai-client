# MCP AI Client API 测试指南 - 简化版

## 项目概述

本项目已简化为核心功能：**5个AI增强工具 + 1个基础数据库查询**

## 基础数据库查询

```bash
# 获取用户列表（唯一的基础查询功能）
curl -X GET "http://localhost:8080/api/v1/db/users" | jq .
```

## AI 增强工具测试

### 5.1 基础 AI 对话

```bash
# 基础AI聊天 - 使用默认提供商
curl -X POST "http://localhost:8080/api/v1/ai/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "你好，请介绍一下MCP协议是什么？50字以内。",
    "max_tokens": 50
  }' | jq .

# 指定提供商和模型
curl -X POST "http://localhost:8080/api/v1/ai/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "解释一下Go语言的并发特性",
    "provider": "ollama",
    "model": "llama2:7b"
  }' | jq .
```

### 5.2 AI 智能文件管理

```bash
# AI文件管理 - 创建Go项目结构
curl -X POST "http://localhost:8080/api/v1/ai/file-manager" \
  -H "Content-Type: application/json" \
  -d '{
    "instruction": "创建一个Go项目的标准目录结构",
    "target_path": "./demo-go-project",
    "operation_mode": "execute"
  }' | jq .

# AI文件管理 - 修改现有项目文件
curl -X POST "http://localhost:8080/api/v1/ai/file-manager" \
  -H "Content-Type: application/json" \
  -d '{
    "instruction": "在demo-go-project目录中添加一个HTTP服务器和配置文件",
    "target_path": "./demo-go-project",
    "operation_mode": "execute"
  }' | jq .
```

### 5.3 AI 智能数据处理

```bash
# AI数据处理 - JSON解析和分析
curl -X POST "http://localhost:8080/api/v1/ai/data-processor" \
  -H "Content-Type: application/json" \
  -d '{
    "instruction": "解析这个JSON数据并提取所有用户的邮箱地址",
    "input_data": "{\"users\":[{\"name\":\"张三\",\"email\":\"zhangsan@example.com\",\"age\":25},{\"name\":\"李四\",\"email\":\"lisi@example.com\",\"age\":30}]}",
    "data_type": "json",
    "output_format": "table",
    "operation_mode": "execute"
  }' | jq .

# AI数据处理 - CSV数据转换
curl -X POST "http://localhost:8080/api/v1/ai/data-processor" \
  -H "Content-Type: application/json" \
  -d '{
    "instruction": "将CSV格式数据转换为JSON格式",
    "input_data": "name,age,city\n张三,25,北京\n李四,30,上海",
    "data_type": "csv",
    "output_format": "json",
    "operation_mode": "execute"
  }' | jq .
```

### 5.4 AI 智能网络请求

```bash
# AI网络请求 - 获取示例用户数据
curl -X POST "http://localhost:8080/api/v1/ai/api-client" \
  -H "Content-Type: application/json" \
  -d '{
    "instruction": "获取用户数据",
    "base_url": "https://jsonplaceholder.typicode.com",
    "request_mode": "execute",
    "response_analysis": true
  }' | jq .

# AI网络请求 - 获取测试数据
curl -X POST "http://localhost:8080/api/v1/ai/api-client" \
  -H "Content-Type: application/json" \
  -d '{
    "instruction": "获取测试数据",
    "base_url": "https://httpbin.org",
    "request_mode": "execute",
    "response_analysis": true
  }' | jq .
```

### 5.5 AI 智能数据库查询

```bash
# AI自然语言数据查询
curl -X POST "http://localhost:8080/api/v1/ai/query-with-analysis" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "查询所有员工信息",
    "analysis_type": "insights",
    "table_name": "mcp_user"
  }' | jq .

# AI数据摘要报告
curl -X POST "http://localhost:8080/api/v1/ai/query-with-analysis" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "生成公司员工整体情况报告",
    "analysis_type": "summary"
  }' | jq .
```

## 项目架构说明

### 简化后的功能结构

本项目已简化为核心功能，专注于AI增强工具和基础数据库查询：

**AI增强工具（5个）**：
- 5.1 `ai_chat`: AI 对话和问答
- 5.2 `ai_file_manager`: 智能文件管理
- 5.3 `ai_data_processor`: 智能数据处理
- 5.4 `ai_api_client`: 智能网络请求
- 5.5 `ai_query_with_analysis`: 智能数据库查询

**基础数据库查询（1个）**：
- 用户列表查询: `GET /api/v1/db/users`

### MCP 协议集成

- **MCP Client**: 内置 WebSocket 客户端，连接到 mcp-ai-server
- **协议支持**: JSON-RPC 2.0 over WebSocket
- **AI 工具调用**: 通过 `/api/v1/ai/*` 路由调用 mcp-ai-server 的AI工具
- **错误处理**: 优雅的降级机制和超时处理

### 快速测试

```bash
# 健康检查
curl http://localhost:8080/health | jq .

# 服务概览
curl http://localhost:8080/ | jq .

# 基础数据库查询
curl http://localhost:8080/api/v1/db/users | jq .

# AI对话测试
curl -X POST http://localhost:8080/api/v1/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"prompt": "你好"}' | jq .
```
