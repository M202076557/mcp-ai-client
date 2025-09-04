package api

import (
	"context"
	"encoding/json"
	"mcp-ai-client/internal/database"
	"mcp-ai-client/internal/mcp"
	"mcp-ai-client/internal/service"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AIConfig AI工具配置
type AIConfig struct {
	ResponseLanguage           string
	DefaultProvider            string
	DefaultModel               string
	IncludeLanguageInstruction bool
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	UserTable string
}

// Handlers API处理器 - 简化版，只保留AI工具和基础数据库查询
type Handlers struct {
	mysqlClient *database.MySQLClient
	mcpClient   *mcp.MCPClient
	userService *service.UserService
	aiConfig    *AIConfig
	dbConfig    *DatabaseConfig
}

// NewHandlers 创建API处理器
func NewHandlers(mysqlClient *database.MySQLClient, mcpClient *mcp.MCPClient, aiConfig *AIConfig, dbConfig *DatabaseConfig) *Handlers {
	// 创建服务层
	userService := service.NewUserService(mysqlClient, dbConfig.UserTable)

	return &Handlers{
		mysqlClient: mysqlClient,
		mcpClient:   mcpClient,
		userService: userService,
		aiConfig:    aiConfig,
		dbConfig:    dbConfig,
	}
}

// getLanguageInstruction 根据配置生成语言指令
func (h *Handlers) getLanguageInstruction() string {
	if !h.aiConfig.IncludeLanguageInstruction {
		return ""
	}

	switch h.aiConfig.ResponseLanguage {
	case "zh-CN":
		return "请用中文回答。"
	case "en-US":
		return "Please respond in English."
	case "auto":
		return "请根据用户的语言进行回答。Please respond in the user's language."
	default:
		return "请用中文回答。"
	}
}

// applyDefaultAIParams 应用默认AI参数
func (h *Handlers) applyDefaultAIParams(args map[string]interface{}) {
	if _, exists := args["provider"]; !exists {
		args["provider"] = h.aiConfig.DefaultProvider
	}
	if _, exists := args["model"]; !exists {
		args["model"] = h.aiConfig.DefaultModel
	}
}

// ===== 健康检查 =====

// HealthCheck 健康检查
func (h *Handlers) HealthCheck(c *gin.Context) {
	status := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "MCP AI Client - 简化版",
		"version":   "2.0.0",
	}

	// 检查MySQL连接
	if h.mysqlClient != nil {
		status["mysql"] = "connected"
	} else {
		status["mysql"] = "not_configured"
	}

	// 检查MCP连接
	if h.mcpClient != nil {
		status["mcp"] = "connected"
	} else {
		status["mcp"] = "not_configured"
	}

	c.JSON(http.StatusOK, status)
}

// ===== 基础数据库查询API =====

// GetUsersTraditional 传统方式获取用户列表
func (h *Handlers) GetUsersTraditional(c *gin.Context) {
	if h.userService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "用户服务不可用",
		})
		return
	}

	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     err.Error(),
			"method":    "traditional",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      users,
		"count":     len(users),
		"method":    "traditional_database",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// ===== AI工具处理器 (5.1-5.5) =====

// MCPChatHandler 5.1 基础AI对话
func (h *Handlers) MCPChatHandler(c *gin.Context) {
	start := time.Now()

	if h.mcpClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "MCP服务不可用",
			"tool":  "ai_chat",
		})
		return
	}

	var request struct {
		Prompt      string  `json:"prompt" binding:"required"`
		Provider    string  `json:"provider"`
		Model       string  `json:"model"`
		MaxTokens   int     `json:"max_tokens"`
		Temperature float64 `json:"temperature"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"tool":    "ai_chat",
		})
		return
	}

	// 构建MCP调用参数
	args := map[string]interface{}{
		"prompt": request.Prompt + " " + h.getLanguageInstruction(),
	}

	if request.Provider != "" {
		args["provider"] = request.Provider
	}
	if request.Model != "" {
		args["model"] = request.Model
	}
	if request.MaxTokens > 0 {
		args["max_tokens"] = request.MaxTokens
	}
	if request.Temperature > 0 {
		args["temperature"] = request.Temperature
	}

	// 应用默认AI参数
	h.applyDefaultAIParams(args)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := h.mcpClient.CallTool(ctx, "ai_chat", args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "AI chat failed",
			"details":  err.Error(),
			"duration": time.Since(start).String(),
			"tool":     "ai_chat",
		})
		return
	}

	// 解析MCP返回的结果
	var mcpResponse struct {
		Tool     string `json:"tool"`
		Status   string `json:"status"`
		Prompt   string `json:"prompt"`
		Provider string `json:"provider"`
		Model    string `json:"model"`
		Response string `json:"response"`
	}

	var responseData map[string]interface{}
	if err := json.Unmarshal([]byte(result.Content[0].Text), &mcpResponse); err == nil {
		responseData = map[string]interface{}{
			"tool":     "ai_chat",
			"status":   mcpResponse.Status,
			"prompt":   request.Prompt,
			"response": mcpResponse.Response,
			"duration": time.Since(start).String(),
		}
		if mcpResponse.Provider != "" {
			responseData["provider"] = mcpResponse.Provider
		}
		if mcpResponse.Model != "" {
			responseData["model"] = mcpResponse.Model
		}
	} else {
		responseData = map[string]interface{}{
			"tool":     "ai_chat",
			"status":   "success",
			"prompt":   request.Prompt,
			"response": result.Content[0].Text,
			"duration": time.Since(start).String(),
		}
	}

	c.JSON(http.StatusOK, responseData)
}

// MCPFileManagerHandler 5.2 AI智能文件管理
func (h *Handlers) MCPFileManagerHandler(c *gin.Context) {
	start := time.Now()

	if h.mcpClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "MCP服务不可用",
			"tool":  "ai_file_manager",
		})
		return
	}

	var request struct {
		Instruction   string `json:"instruction" binding:"required"`
		TargetPath    string `json:"target_path"`
		OperationMode string `json:"operation_mode"`
		Provider      string `json:"provider"`
		Model         string `json:"model"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"tool":    "ai_file_manager",
		})
		return
	}

	// 构建MCP调用参数
	args := map[string]interface{}{
		"instruction": request.Instruction + " " + h.getLanguageInstruction(),
	}

	// 将相对/特殊路径重写为调用方的工作目录下的安全绝对路径，避免影响服务提供方
	if request.TargetPath != "" {
		var cleaned string = request.TargetPath
		// 去除可能的危险前缀
		cleaned = strings.TrimSpace(cleaned)
		cleaned = strings.TrimPrefix(cleaned, "~")
		cleaned = strings.ReplaceAll(cleaned, "..", "")
		// 统一将相对路径锚定到当前进程工作目录
		cwd, _ := os.Getwd()
		abs := cleaned
		if !filepath.IsAbs(cleaned) {
			abs = filepath.Join(cwd, cleaned)
		}
		// 规范化
		abs = filepath.Clean(abs)
		args["target_path"] = abs
	}
	if request.OperationMode != "" {
		args["operation_mode"] = request.OperationMode
	}
	if request.Provider != "" {
		args["provider"] = request.Provider
	}
	if request.Model != "" {
		args["model"] = request.Model
	}

	// 应用默认AI参数
	h.applyDefaultAIParams(args)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	result, err := h.mcpClient.CallTool(ctx, "ai_file_manager", args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "File manager operation failed",
			"details":  err.Error(),
			"duration": time.Since(start).String(),
			"tool":     "ai_file_manager",
		})
		return
	}

	// 返回原始结果
	responseData := map[string]interface{}{
		"tool":        "ai_file_manager",
		"status":      "success",
		"instruction": request.Instruction,
		"result":      result.Content[0].Text,
		"duration":    time.Since(start).String(),
	}

	c.JSON(http.StatusOK, responseData)
}

// MCPDataProcessorHandler 5.3 AI智能数据处理
func (h *Handlers) MCPDataProcessorHandler(c *gin.Context) {
	start := time.Now()

	if h.mcpClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "MCP服务不可用",
			"tool":  "ai_data_processor",
		})
		return
	}

	var request struct {
		Instruction   string `json:"instruction" binding:"required"`
		InputData     string `json:"input_data" binding:"required"`
		DataType      string `json:"data_type"`
		OutputFormat  string `json:"output_format"`
		OperationMode string `json:"operation_mode"`
		Provider      string `json:"provider"`
		Model         string `json:"model"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"tool":    "ai_data_processor",
		})
		return
	}

	// 构建MCP调用参数
	args := map[string]interface{}{
		"instruction": request.Instruction + " " + h.getLanguageInstruction(),
		"input_data":  request.InputData,
	}

	if request.DataType != "" {
		args["data_type"] = request.DataType
	}
	if request.OutputFormat != "" {
		args["output_format"] = request.OutputFormat
	}
	if request.OperationMode != "" {
		args["operation_mode"] = request.OperationMode
	}
	if request.Provider != "" {
		args["provider"] = request.Provider
	}
	if request.Model != "" {
		args["model"] = request.Model
	}

	// 应用默认AI参数
	h.applyDefaultAIParams(args)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	result, err := h.mcpClient.CallTool(ctx, "ai_data_processor", args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Data processing failed",
			"details":  err.Error(),
			"duration": time.Since(start).String(),
			"tool":     "ai_data_processor",
		})
		return
	}

	// 返回原始结果
	responseData := map[string]interface{}{
		"tool":        "ai_data_processor",
		"status":      "success",
		"instruction": request.Instruction,
		"result":      result.Content[0].Text,
		"duration":    time.Since(start).String(),
	}

	c.JSON(http.StatusOK, responseData)
}

// MCPAPIClientHandler 5.4 AI智能网络请求
func (h *Handlers) MCPAPIClientHandler(c *gin.Context) {
	start := time.Now()

	if h.mcpClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "MCP服务不可用",
			"tool":  "ai_api_client",
		})
		return
	}

	var request struct {
		Instruction      string `json:"instruction" binding:"required"`
		BaseURL          string `json:"base_url"`
		AuthInfo         string `json:"auth_info"`
		RequestMode      string `json:"request_mode"`
		ResponseAnalysis bool   `json:"response_analysis"`
		Provider         string `json:"provider"`
		Model            string `json:"model"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"tool":    "ai_api_client",
		})
		return
	}

	// 构建MCP调用参数
	args := map[string]interface{}{
		"instruction": request.Instruction + " " + h.getLanguageInstruction(),
	}

	if request.BaseURL != "" {
		args["base_url"] = request.BaseURL
	}
	if request.AuthInfo != "" {
		args["auth_info"] = request.AuthInfo
	}
	if request.RequestMode != "" {
		args["request_mode"] = request.RequestMode
	}
	args["response_analysis"] = request.ResponseAnalysis
	if request.Provider != "" {
		args["provider"] = request.Provider
	}
	if request.Model != "" {
		args["model"] = request.Model
	}

	// 应用默认AI参数
	h.applyDefaultAIParams(args)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	result, err := h.mcpClient.CallTool(ctx, "ai_api_client", args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "API client operation failed",
			"details":  err.Error(),
			"duration": time.Since(start).String(),
			"tool":     "ai_api_client",
		})
		return
	}

	// 返回原始结果
	responseData := map[string]interface{}{
		"tool":        "ai_api_client",
		"status":      "success",
		"instruction": request.Instruction,
		"result":      result.Content[0].Text,
		"duration":    time.Since(start).String(),
	}

	c.JSON(http.StatusOK, responseData)
}

// MCPQueryWithAnalysisHandler 5.5 AI智能数据库查询
func (h *Handlers) MCPQueryWithAnalysisHandler(c *gin.Context) {
	start := time.Now()

	if h.mcpClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "MCP服务不可用",
			"tool":  "ai_query_with_analysis",
		})
		return
	}

	var request struct {
		Description  string `json:"description" binding:"required"`
		AnalysisType string `json:"analysis_type"`
		TableName    string `json:"table_name"`
		Context      string `json:"context"`
		InsightLevel string `json:"insight_level"`
		Provider     string `json:"provider"`
		Model        string `json:"model"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"tool":    "ai_query_with_analysis",
		})
		return
	}

	// 构建MCP调用参数
	args := map[string]interface{}{
		"description": request.Description + " " + h.getLanguageInstruction(),
	}

	if request.AnalysisType != "" {
		args["analysis_type"] = request.AnalysisType
	}
	if request.TableName != "" {
		args["table_name"] = request.TableName
	}
	if request.Context != "" {
		args["context"] = request.Context
	}
	if request.InsightLevel != "" {
		args["insight_level"] = request.InsightLevel
	}
	if request.Provider != "" {
		args["provider"] = request.Provider
	}
	if request.Model != "" {
		args["model"] = request.Model
	}

	// 应用默认AI参数
	h.applyDefaultAIParams(args)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	result, err := h.mcpClient.CallTool(ctx, "ai_query_with_analysis", args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Query with analysis failed",
			"details":  err.Error(),
			"duration": time.Since(start).String(),
			"tool":     "ai_query_with_analysis",
		})
		return
	}

	// 尝试解析MCP返回的结果
	var mcpResponse struct {
		Tool         string      `json:"tool"`
		Status       string      `json:"status"`
		Description  string      `json:"description"`
		AnalysisType string      `json:"analysis_type"`
		TableName    string      `json:"table_name"`
		Provider     string      `json:"provider"`
		Model        string      `json:"model"`
		Result       interface{} `json:"result"`
		Analysis     string      `json:"analysis"`
		Insights     string      `json:"insights"`
	}

	var responseData map[string]interface{}
	if err := json.Unmarshal([]byte(result.Content[0].Text), &mcpResponse); err == nil {
		responseData = map[string]interface{}{
			"tool":          "ai_query_with_analysis",
			"status":        mcpResponse.Status,
			"description":   request.Description,
			"analysis_type": mcpResponse.AnalysisType,
			"duration":      time.Since(start).String(),
		}

		if mcpResponse.TableName != "" {
			responseData["table_name"] = mcpResponse.TableName
		}
		if mcpResponse.Provider != "" {
			responseData["provider"] = mcpResponse.Provider
		}
		if mcpResponse.Model != "" {
			responseData["model"] = mcpResponse.Model
		}
		if mcpResponse.Result != nil {
			responseData["result"] = mcpResponse.Result
		}
		if mcpResponse.Analysis != "" {
			responseData["analysis"] = mcpResponse.Analysis
		}
		if mcpResponse.Insights != "" {
			responseData["insights"] = mcpResponse.Insights
		}
	} else {
		responseData = map[string]interface{}{
			"tool":        "ai_query_with_analysis",
			"status":      "success",
			"description": request.Description,
			"result":      result.Content[0].Text,
			"duration":    time.Since(start).String(),
		}
	}

	c.JSON(http.StatusOK, responseData)
}
