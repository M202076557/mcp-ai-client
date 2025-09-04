# 演示指南（与幻灯片页码对应）

注：本指南与 share/MCP_Talk_Slides_Marp.md 幻灯片的页码一一对应，便于现场操作。

## 环境准备
- mcp-ai-server（WebSocket 模式，建议 8081）
- mcp-ai-server 本地 client（用于演示 1、2）
- mcp-ai-client（HTTP 服务，建议 8080）
- MySQL（mcp_test 库，存在 mcp_user 表或按脚本创建）
- 如用 AI：Ollama 已启动并拉取需要的模型（codellama:7b、llama3.2:1b）

---

## 幻灯片 (7) 演示 1：非 AI 工具（server 本地 client）
1) 启动 server（另一个终端）：
   - ./bin/mcp-server -mode=websocket -port=8081
2) 启动本地 client：
   - ./bin/mcp-client
3) 执行最小用例（片段，参考 test/docs/QUICK_TEST_SCRIPT.md）：
   - call file_write path:"./demo.txt" content:"MCP Server 演示文件"
   - call file_read path:"./demo.txt"
   - call http_get url:"https://httpbin.org/json"
   - call db_connect driver:"mysql" dsn:"root:root@tcp(127.0.0.1:3306)/mcp_test" alias:"demo"
   - call db_execute alias:"demo" sql:"CREATE TABLE IF NOT EXISTS mcp_user (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(50), email VARCHAR(100), department VARCHAR(30), age INT, salary DECIMAL(10,2))"

---

## 幻灯片 (15) 演示 2：AI 工具（server 本地 client）
依次执行（每个只做最小用例）：
- call ai_chat prompt:"你好，请介绍一下MCP协议是什么？50字以内"
- call ai_file_manager instruction:"创建一个Go项目的标准目录结构" target_path:"./demo-go-project" operation_mode:"execute"
- call ai_data_processor instruction:"解析这个JSON数据并提取所有用户的邮箱地址" input_data:'{"users":[{"name":"张三","email":"zhangsan@example.com"},{"name":"李四","email":"lisi@example.com"}]}' data_type:"json" output_format:"table" operation_mode:"execute"
- call ai_api_client instruction:"获取测试数据" base_url:"https://httpbin.org" request_mode:"execute" response_analysis:true
- call ai_query_with_analysis description:"查询所有员工信息" analysis_type:"insights" table_name:"mcp_user"

注意：如果使用 ai_file_manager，请确认调用方路径隔离逻辑已在 mcp-ai-client 中生效（见下一个演示）。

---

## 幻灯片 (18) 演示 3：HTTP API（mcp-ai-client）
1) 启动 mcp-ai-client（另一个终端）：
   - ./bin/mcp-ai-client
2) 健康检查：
   - curl -s http://localhost:8080/health | jq .
3) 基础 DB 查询（对齐 mcp_user 示例表）：
   - curl -s http://localhost:8080/api/v1/db/users | jq .
4) AI 对话：
   - curl -s -X POST http://localhost:8080/api/v1/ai/chat -H 'Content-Type: application/json' -d '{"prompt":"请介绍一下MCP协议的主要特点"}' | jq .
5) AI 文件管理（验证路径重写在调用方生效）：
   - curl -s -X POST http://localhost:8080/api/v1/ai/file-manager -H 'Content-Type: application/json' -d '{"instruction":"创建一个Go项目的标准目录结构","target_path":"./demo-go-project","operation_mode":"execute"}' | jq .
   - 预期：demo-go-project 目录在“调用方”工作目录创建，而非服务端目录。
6) AI 智能数据库查询：
   - curl -s -X POST http://localhost:8080/api/v1/ai/query-with-analysis -H 'Content-Type: application/json' -d '{"description":"查询所有员工信息","analysis_type":"insights","table_name":"mcp_user"}' | jq .

---

## 清理与复位
- server 侧：
  - call db_execute alias:"demo" sql:"DROP TABLE IF EXISTS mcp_user"
  - 删除临时文件/目录（若需要）
- client 侧：
  - rm -rf ./demo-go-project ./demo.txt（确认路径无误后再执行）

---

## 常见问题
- AI 超时或报错：检查模型是否可用与网络；适当增大超时。
- DB 查询字段不对齐：确认客户端数据类型转换逻辑已兼容 []uint8/string（问题已修复）。
- 文件落到服务端：确认调用方 target_path 路径清洗与绝对化逻辑已生效。

