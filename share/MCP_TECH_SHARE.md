## MCP 技术分享：从协议到实战（mcp-ai-server 与 mcp-ai-client）

### 演示版 PPT 大纲（可直接做成幻灯片）

- 封面

  - 标题：MCP 协议实战：从协议到服务，从服务到调用
  - 讲者/日期

- 目录

  1. 什么是 MCP（协议、场景、优势）
  2. mcp-ai-server：基础能力与 AI 增强
  3. 现场演示（server 本地 client）
  4. mcp-ai-client：作为调用方的封装与 API
  5. 现场演示（HTTP API）
  6. 安全与工程治理
  7. 总结与展望、Q&A

- MCP 协议简介

  - 定义：基于 JSON-RPC 的“工具调用协议 ”（WebSocket/stdio）
  - 角色：客户端（AI/应用）⇄ 服务器（工具）
  - 核心概念：Tools/Request-Response/Capabilities

- 使用场景

  - 智能文件/项目操作、数据处理、API 调用
  - 数据库检索与分析、自动化编排、统一访问接口

- 核心优势

  - 解耦“意图”与“执行”，AI 做“翻译”
  - 标准化工具调用，跨系统一致
  - 安全可控：白名单/限流/超时/沙箱

- mcp-ai-server：非 AI 基础能力

  - 系统：file_read/write、command_execute、directory_list
  - 数据：json_parse、base64、hash
  - 网络：http_get、ping、dns_lookup
  - 数据库：db_connect/query/execute（含安全限制）

- 演示 1：非 AI 工具（用本地 client）

  - 文件 → 数据 → 网络 → 数据库（最小用例）

- AI 增强理念

  - 自然语言 → 操作计划/命令/SQL → 执行 → 总结
  - 行业对齐：Cursor/Augment Agent、Copilot Workspace

- mcp-ai-server：AI 工具

  - ai_chat / ai_file_manager / ai_data_processor / ai_api_client / ai_query_with_analysis

- 演示 2：AI 工具（用本地 client）

  - 逐个演示 5 个工具（最小用例）

- mcp-ai-client：调用方架构

  - MCP Client → 调 server → 暴露 REST API
  - API 映射：/api/v1/ai/_、/api/v1/db/_

- 演示 3：HTTP API

  - 演示 2–3 个 AI API + 1 个 DB API
  - 路径隔离：target_path 在调用方重写为安全绝对路径

- 安全与工程治理

  - SQL 白名单/禁危险 DDL、超时/限流
  - 文件/命令白名单、路径清洗与重写

- 总结与展望

  - MCP = 智能翻译层；Prompt 很关键
  - 展望：Prompt AI 增强工具：一句话生成可执行工具链并编排执行

- Q&A

---

### 演讲稿（口语化讲稿，10–15 分钟，含演示）

大家好，今天我们来聊聊 MCP 协议，以及我们在这个协议之上做的两件事：
第一，在 mcp-ai-server 上把工具能力标准化，并用 AI 做了增强；
第二，在 mcp-ai-client 里把 MCP 能力封装成 HTTP API，让业务更容易用起来。

先说 MCP 是什么。它可以理解成一套“让 AI/应用去安全调用工具”的协议。消息是 JSON-RPC，通道可以是 WebSocket 或者 stdio。优势有三个：

- 一是解耦“意图”和“执行”。人说自然语言，AI 负责翻译，MCP 负责连接工具并执行；
- 二是工具标准化。跨系统、跨语言，用统一的接口来调用；
- 三是可治理。我们可以做白名单、超时、限流、沙箱策略，让自动化“可控、可审计”。

有哪些用法？做智能文件/项目操作、数据格式转换、自动调用第三方 API、让业务同学也能一句话查数据库并拿到分析结果。这些在日常里非常实用。

好，先来看看 mcp-ai-server 提供的基础能力。我们提供了几类常用工具：

- 系统类：读写文件、执行命令、列目录；
- 数据类：JSON 解析、Base64 编解码、哈希；
- 网络类：HTTP GET、ping、DNS 解析；
- 数据库类：连接、查询、执行，默认也做了安全限制，比如关闭 DROP/TRUNCATE 这类危险指令。

我们先用 server 自带的本地 client 演示一下非 AI 的工具。顺序很简单：
创建文件 → 读取 → 做一次 JSON 解析 → 发个 HTTP 请求 → 跑一遍数据库的 CRUD。
这部分是“地基”，稳定很关键。大家先建立一个直觉：这些工具就像积木，随时可以被 AI 编排起来。

接下来是今天的重点：AI 增强。我们的思路是让 AI 做“翻译/编排”。
当用户说“创建一个 Go 项目的标准目录结构”，AI 会先理解“你想做什么”，然后把它翻译成一组文件操作计划，再用基础工具一步步去执行。
这和行业里一些产品的 Agent 体验很像：

- 比如 Cursor、Augment 会根据你的意图，自动改写文件、生成代码；
- Copilot Workspace 也是“先规划、再执行”，最后给你产出结果。

在 mcp-ai-server 里，我们做了 5 个 AI 工具：

1. ai_chat：基础对话、问答；
2. ai_file_manager：智能文件管理，能根据意图创建/修改项目结构；
3. ai_data_processor：识别 JSON/CSV 等数据并做解析、转换；
4. ai_api_client：根据自然语言生成 HTTP 请求，调用并分析响应；
5. ai_query_with_analysis：一句话查数据库，AI 生成 SQL、执行、并且给出结果分析/摘要。

我们用 server 的本地 client 逐个演示一下最小用例。这里有个小提醒：用 Ollama 的同学，记得先把模型准备好，比如 codellama:7b、llama3.2:1b。

到这儿为止，MCP 的基础和我们在 server 侧的实现大家都看过了。接下来聊聊作为调用方，怎么真正把这套能力接到业务里。

在 mcp-ai-client 里，我们内置了一个 MCP WebSocket 客户端，它连到 mcp-ai-server，然后我们把这些能力包装成简单的 HTTP API 暴露出去。
也就是说，业务可以不懂 MCP 细节，直接用 /api/v1/ai/_ 这几个接口，就能用上 AI 增强的能力；/api/v1/db/_ 则是一些基础的数据库查询接口。

我们现场演示一下：

- 用 curl 调用 ai_file_manager，新建一个 demo-go-project 目录；
- 再来一次 ai_query_with_analysis，让 AI 直接查库并做分析；
- 最后用 ai_chat 问一个简短问题，看看它的响应。

安全提醒也很重要。像文件相关的 target_path，我们在调用方会做路径清洗与重写：相对路径一律锚定到调用方的工作目录，去掉 ~ 和 ..，这样就不会写到服务提供方的目录里。这在产品化时是必须的边界隔离。

到尾声，我们简单做一个总结。MCP 其实就是“智能翻译层”，把自然语言变成可执行的操作。Prompt 在这里非常关键，它决定了 AI 翻译的质量。
未来我们还能在 MCP 上做一个“Prompt 增强工具”，它负责读懂一句话，自动生成一串可执行工具链，然后编排执行、输出结果报告。这样从“理解 → 执行 → 汇报”就是全自动的闭环。当然今天我们就先到这里，这部分算是小小的展望。

最后欢迎大家提问，尤其是关于“如何在你们的项目里接入这套能力、如何做安全治理与审计、以及有哪些落地最佳实践”的问题。谢谢大家！

---

面向对象：研发与非研发同学。目标是在 30–60 分钟内，帮助大家理解 MCP 协议、看懂我们在 server 与 client 两端的实现，并能跟随文档完成可复现的演示。

---

### 全局架构与角色

- mcp-ai-server：模拟的 MCP 服务提供方，暴露“工具”能力（系统、网络、数据、数据库、AI 等）
- mcp-ai-server 本地 CLI client：用于演示 server 的工具调用（纯 MCP 调用）
- mcp-ai-client：模拟 MCP 服务的业务调用方，通过内置 MCP 客户端调用 server，再用 HTTP API 对外提供能力

交互链路（简化）：HTTP/CLI → MCP Client → MCP Server → 工具执行（系统/DB/AI…）→ 结果

---

## 一、MCP 协议简介（给非研发也能听懂的版本）

- MCP（Model Context Protocol）是一套“让 AI/应用安全地调用工具”的开放协议。
- 传输层常见为 JSON-RPC 2.0 over WebSocket 或 stdio。
- 核心概念：
  - Tool（工具）：一个可调用的能力（如 file_read、db_query、ai_chat）
  - Request/Response：标准 JSON-RPC 请求/响应格式
  - Capabilities：服务端声明支持哪些工具与元数据
- 我们的实现要点：
  - WebSocket 模式，路径与端口可配置
  - 工具分层清晰：非 AI 工具（系统/网络/数据/数据库）+ AI 增强工具（chat、文件管理、数据处理、API 调用、智能查询）

面向研发的补充：MCP 将“意图”和“执行”解耦。AI 常作为“翻译器”，把人类自然语言翻译成结构化调用（SQL、HTTP 请求、文件操作计划等）。

---

## 二、mcp-ai-server 能力总览

### 2.1 非 AI 工具（基础、可组合）

- 系统类：file_read、file_write、command_execute、directory_list
- 数据类：json_parse、base64_encode、hash
- 网络类：http_get、ping、dns_lookup
- 数据库类：db_connect、db_query、db_execute（含 DDL/DML；有安全限制如禁 DROP 等）

示例（片段，来自演示脚本）：
<augment_code_snippet path="test/docs/QUICK_TEST_SCRIPT.md" mode="EXCERPT">

```md
# 创建演示文件

call file_write path:"./demo.txt" content:"MCP Server 演示文件"

# 读取文件

call file_read path:"./demo.txt"

# 执行系统命令

call command_execute command:"date"
```

</augment_code_snippet>

### 2.2 AI 增强工具（在非 AI 工具之上“加智能”）

- ai_chat：对话/问答（文本生成）
- ai*file_manager：理解“意图”，生成/修改项目结构（底层用 file*\*、command_execute）
- ai_data_processor：理解数据、解析转换（底层用 json/encoding 等）
- ai_api_client：根据意图构造 HTTP 请求并分析响应（底层用 http_get 等）
- ai*query_with_analysis：自然语言 → SQL 生成 → 安全执行 → 结果分析（底层用 db*\*）

配置中的“功能特定模型（function_models）”将不同任务路由到更合适的模型：
<augment_code_snippet path="configs/config.yaml" mode="EXCERPT">

```yaml
function_models:
  sql_generation:
    provider: "ollama"
    model: "codellama:7b"
  data_analysis:
    provider: "ollama"
    model: "llama3.2:1b"
```

</augment_code_snippet>

要点：AI 工具不是“另一个世界”，而是“对非 AI 工具的增强编排层”。例如 AI 文件管理会先理解意图，再调用 file_write/command_execute 完成操作。

---

## 三、结合本地 CLI Client 的现场演示（基于 mcp-ai-server）

准备：

- 启动 Server（WebSocket 模式）：
  - ./bin/mcp-server -mode=websocket -port=8081
- 启动本地 CLI 客户端：
  - ./bin/mcp-client（进入交互式 call … 命令）

建议演示顺序：

1. 系统工具
   <augment_code_snippet path="test/docs/QUICK_TEST_SCRIPT.md" mode="EXCERPT">

```md
call directory_list path:"."
call command_execute command:"date"
```

</augment_code_snippet>

2. 数据处理
   <augment_code_snippet path="test/docs/QUICK_TEST_SCRIPT.md" mode="EXCERPT">

```md
call json_parse json_string:'{"name":"MCP Demo"}' pretty:true
call base64_encode text:"Hello MCP Server"
```

</augment_code_snippet>

3. 网络
   <augment_code_snippet path="test/docs/QUICK_TEST_SCRIPT.md" mode="EXCERPT">

```md
call http_get url:"https://httpbin.org/json"
call dns_lookup domain:"github.com"
```

</augment_code_snippet>

4. 数据库（连接 → 建表 →CRUD→ 清理）
   <augment_code_snippet path="test/docs/QUICK_TEST_SCRIPT.md" mode="EXCERPT">

```md
call db_connect driver:"mysql" dsn:"root:root@tcp(127.0.0.1:3306)/mcp_test" alias:"demo"
call db_execute alias:"demo" sql:"CREATE TABLE IF NOT EXISTS mcp_user (...)"
```

</augment_code_snippet>

5. AI 工具
   <augment_code_snippet path="test/docs/QUICK_TEST_SCRIPT.md" mode="EXCERPT">

```md
call ai_chat prompt:"你好，请介绍一下 MCP 协议是什么？50 字以内"
call ai_query_with_analysis description:"查询所有员工信息" analysis_type:"insights" table_name:"mcp_user"
```

</augment_code_snippet>

注意事项：

- 需要本地 MySQL（mcp_test 库）
- 使用 Ollama 时需先启动服务并准备模型（如 codellama:7b、llama3.2:1b）
- 如使用 OpenAI/Anthropic，需设置环境变量并在配置中启用

---

## 四、mcp-ai-client（调用方）项目：把 MCP 能力“API 化”

定位：通过内置的 MCP WebSocket 客户端连接 mcp-ai-server，把工具能力以 REST API 暴露给业务侧。

典型 API 映射（片段）：

- POST /api/v1/ai/chat → 调用 ai_chat
- POST /api/v1/ai/file-manager → 调用 ai_file_manager
- POST /api/v1/ai/data-processor → 调用 ai_data_processor
- POST /api/v1/ai/api-client → 调用 ai_api_client
- POST /api/v1/ai/query-with-analysis → 调用 ai_query_with_analysis
- GET /api/v1/db/users → 基础数据库查询（示例）

演示（用 curl 即可）：
<augment_code_snippet path="test/docs/AI_TOOLS_TEST_GUIDE.md" mode="EXCERPT">

```md
curl -X POST "http://localhost:8080/api/v1/ai/chat" \
 -H "Content-Type: application/json" \
 -d '{"prompt":"你好，请介绍一下 MCP 协议是什么？50 字以内。"}'
```

</augment_code_snippet>

<augment_code_snippet path="test/docs/AI_TOOLS_TEST_GUIDE.md" mode="EXCERPT">

```md
curl -X POST "http://localhost:8080/api/v1/ai/file-manager" \
 -H "Content-Type: application/json" \
 -d '{"instruction":"创建一个 Go 项目的标准目录结构","target_path":"./demo-go-project","operation_mode":"execute"}'
```

</augment_code_snippet>

配置关键点（mcp-ai-client）：

- MCP Server 地址：ws://localhost:8081（可在 configs/config.yaml 设置）
- 数据库连接、默认 AI 提供商/模型等可配置

---

## 五、讲解与演示串联脚本（建议 15–20 分钟）

1. 3 分钟：什么是 MCP？为什么要“AI 作为翻译器”？
2. 5 分钟：看 mcp-ai-server 的非 AI 工具 → 再看 AI 工具如何“增强组合”
3. 5 分钟：切到本地 CLI client，按 3 节顺序执行脚本片段，展示端到端能力
4. 5 分钟：切换到 mcp-ai-client，用 HTTP API 连通 server，演示 2–3 个典型接口
5. Q&A：模型选择、安全策略、落地场景讨论

---

## 六、给研发同学的“深潜”

- 协议与传输：JSON-RPC 2.0 over WebSocket；方法名即工具名，参数/返回为结构化 JSON。
- AI 编排：按“功能特定模型”路由到更合适的 LLM，降温/长度等公共参数可控。
- 安全与治理：
  - SQL 安全（白名单、禁用危险 DDL、行数限制等）
  - 文件/命令白名单与路径控制
  - 资源限制（最大响应/超时/并发）
- 设计理念：非 AI 工具是“地基”，AI 工具是“智能管家”。先稳，再巧。

---

## 七、常见问题与排障

- 启动后 AI 接口无响应？检查 AI 提供商是否启用、模型是否存在、网络连通性。
- 数据库相关报错？确认 MySQL 连接、库表是否存在；遵循安全限制（禁 DROP/TRUNCATE 等）。
- HTTP API 调用 504/超时？查看 server 与 client 的超时配置，缩小请求数据量或提高限额。

---

## 附录：快捷清单（可复制演示）

- Server 端（CLI 演示起手式）
  <augment_code_snippet path="test/docs/QUICK_TEST_SCRIPT.md" mode="EXCERPT">

```md
./bin/mcp-server -mode=websocket -port=8081
./bin/mcp-client

# 连接后执行 call ... 命令
```

</augment_code_snippet>

- Client 端（HTTP 演示起手式）
  <augment_code_snippet path="test/docs/AI_TOOLS_TEST_GUIDE.md" mode="EXCERPT">

```md
curl -X GET "http://localhost:8080/api/v1/db/users" | jq .
```

</augment_code_snippet>

参考资料：

- MCP 官方：https://modelcontextprotocol.io/
- Anthropic MCP 文档：https://docs.anthropic.com/en/docs/build-with-claude/mcp
- 本项目演示脚本与指南：test/docs/QUICK_TEST_SCRIPT.md、test/docs/AI_TOOLS_TEST_GUIDE.md
