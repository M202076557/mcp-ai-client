---
marp: true
paginate: true
theme: default
title: MCP 协议实战：从协议到服务，从服务到调用
---

# MCP 协议实战
从协议到服务，从服务到调用

讲者：<你的名字>｜日期：<日期>

---

## 目录
1. 什么是 MCP（协议、场景、优势）
2. mcp-ai-server：基础能力与 AI 增强
3. 现场演示（server 本地 client）
4. mcp-ai-client：调用方封装与 API
5. 现场演示（HTTP API）
6. 安全与工程治理
7. 总结与展望、Q&A

---

## 什么是 MCP
- 定义：基于 JSON-RPC 的工具调用协议（WebSocket/stdio）
- 角色：客户端（AI/应用）⇄ 服务器（工具）
- 核心概念：Tools / Request-Response / Capabilities

---

## 使用场景
- 智能文件/项目操作、数据处理
- API 调用与响应分析
- 数据库检索与分析
- 自动化编排、统一访问接口

---

## 核心优势
- 解耦“意图”与“执行”，AI 做“翻译官”
- 调用标准化，跨系统一致
- 安全可控：白名单/限流/超时/沙箱

---

## mcp-ai-server：非 AI 基础能力
- 系统：file_read/write、command_execute、directory_list
- 数据：json_parse、base64、hash
- 网络：http_get、ping、dns_lookup
- 数据库：db_connect/query/execute（含安全限制）

---

## 演示 1（非 AI 工具）
- 文件 → 数据 → 网络 → 数据库（最小用例）
- 用 server 本地 client 执行

---

## AI 增强理念（翻译/编排）
- 自然语言 → 操作计划/命令/SQL → 执行 → 总结
- 行业映射：Cursor/Augment Agent、Copilot Workspace

---

## mcp-ai-server：AI 工具清单
- ai_chat：对话/问答
- ai_file_manager：智能文件管理
- ai_data_processor：智能数据处理
- ai_api_client：智能网络请求
- ai_query_with_analysis：自然语言查库 + 分析

---

## AI 工具：ai_chat
- 适合解释概念、回答使用疑问
- 提示词可控：语言、风格、限制

---

## AI 工具：ai_file_manager
- 生成/修改项目结构与文件
- 执行计划基于基础工具（file_* / command_execute）

---

## AI 工具：ai_data_processor
- 识别 JSON/CSV 等
- 解析、转换、结构化输出

---

## AI 工具：ai_api_client
- 根据意图构造 HTTP 请求
- 调用与响应分析

---

## AI 工具：ai_query_with_analysis
- 自然语言 → SQL → 执行 → 洞察/摘要
- SQL 安全与行数限制

---

## 演示 2（AI 工具）
- 逐个演示 5 个工具（最小用例）
- 需要准备本地模型（Ollama）

---

## mcp-ai-client：调用方架构
- 内置 MCP WebSocket 客户端
- 调 mcp-ai-server → 暴露 REST API

---

## API 映射
- /api/v1/ai/*（5 个 AI 工具）
- /api/v1/db/*（基础查询）
- 示例：file-manager / query-with-analysis / chat

---

## 演示 3（HTTP API）
- curl 调用 2–3 个 AI API + 1 个 DB 查询
- 路径隔离：target_path 在调用方重写为安全绝对路径

---

## 安全与工程治理
- SQL 白名单/禁危险 DDL、超时/限流
- 文件/命令白名单、路径清洗与重写
- 可观测性与日志

---

## 总结与展望
- MCP = 智能翻译层，Prompt 很关键
- 展望：Prompt AI 增强工具→一语生成可执行工具链

---

## Q&A
- 欢迎讨论接入、治理与最佳实践

