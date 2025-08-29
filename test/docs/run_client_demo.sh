#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

echo "[1/5] ensure server is up (ws://localhost:8081)"
if ! curl -sS http://localhost:8081/health >/dev/null; then
  echo "server not up; please run: /Users/ksc/Desktop/study/mcp-ai-server/bin/mcp-server -mode websocket -port 8081" >&2
  exit 1
fi

echo "[2/5] build client"
make build >/dev/null || true

echo "[3/5] start client"
./bin/mcp-ai-client >/tmp/mcp-ai-client.demo.log 2>&1 &
CLIENT_PID=$!
sleep 1

echo "[4/5] health checks"
curl -sS http://localhost:8080/health | jq .

echo "[5/5] run demo calls"
set +e
curl -sS "http://localhost:8080/api/v1/db/users" | jq .
curl -sS -X POST "http://localhost:8080/api/v1/ai/chat" -H "Content-Type: application/json" -d '{"prompt":"你好，请介绍一下MCP协议是什么？50字以内。","max_tokens":50}' | jq .
curl -sS -X POST "http://localhost:8080/api/v1/ai/chat" -H "Content-Type: application/json" -d '{"prompt":"解释一下Go语言的并发特性","provider":"ollama","model":"codellama:7b"}' | jq .
curl -sS -X POST "http://localhost:8080/api/v1/ai/file-manager" -H "Content-Type: application/json" -d '{"instruction":"创建一个Go项目的标准目录结构","target_path":"./demo-go-project","operation_mode":"execute"}' | jq .
curl -sS -X POST "http://localhost:8080/api/v1/ai/data-processor" -H "Content-Type: application/json" -d '{"instruction":"解析这个JSON数据并提取所有用户的邮箱地址","input_data":"{\"users\":[{\"name\":\"张三\",\"email\":\"zhangsan@example.com\",\"age\":25},{\"name\":\"李四\",\"email\":\"lisi@example.com\",\"age\":30}]}","data_type":"json","output_format":"table","operation_mode":"execute"}' | jq .
curl -sS -X POST "http://localhost:8080/api/v1/ai/api-client" -H "Content-Type: application/json" -d '{"instruction":"获取用户数据","base_url":"https://jsonplaceholder.typicode.com","request_mode":"execute","response_analysis":true}' | jq .
curl -sS -X POST "http://localhost:8080/api/v1/ai/api-client" -H "Content-Type: application/json" -d '{"instruction":"获取测试数据","base_url":"https://httpbin.org","request_mode":"execute","response_analysis":true}' | jq .
curl -sS -X POST "http://localhost:8080/api/v1/ai/query-with-analysis" -H "Content-Type: application/json" -d '{"description":"查询所有员工信息","analysis_type":"insights","table_name":"mcp_user"}' | jq .

echo "done. logs: /tmp/mcp-ai-client.demo.log"

