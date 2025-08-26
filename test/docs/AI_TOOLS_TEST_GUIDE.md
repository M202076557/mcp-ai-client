# MCP AI Client API 测试指南

### 基础查询示例

```bash
# 查询 mcp_user 表
curl -X GET "http://localhost:8080/api/v1/traditional/users?table=mcp_user" | jq .
```

这## MCP AI API 测试

### 1 基础 AI 对话

```bash
# 基础AI聊天 - 使用默认提供商
curl -X POST "http://localhost:8080/api/v1/mcp/chat"
  -H "Content-Type: application/json"
  -d '{
    "prompt": "你好，请介绍一下MCP协议是什么？"
  }' | jq .

# 指定提供商和模型
curl -X POST "http://localhost:8080/api/v1/mcp/chat"
  -H "Content-Type: application/json"
  -d '{
    "prompt": "解释一下Go语言的并发特性",
    "provider": "ollama",
    "model": "llama2:7b"
  }' | jq .
```

### 2 AI 智能文件管理

```bash
# AI文件管理 - 创建Go项目结构
curl -X POST "http://localhost:8080/api/v1/mcp/file-manager"
  -H "Content-Type: application/json"
  -d '{
    "instruction": "创建一个Go项目的标准目录结构",
    "target_path": "./demo-go-project",
    "operation_mode": "execute"
  }' | jq .

# AI文件管理 - 修改现有项目文件
curl -X POST "http://localhost:8080/api/v1/mcp/file-manager"
  -H "Content-Type: application/json"
  -d '{
    "instruction": "在demo-go-project目录中添加一个HTTP服务器和配置文件",
    "target_path": "./demo-go-project",
    "operation_mode": "execute"
  }' | jq .
```

### 3 AI 智能数据处理

```bash
# AI数据处理 - JSON解析和分析
curl -X POST "http://localhost:8080/api/v1/mcp/data-processor"
  -H "Content-Type: application/json"
  -d '{
    "instruction": "解析这个JSON数据并提取所有用户的邮箱地址",
    "input_data": "{\"users\":[{\"name\":\"张三\",\"email\":\"zhangsan@example.com\",\"age\":25},{\"name\":\"李四\",\"email\":\"lisi@example.com\",\"age\":30}]}",
    "data_type": "json",
    "output_format": "table",
    "operation_mode": "execute"
  }' | jq .

# AI数据处理 - CSV数据转换
curl -X POST "http://localhost:8080/api/v1/mcp/data-processor"
  -H "Content-Type: application/json"
  -d '{
    "instruction": "将CSV格式数据转换为JSON格式",
    "input_data": "name,age,city\n张三,25,北京\n李四,30,上海",
    "data_type": "csv",
    "output_format": "json",
    "operation_mode": "execute"
  }' | jq .
```

### 4 AI 智能网络请求

```bash
# AI网络请求 - 获取示例用户数据
curl -X POST "http://localhost:8080/api/v1/mcp/api-client"
  -H "Content-Type: application/json"
  -d '{
    "instruction": "获取用户数据",
    "base_url": "https://jsonplaceholder.typicode.com",
    "request_mode": "execute",
    "response_analysis": true
  }' | jq .

# AI网络请求 - 获取测试数据
curl -X POST "http://localhost:8080/api/v1/mcp/api-client"
  -H "Content-Type: application/json"
  -d '{
    "instruction": "获取测试数据",
    "base_url": "https://httpbin.org",
    "request_mode": "execute",
    "response_analysis": true
  }' | jq .
```

### 5 AI 智能数据库查询

```bash
# AI自然语言数据查询
curl -X POST "http://localhost:8080/api/v1/mcp/query-with-analysis"
  -H "Content-Type: application/json"
  -d '{
    "description": "查询所有员工信息",
    "analysis_type": "insights",
    "table_name": "mcp_user"
  }' | jq .

# AI数据摘要报告
curl -X POST "http://localhost:8080/api/v1/mcp/query-with-analysis"
  -H "Content-Type: application/json"
  -d '{
    "description": "生成公司员工整体情况报告",
    "analysis_type": "summary"
  }' | jq .
```

## MCP 协议集成说明

本项目的核心价值在于集成 MCP (Model Context Protocol) 服务，通过标准化的协议与 AI 服务进行交互。

### MCP 集成架构

- **MCP Client**: 内置 WebSocket 客户端，连接到 mcp-ai-server
- **协议支持**: JSON-RPC 2.0 over WebSocket
- **AI 工具调用**: 支持调用 mcp-ai-server 提供的各种 AI 工具
- **错误处理**: 优雅的降级机制，MCP 不可用时回退到传统功能

### 支持的 AI 工具

通过 MCP 协议可以调用以下 AI 工具：

- `ai_chat`: AI 对话和问答
- `ai_data_processor`: 智能数据处理
- `ai_file_manager`: 文件管理
- `ai_api_client`: 网络请求处理
- `ai_query_with_analysis`: 数据库查询分析
