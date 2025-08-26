# MCP AI Client - 简化版

基于MCP协议的AI增强客户端服务，专注于提供5类核心AI工具和基础数据库查询功能。

## 🚀 功能特性

### AI增强工具 (5.1-5.5)

1. **5.1 基础AI对话** (`ai_chat`)
   - 支持自然语言对话
   - 多AI提供商支持 (Ollama, OpenAI, Anthropic)
   - 可配置模型和参数

2. **5.2 AI智能文件管理** (`ai_file_manager`)
   - 智能理解文件操作需求
   - 自动生成项目结构
   - 支持多种项目类型创建

3. **5.3 AI智能数据处理** (`ai_data_processor`)
   - 自动识别数据格式 (JSON, CSV等)
   - 智能数据解析和转换
   - 支持多种输出格式

4. **5.4 AI智能网络请求** (`ai_api_client`)
   - 理解API调用意图
   - 自动构造HTTP请求
   - 智能响应分析

5. **5.5 AI智能数据库查询** (`ai_query_with_analysis`)
   - 自然语言转SQL查询
   - 智能数据分析和洞察
   - 业务报告生成

### 基础数据库查询

- 用户列表查询
- 用户详情查询
- 用户搜索功能
- 用户统计信息

## 📋 API端点

### AI工具API (POST)

```bash
# 5.1 AI对话
POST /api/v1/ai/chat
{
  "prompt": "你好，请介绍一下MCP协议",
  "provider": "ollama",
  "model": "codellama:7b"
}

# 5.2 AI文件管理
POST /api/v1/ai/file-manager
{
  "instruction": "创建一个Go项目的标准目录结构",
  "target_path": "./demo-go-project",
  "operation_mode": "execute"
}

# 5.3 AI数据处理
POST /api/v1/ai/data-processor
{
  "instruction": "解析JSON数据并提取所有用户的邮箱地址",
  "input_data": "{\"users\":[{\"name\":\"张三\",\"email\":\"zhangsan@example.com\"}]}",
  "data_type": "json",
  "output_format": "table",
  "operation_mode": "execute"
}

# 5.4 AI网络请求
POST /api/v1/ai/api-client
{
  "instruction": "获取用户数据",
  "base_url": "https://jsonplaceholder.typicode.com",
  "request_mode": "execute",
  "response_analysis": true
}

# 5.5 AI数据库查询
POST /api/v1/ai/query-with-analysis
{
  "description": "查询所有员工信息",
  "analysis_type": "insights",
  "table_name": "mcp_user"
}
```

### 基础数据库API (GET)

```bash
# 用户列表
GET /api/v1/db/users

# 用户详情
GET /api/v1/db/users/:id

# 用户搜索
GET /api/v1/db/search/users?keyword=张三

# 用户统计
GET /api/v1/db/stats/users?table=mcp_user
```

### 系统API

```bash
# 健康检查
GET /health

# 服务概览
GET /
```

## 🛠️ 快速开始

### 前置要求

1. **Go 1.19+**
2. **MySQL数据库**
3. **MCP AI Server** (运行在 ws://localhost:8081)
4. **AI提供商** (Ollama/OpenAI/Anthropic)

### 安装和运行

```bash
# 1. 克隆项目
git clone <repository-url>
cd mcp-ai-client

# 2. 安装依赖
go mod download

# 3. 配置文件
cp configs/config.yaml.example configs/config.yaml
# 编辑配置文件，设置数据库和MCP服务器连接信息

# 4. 构建项目
make build

# 5. 启动服务
./bin/mcp-ai-client
```

### 配置说明

```yaml
# configs/config.yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  mysql:
    host: "localhost"
    port: 3306
    username: "root"
    password: "root"
    database: "mcp_test"
  tables:
    user_table: "mcp_user"

mcp:
  server_url: "ws://localhost:8081"
  timeout: 30s

ai:
  response_language: "zh-CN"
  default_provider: "ollama"
  default_model: "codellama:7b"
  include_language_instruction: true
```

## 🧪 测试示例

### 测试AI对话

```bash
curl -X POST http://localhost:8080/api/v1/ai/chat \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "请介绍一下MCP协议的主要特点",
    "provider": "ollama",
    "model": "codellama:7b"
  }'
```

### 测试AI文件管理

```bash
curl -X POST http://localhost:8080/api/v1/ai/file-manager \
  -H "Content-Type: application/json" \
  -d '{
    "instruction": "创建一个Go项目的标准目录结构",
    "target_path": "./demo-go-project",
    "operation_mode": "execute"
  }'
```

### 测试基础数据库查询

```bash
# 获取用户列表
curl http://localhost:8080/api/v1/db/users

# 搜索用户
curl "http://localhost:8080/api/v1/db/search/users?keyword=张三"
```

## 📊 架构设计

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Client   │───▶│  MCP AI Client  │───▶│  MCP AI Server  │
│                 │    │   (简化版)      │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │  MySQL Database │    │  AI Providers   │
                       │                 │    │ (Ollama/OpenAI) │
                       └─────────────────┘    └─────────────────┘
```

## 🔧 开发说明

### 项目结构

```
mcp-ai-client/
├── cmd/server/          # 服务器主程序
├── internal/
│   ├── api/            # API处理器
│   ├── database/       # 数据库客户端
│   ├── mcp/           # MCP客户端
│   └── service/       # 业务服务层
├── configs/           # 配置文件
└── test/docs/         # 测试文档
```

### 核心组件

- **API处理器**: 处理HTTP请求，调用MCP客户端
- **MCP客户端**: 与MCP AI Server通信
- **数据库客户端**: MySQL数据库操作
- **服务层**: 业务逻辑封装

## 📝 版本历史

### v2.0.0 - 简化版

- ✅ 专注于5类核心AI工具 (5.1-5.5)
- ✅ 保留基础数据库查询功能
- ❌ 移除传统API对比功能
- ❌ 移除演示和展示功能
- ❌ 移除复杂的服务对比分析
- ✅ 简化路由配置
- ✅ 优化代码结构

### v1.0.0 - 完整版

- 支持传统API和MCP API对比
- 包含完整的演示功能
- 提供服务能力展示

## 🤝 贡献

欢迎提交Issue和Pull Request来改进项目。

## 📄 许可证

MIT License
