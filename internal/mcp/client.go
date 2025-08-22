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

// MCPClient MCP客户端
type MCPClient struct {
	conn    *websocket.Conn
	timeout time.Duration
	// 添加配置信息
	dbAlias  string
	dbDriver string
	dbDSN    string
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
func NewMCPClient(serverURL string, timeout time.Duration, dbAlias, dbDriver, dbDSN string) (*MCPClient, error) {
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(serverURL, nil)
	if err != nil {
		return nil, fmt.Errorf("连接MCP服务器失败: %v", err)
	}

	log.Printf("MCP服务器连接成功: %s", serverURL)
	return &MCPClient{
		conn:     conn,
		timeout:  timeout,
		dbAlias:  dbAlias,
		dbDriver: dbDriver,
		dbDSN:    dbDSN,
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

// QueryUserViaMCP 通过MCP查询mcp_user表
func (c *MCPClient) QueryUserViaMCP(ctx context.Context) (string, error) {
	// 首先建立数据库连接
	connectArgs := map[string]interface{}{
		"driver": c.dbDriver,
		"dsn":    c.dbDSN,
		"alias":  c.dbAlias,
	}

	_, err := c.CallTool(ctx, "db_connect", connectArgs)
	if err != nil {
		return "", fmt.Errorf("MCP数据库连接失败: %v", err)
	}

	// 然后执行查询
	queryArgs := map[string]interface{}{
		"alias": c.dbAlias, // 使用数据库连接别名
		"sql":   "SELECT * FROM `mcp_user` LIMIT 100",
	}

	result, err := c.CallTool(ctx, "db_query", queryArgs)
	if err != nil {
		return "", fmt.Errorf("MCP查询mcp_user表失败: %v", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("MCP查询结果为空")
	}

	return result.Content[0].Text, nil
}

// QueryUserByIDViaMCP 通过MCP根据ID查询mcp_user表
func (c *MCPClient) QueryUserByIDViaMCP(ctx context.Context, id int) (string, error) {
	// 首先建立数据库连接
	connectArgs := map[string]interface{}{
		"driver": c.dbDriver,
		"dsn":    c.dbDSN,
		"alias":  c.dbAlias,
	}

	_, err := c.CallTool(ctx, "db_connect", connectArgs)
	if err != nil {
		return "", fmt.Errorf("MCP数据库连接失败: %v", err)
	}

	// 然后执行查询
	queryArgs := map[string]interface{}{
		"alias": c.dbAlias, // 使用数据库连接别名
		"sql":   fmt.Sprintf("SELECT * FROM `mcp_user` WHERE id = %d", id),
	}

	result, err := c.CallTool(ctx, "db_query", queryArgs)
	if err != nil {
		return "", fmt.Errorf("MCP查询mcp_user表失败: %v", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("MCP查询结果为空")
	}

	return result.Content[0].Text, nil
}

// GetMenuCountViaMCP 通过MCP获取menu表记录数
func (c *MCPClient) GetMenuCountViaMCP(ctx context.Context) (string, error) {
	// 首先建立数据库连接
	connectArgs := map[string]interface{}{
		"driver": c.dbDriver,
		"dsn":    c.dbDSN,
		"alias":  c.dbAlias,
	}

	_, err := c.CallTool(ctx, "db_connect", connectArgs)
	if err != nil {
		return "", fmt.Errorf("MCP数据库连接失败: %v", err)
	}

	// 然后执行查询
	queryArgs := map[string]interface{}{
		"alias": c.dbAlias, // 使用数据库连接别名
		"sql":   "SELECT COUNT(*) as count FROM menu",
	}

	result, err := c.CallTool(ctx, "db_query", queryArgs)
	if err != nil {
		return "", fmt.Errorf("MCP查询menu表记录数失败: %v", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("MCP查询结果为空")
	}

	return result.Content[0].Text, nil
}
