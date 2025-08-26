package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

// MCPClient MCP客户端 - 专门用于AI工具演示
type MCPClient struct {
	conn    *websocket.Conn
	timeout time.Duration
}

// MCPMessage MCP消息结构
type MCPMessage struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError MCP错误结构
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ToolCallParams 工具调用参数
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolCallResult 工具调用结果
type ToolCallResult struct {
	Content []Content `json:"content"`
}

// Content 内容结构
type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// NewMCPClient 创建MCP客户端
func NewMCPClient(serverURL string, timeout time.Duration) (*MCPClient, error) {
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(serverURL, nil)
	if err != nil {
		return nil, fmt.Errorf("连接MCP服务器失败: %v", err)
	}

	log.Printf("MCP服务器连接成功: %s", serverURL)
	return &MCPClient{
		conn:    conn,
		timeout: timeout,
	}, nil
}

// Close 关闭连接
func (c *MCPClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Initialize 初始化MCP连接
func (c *MCPClient) Initialize(ctx context.Context) error {
	initMsg := MCPMessage{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"clientInfo": map[string]interface{}{
				"name":    "mcp-ai-client",
				"version": "1.0.0",
			},
		},
	}

	log.Printf("发送初始化消息: %+v", initMsg)

	response, err := c.sendMessage(ctx, initMsg)
	if err != nil {
		return fmt.Errorf("初始化失败: %v", err)
	}

	log.Printf("收到初始化响应: %+v", response)

	if response.Error != nil {
		// 如果错误是"已经初始化"，则认为是成功的
		if response.Error.Code == -32000 && response.Error.Message == "Already initialized" {
			log.Println("MCP连接已经初始化，继续执行")
			return nil
		}
		return fmt.Errorf("初始化错误: %d - %s", response.Error.Code, response.Error.Message)
	}

	log.Println("MCP连接初始化成功")
	return nil
}

// CallTool 调用MCP工具
func (c *MCPClient) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*ToolCallResult, error) {
	callMsg := MCPMessage{
		JSONRPC: "2.0",
		ID:      time.Now().UnixNano(),
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": arguments,
		},
	}

	response, err := c.sendMessage(ctx, callMsg)
	if err != nil {
		return nil, fmt.Errorf("调用工具失败: %v", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("工具调用错误: %d - %s", response.Error.Code, response.Error.Message)
	}

	// 解析结果
	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		return nil, fmt.Errorf("序列化结果失败: %v", err)
	}

	var toolResult ToolCallResult
	if err := json.Unmarshal(resultBytes, &toolResult); err != nil {
		return nil, fmt.Errorf("解析工具结果失败: %v", err)
	}

	return &toolResult, nil
}

// isIDMatch 检查两个ID是否匹配
func isIDMatch(id1, id2 interface{}) bool {
	// 如果类型相同，直接比较
	if id1 == id2 {
		return true
	}

	// 处理数字类型的比较
	switch v1 := id1.(type) {
	case int:
		if v2, ok := id2.(int); ok {
			return v1 == v2
		}
		if v2, ok := id2.(float64); ok {
			return float64(v1) == v2
		}
	case float64:
		if v2, ok := id2.(float64); ok {
			return v1 == v2
		}
		if v2, ok := id2.(int); ok {
			return v1 == float64(v2)
		}
	case string:
		if v2, ok := id2.(string); ok {
			return v1 == v2
		}
	}

	// 转换为字符串进行比较，处理科学计数法
	str1 := fmt.Sprintf("%v", id1)
	str2 := fmt.Sprintf("%v", id2)

	// 如果都是数字字符串，转换为float64进行比较
	if f1, err1 := strconv.ParseFloat(str1, 64); err1 == nil {
		if f2, err2 := strconv.ParseFloat(str2, 64); err2 == nil {
			return f1 == f2
		}
	}

	return str1 == str2
}

// sendMessage 发送消息并等待响应
func (c *MCPClient) sendMessage(ctx context.Context, msg MCPMessage) (*MCPMessage, error) {
	// 设置超时上下文
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// 序列化消息
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("序列化消息失败: %v", err)
	}

	log.Printf("发送MCP消息: %s", string(msgBytes))

	// 发送消息
	if err := c.conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
		return nil, fmt.Errorf("发送消息失败: %v", err)
	}

	log.Printf("消息发送成功，等待响应...")

	// 等待响应
	responseChan := make(chan *MCPMessage, 1)
	errorChan := make(chan error, 1)

	go func() {
		// 读取响应，直到找到匹配的ID
		for {
			log.Printf("等待读取WebSocket消息...")
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket读取错误: %v", err)
				errorChan <- fmt.Errorf("读取响应失败: %v", err)
				return
			}

			log.Printf("收到MCP响应: %s", string(message))

			var response MCPMessage
			if err := json.Unmarshal(message, &response); err != nil {
				log.Printf("解析响应失败: %v", err)
				continue // 继续读取下一个消息
			}

			log.Printf("解析后的响应: ID=%v, Method=%s, Error=%v", response.ID, response.Method, response.Error)

			// 检查是否是我们要的响应（ID匹配）
			if isIDMatch(msg.ID, response.ID) {
				log.Printf("找到匹配的响应ID: %v", response.ID)
				responseChan <- &response
				return
			} else {
				log.Printf("收到不匹配的响应ID: 期望 %v, 实际 %v", msg.ID, response.ID)
			}
		}
	}()

	// 等待响应或超时
	select {
	case response := <-responseChan:
		log.Printf("成功收到响应")
		return response, nil
	case err := <-errorChan:
		log.Printf("读取响应时发生错误: %v", err)
		return nil, err
	case <-ctx.Done():
		log.Printf("等待响应超时")
		return nil, fmt.Errorf("等待响应超时")
	}
}

// AI工具方法 - 集成5种AI工具功能 (5.1-5.5)

// CallAIChat 调用AI聊天工具 (5.1)
func (c *MCPClient) CallAIChat(ctx context.Context, prompt string, provider string, model string) (string, error) {
	args := map[string]interface{}{
		"prompt": prompt,
	}

	// 添加可选参数
	if provider != "" {
		args["provider"] = provider
	}
	if model != "" {
		args["model"] = model
	}

	result, err := c.CallTool(ctx, "ai_chat", args)
	if err != nil {
		return "", fmt.Errorf("AI聊天调用失败: %v", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("AI聊天结果为空")
	}

	return result.Content[0].Text, nil
}

// CallAIFileManager 调用AI文件管理工具 (5.2)
func (c *MCPClient) CallAIFileManager(ctx context.Context, instruction string, targetPath string, operationMode string) (string, error) {
	args := map[string]interface{}{
		"instruction":    instruction,
		"target_path":    targetPath,
		"operation_mode": operationMode,
	}

	result, err := c.CallTool(ctx, "ai_file_manager", args)
	if err != nil {
		return "", fmt.Errorf("AI文件管理调用失败: %v", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("AI文件管理结果为空")
	}

	return result.Content[0].Text, nil
}

// CallAIDataProcessor 调用AI数据处理工具 (5.3)
func (c *MCPClient) CallAIDataProcessor(ctx context.Context, instruction string, inputData string, dataType string, outputFormat string, operationMode string) (string, error) {
	args := map[string]interface{}{
		"instruction":    instruction,
		"input_data":     inputData,
		"data_type":      dataType,
		"output_format":  outputFormat,
		"operation_mode": operationMode,
	}

	result, err := c.CallTool(ctx, "ai_data_processor", args)
	if err != nil {
		return "", fmt.Errorf("AI数据处理调用失败: %v", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("AI数据处理结果为空")
	}

	return result.Content[0].Text, nil
}

// CallAIAPIClient 调用AI网络请求工具 (5.4)
func (c *MCPClient) CallAIAPIClient(ctx context.Context, instruction string, baseURL string, requestMode string, responseAnalysis bool) (string, error) {
	args := map[string]interface{}{
		"instruction":       instruction,
		"base_url":          baseURL,
		"request_mode":      requestMode,
		"response_analysis": responseAnalysis,
	}

	result, err := c.CallTool(ctx, "ai_api_client", args)
	if err != nil {
		return "", fmt.Errorf("AI网络请求调用失败: %v", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("AI网络请求结果为空")
	}

	return result.Content[0].Text, nil
}

// CallAIQueryWithAnalysis 调用AI数据库查询工具 (5.5)
func (c *MCPClient) CallAIQueryWithAnalysis(ctx context.Context, description string, analysisType string, tableName string) (string, error) {
	args := map[string]interface{}{
		"description":   description,
		"analysis_type": analysisType,
	}

	// 添加可选参数
	if tableName != "" {
		args["table_name"] = tableName
	}

	result, err := c.CallTool(ctx, "ai_query_with_analysis", args)
	if err != nil {
		return "", fmt.Errorf("AI数据库查询调用失败: %v", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("AI数据库查询结果为空")
	}

	return result.Content[0].Text, nil
}
