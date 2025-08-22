# MCP AI Client

MCP (Model Context Protocol) AI 工具的 HTTP API 客户端，提供 RESTful 接口访问 7 个 AI 工具。

## 🚀 快速开始

### 前置要求

- Go 1.21+
- MySQL 5.7+
- MCP 服务器 (mcp-ai-server)

### 配置和启动

1. 配置数据库 `configs/config.yaml`：

```yaml
database:
  mysql:
    host: "localhost"
    port: 3306
    username: "root"
    password: "root"
    database: "mcp_test"
```

2. 构建并运行：

```bash
make build && make run
```

## 📡 API 接口

### 基础查询

```bash
# 健康检查
curl http://localhost:8080/health | jq .

# 直接MySQL查询
curl http://localhost:8080/api/v1/user | jq .

# MCP查询
curl http://localhost:8080/api/v1/mcp/user | jq .
```

### AI 工具 (7 个递增复杂度)

#### 1. AI 对话

```bash
curl -X POST "http://localhost:8080/api/v1/ai/chat" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Hello, what is this system?"}' | jq .
```

#### 2. SQL 生成

```bash
curl -X POST "http://localhost:8080/api/v1/ai/generate-sql" \
  -H "Content-Type: application/json" \
  -d '{"description": "查询IT部门员工", "table_name": "mcp_user"}' | jq .
```

#### 3. 智能查询（统一入口）

```bash
# 自然语言查询
curl -X POST "http://localhost:8080/api/v1/ai/smart-query" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "查询所有IT部门的员工"}' | jq .

# 直接SQL查询
curl -X POST "http://localhost:8080/api/v1/ai/smart-query" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "SELECT * FROM mcp_user"}' | jq .
```

#### 4. 数据分析

```bash
curl -X POST "http://localhost:8080/api/v1/ai/analyze-data" \
  -H "Content-Type: application/json" \
  -d '{"data": [{"name": "张三", "salary": 8000}], "analysis_type": "summary"}' | jq .
```

#### 5. 查询+分析

```bash
curl -X POST "http://localhost:8080/api/v1/ai/query-with-analysis" \
  -H "Content-Type: application/json" \
  -d '{"description": "分析IT部门员工薪资", "analysis_type": "detailed"}' | jq .
```

#### 6. 智能洞察

```bash
curl -X POST "http://localhost:8080/api/v1/ai/smart-insights" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "分析用户数据", "insight_level": "basic"}' | jq .
```

## 🛠️ 常用命令

```bash
make build     # 构建
make run       # 运行
make test      # 测试
make clean     # 清理
```

## 🔗 依赖项目

需要启动 `mcp-ai-server` 服务端（端口 8081）作为 MCP 协议服务提供者。

## 🔍 故障排除

1. **MCP 连接失败**: 检查 mcp-ai-server 服务是否运行在 ws://localhost:8081
2. **MySQL 连接失败**: 检查数据库服务和连接配置
3. **AI 工具错误**: 按顺序测试，从基础工具开始调试

## 📄 许可证

MIT License
