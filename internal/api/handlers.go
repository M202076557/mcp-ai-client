package api

import (
	"context"
	"encoding/json"
	"mcp-ai-client/internal/database"
	"mcp-ai-client/internal/mcp"
	"net/http"
	"strconv"
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

// Handlers API处理器
type Handlers struct {
	mysqlClient *database.MySQLClient
	mcpClient   *mcp.MCPClient
	aiConfig    *AIConfig
}

// NewHandlers 创建API处理器
func NewHandlers(mysqlClient *database.MySQLClient, mcpClient *mcp.MCPClient, aiConfig *AIConfig) *Handlers {
	return &Handlers{
		mysqlClient: mysqlClient,
		mcpClient:   mcpClient,
		aiConfig:    aiConfig,
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

// enhancePromptWithLanguage 为提示词添加语言指令
func (h *Handlers) enhancePromptWithLanguage(prompt string) string {
	langInstruction := h.getLanguageInstruction()
	if langInstruction == "" {
		return prompt
	}
	return prompt + "\n\n" + langInstruction
}

// applyDefaultAIParams 应用默认AI参数
func (h *Handlers) applyDefaultAIParams(args map[string]interface{}) {
	if _, exists := args["provider"]; !exists && h.aiConfig.DefaultProvider != "" {
		args["provider"] = h.aiConfig.DefaultProvider
	}
	if _, exists := args["model"]; !exists && h.aiConfig.DefaultModel != "" {
		args["model"] = h.aiConfig.DefaultModel
	}
}

// HealthCheck 健康检查
func (h *Handlers) HealthCheck(c *gin.Context) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"services": map[string]string{
			"mysql": "connected",
			"mcp":   "connected",
		},
	}

	if h.mcpClient == nil {
		response["services"].(map[string]string)["mcp"] = "disconnected"
	}

	c.JSON(http.StatusOK, response)
}

// QueryUserDirect 直接查询MySQL mcp_user表
func (h *Handlers) QueryUserDirect(c *gin.Context) {
	start := time.Now()

	userData, err := h.mysqlClient.QueryUser()
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":     err.Error(),
			"method":    "direct_mysql",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	responseTime := time.Since(start)
	c.JSON(http.StatusOK, map[string]interface{}{
		"data":          userData,
		"method":        "direct_mysql",
		"response_time": responseTime.String(),
		"timestamp":     time.Now().Format(time.RFC3339),
	})
}

// QueryUserByIDDirect 直接根据ID查询MySQL mcp_user表
func (h *Handlers) QueryUserByIDDirect(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":     "无效的ID参数",
			"method":    "direct_mysql",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	start := time.Now()
	userData, err := h.mysqlClient.QueryUserByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":     err.Error(),
			"method":    "direct_mysql",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	responseTime := time.Since(start)
	c.JSON(http.StatusOK, map[string]interface{}{
		"data":          userData,
		"method":        "direct_mysql",
		"response_time": responseTime.String(),
		"timestamp":     time.Now().Format(time.RFC3339),
	})
}

// QueryUserViaMCP 通过MCP查询mcp_user表
func (h *Handlers) QueryUserViaMCP(c *gin.Context) {
	if h.mcpClient == nil {
		c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"error":     "MCP服务不可用",
			"method":    "mcp_service",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := h.mcpClient.QueryUserViaMCP(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":     err.Error(),
			"method":    "mcp_service",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	responseTime := time.Since(start)
	c.JSON(http.StatusOK, map[string]interface{}{
		"data":          result,
		"method":        "mcp_service",
		"response_time": responseTime.String(),
		"timestamp":     time.Now().Format(time.RFC3339),
	})
}

// QueryUserByIDViaMCP 通过MCP根据ID查询mcp_user表
func (h *Handlers) QueryUserByIDViaMCP(c *gin.Context) {
	if h.mcpClient == nil {
		c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"error":     "MCP服务不可用",
			"method":    "mcp_service",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":     "无效的ID参数",
			"method":    "mcp_service",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := h.mcpClient.QueryUserByIDViaMCP(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":     err.Error(),
			"method":    "mcp_service",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	responseTime := time.Since(start)
	c.JSON(http.StatusOK, map[string]interface{}{
		"data":          result,
		"method":        "mcp_service",
		"response_time": responseTime.String(),
		"timestamp":     time.Now().Format(time.RFC3339),
	})
}

// AIGenerateSQL 通过AI生成SQL查询
func (h *Handlers) AIGenerateSQL(c *gin.Context) {
	if h.mcpClient == nil {
		c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"error":     "MCP服务不可用",
			"method":    "ai_generate_sql",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	// 获取请求参数
	var request struct {
		Description string `json:"description" binding:"required"`
		TableSchema string `json:"table_schema"`
		TableName   string `json:"table_name"`
		Model       string `json:"model"`
		Provider    string `json:"provider"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":     "无效的请求参数",
			"method":    "ai_generate_sql",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 调用MCP的AI工具生成SQL
	arguments := map[string]interface{}{
		"description": request.Description,
	}
	if request.Model != "" {
		arguments["model"] = request.Model
	}
	if request.TableSchema != "" {
		arguments["table_schema"] = request.TableSchema
	}
	if request.TableName != "" {
		arguments["table_name"] = request.TableName
	}
	if request.Provider != "" {
		arguments["provider"] = request.Provider
	}

	// 应用默认AI参数
	h.applyDefaultAIParams(arguments)

	result, err := h.mcpClient.CallTool(ctx, "ai_generate_sql", arguments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":     err.Error(),
			"method":    "ai_generate_sql",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	responseTime := time.Since(start)

	// 解析MCP返回的结果
	var mcpResponse struct {
		Tool         string `json:"tool"`
		Status       string `json:"status"`
		Description  string `json:"description"`
		TableName    string `json:"table_name"`
		GeneratedSQL string `json:"generated_sql"`
		Provider     string `json:"provider"`
		Model        string `json:"model"`
	}

	// 尝试解析MCP返回的JSON
	var sqlStatement string
	var provider string
	var model string

	if err := json.Unmarshal([]byte(result.Content[0].Text), &mcpResponse); err == nil {
		sqlStatement = mcpResponse.GeneratedSQL
		provider = mcpResponse.Provider
		model = mcpResponse.Model
	} else {
		// 如果解析失败，直接使用原始文本
		sqlStatement = result.Content[0].Text
		provider = request.Provider
		model = request.Model
	}

	// 构建友好的响应格式
	response := map[string]interface{}{
		"tool":          "ai_generate_sql",
		"status":        "success",
		"description":   request.Description,
		"generated_sql": sqlStatement,
		"response_time": responseTime.String(),
		"timestamp":     time.Now().Format(time.RFC3339),
	}

	// 添加可选字段
	if request.TableName != "" {
		response["table_name"] = request.TableName
	}
	if provider != "" {
		response["provider"] = provider
	}
	if model != "" {
		response["model"] = model
	}

	// AIGenerateSQL工具只负责生成SQL，不执行数据库查询
	// 如果需要执行SQL，应该调用 ai_smart_sql 工具
	response["execution"] = map[string]interface{}{
		"success": false,
		"message": "SQL generation only. Use ai_smart_sql tool to execute the generated SQL.",
	}

	c.JSON(http.StatusOK, response)
}

// CompareQueryMethods 对比三种查询方式的性能
func (h *Handlers) CompareQueryMethods(c *gin.Context) {
	// 创建结果通道
	directChan := make(chan MethodResult, 1)
	mcpChan := make(chan MethodResult, 1)
	aiChan := make(chan MethodResult, 1)

	// 并行执行三种查询方式
	go func() {
		start := time.Now()
		userData, err := h.mysqlClient.QueryUser()
		responseTime := time.Since(start)

		result := MethodResult{
			Method:       "direct_mysql",
			Success:      err == nil,
			ResponseTime: responseTime,
			Error:        "",
			DataCount:    0,
		}

		if err == nil {
			result.DataCount = len(userData)
		} else {
			result.Error = err.Error()
		}

		directChan <- result
	}()

	// MCP查询（如果可用）
	if h.mcpClient != nil {
		go func() {
			start := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			result, err := h.mcpClient.QueryUserViaMCP(ctx)
			responseTime := time.Since(start)

			mcpResult := MethodResult{
				Method:       "mcp_service",
				Success:      err == nil,
				ResponseTime: responseTime,
				Error:        "",
				DataCount:    0,
			}

			if err == nil {
				// 简单估算数据量（基于返回字符串长度）
				mcpResult.DataCount = len(result) / 100 // 粗略估算
			} else {
				mcpResult.Error = err.Error()
			}

			mcpChan <- mcpResult
		}()
	} else {
		// MCP不可用，发送错误结果
		mcpChan <- MethodResult{
			Method:       "mcp_service",
			Success:      false,
			ResponseTime: 0,
			Error:        "MCP服务不可用",
			DataCount:    0,
		}
	}

	// AI查询（如果可用）
	if h.mcpClient != nil {
		go func() {
			start := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			// 使用AI智能SQL查询（通过MCP协议）
			arguments := map[string]interface{}{
				"prompt": "查询所有用户记录",
				"model":  "codellama:7b",
			}

			result, err := h.mcpClient.CallTool(ctx, "ai_smart_sql", arguments)
			responseTime := time.Since(start)

			aiResult := MethodResult{
				Method:       "ai_enhanced",
				Success:      err == nil,
				ResponseTime: responseTime,
				Error:        "",
				DataCount:    0,
			}

			if err == nil {
				// 尝试解析AI查询结果中的数据计数
				if len(result.Content) > 0 {
					var aiResponse map[string]interface{}
					if json.Unmarshal([]byte(result.Content[0].Text), &aiResponse) == nil {
						if rows, ok := aiResponse["rows"].([]interface{}); ok {
							aiResult.DataCount = len(rows)
						}
					}
				}
			} else {
				aiResult.Error = err.Error()
			}

			aiChan <- aiResult
		}()
	} else {
		// AI不可用，发送错误结果
		aiChan <- MethodResult{
			Method:       "ai_enhanced",
			Success:      false,
			ResponseTime: 0,
			Error:        "AI服务不可用",
			DataCount:    0,
		}
	}

	// 收集所有结果
	directResult := <-directChan
	mcpResult := <-mcpChan
	aiResult := <-aiChan

	// 分析结果
	analysis := analyzeComparisonResults(directResult, mcpResult, aiResult)

	// 构建响应
	response := ComparisonResult{
		Timestamp: time.Now().Format(time.RFC3339),
		Methods:   []MethodResult{directResult, mcpResult, aiResult},
		Analysis:  analysis,
	}

	c.JSON(http.StatusOK, response)
}

// MethodResult 单个查询方法的结果
type MethodResult struct {
	Method       string        `json:"method"`
	Success      bool          `json:"success"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
	DataCount    int           `json:"data_count"`
}

// ComparisonResult 对比结果
type ComparisonResult struct {
	Timestamp string         `json:"timestamp"`
	Methods   []MethodResult `json:"methods"`
	Analysis  AnalysisResult `json:"analysis"`
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	FastestMethod   string  `json:"fastest_method"`
	SlowestMethod   string  `json:"slowest_method"`
	MostReliable    string  `json:"most_reliable"`
	PerformanceGap  string  `json:"performance_gap"`
	Recommendation  string  `json:"recommendation"`
	EfficiencyScore float64 `json:"efficiency_score"`
}

// analyzeComparisonResults 分析对比结果
func analyzeComparisonResults(direct, mcp, ai MethodResult) AnalysisResult {
	// 找出最快和最慢的方法
	methods := []MethodResult{direct, mcp, ai}
	var fastest, slowest MethodResult
	var fastestTime, slowestTime time.Duration

	for _, method := range methods {
		if method.Success {
			if fastestTime == 0 || method.ResponseTime < fastestTime {
				fastest = method
				fastestTime = method.ResponseTime
			}
			if method.ResponseTime > slowestTime {
				slowest = method
				slowestTime = method.ResponseTime
			}
		}
	}

	// 找出最可靠的方法（成功率最高）
	var mostReliable string
	successCount := map[string]int{
		"direct_mysql": 0,
		"mcp_service":  0,
		"ai_enhanced":  0,
	}

	if direct.Success {
		successCount["direct_mysql"]++
	}
	if mcp.Success {
		successCount["mcp_service"]++
	}
	if ai.Success {
		successCount["ai_enhanced"]++
	}

	for method, count := range successCount {
		if count > 0 {
			mostReliable = method
			break
		}
	}

	// 计算性能差距
	performanceGap := ""
	if fastestTime > 0 && slowestTime > 0 {
		gap := slowestTime - fastestTime
		performanceGap = gap.String()
	}

	// 生成推荐
	recommendation := ""
	if fastest.Method == "direct_mysql" {
		recommendation = "对于简单查询，直接数据库访问性能最佳"
	} else if fastest.Method == "mcp_service" {
		recommendation = "MCP服务在复杂查询场景下表现良好"
	} else if fastest.Method == "ai_enhanced" {
		recommendation = "AI增强查询在自然语言交互场景下价值显著"
	}

	// 计算效率分数（基于响应时间和成功率）
	var efficiencyScore float64
	totalMethods := 3.0
	for _, method := range methods {
		if method.Success {
			// 成功的方法得分更高
			efficiencyScore += 0.5
			// 响应时间越短得分越高
			if method.ResponseTime < 10*time.Millisecond {
				efficiencyScore += 0.3
			} else if method.ResponseTime < 100*time.Millisecond {
				efficiencyScore += 0.2
			} else if method.ResponseTime < 1*time.Second {
				efficiencyScore += 0.1
			}
		}
	}
	efficiencyScore = efficiencyScore / totalMethods

	return AnalysisResult{
		FastestMethod:   fastest.Method,
		SlowestMethod:   slowest.Method,
		MostReliable:    mostReliable,
		PerformanceGap:  performanceGap,
		Recommendation:  recommendation,
		EfficiencyScore: efficiencyScore,
	}
}

// GetPerformanceStats 获取性能统计信息
func (h *Handlers) GetPerformanceStats(c *gin.Context) {
	// 这里可以实现更复杂的统计逻辑
	// 目前返回基础的系统状态信息

	// 检查MySQL连接状态
	mysqlStatus := "connected"
	if _, err := h.mysqlClient.GetUserCount(); err != nil {
		mysqlStatus = "disconnected"
	}

	// 检查MCP连接状态
	mcpStatus := "disconnected"
	if h.mcpClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if _, err := h.mcpClient.CallTool(ctx, "ai_query", map[string]interface{}{
			"prompt": "test",
		}); err == nil {
			mcpStatus = "connected"
		}
	}

	// 获取数据库基本信息
	var userCount int
	var schemaInfo []map[string]interface{}

	if count, err := h.mysqlClient.GetUserCount(); err == nil {
		userCount = count
	}

	if schema, err := h.mysqlClient.GetUserSchema(); err == nil {
		schemaInfo = schema
	}

	response := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"system_status": map[string]string{
			"mysql": mysqlStatus,
			"mcp":   mcpStatus,
		},
		"database_info": map[string]interface{}{
			"user_count":   userCount,
			"table_schema": schemaInfo,
		},
		"performance_metrics": map[string]interface{}{
			"total_apis": 6, // 包括新增的对比API
			"available_methods": []string{
				"direct_mysql",
				"mcp_service",
				"ai_enhanced",
				"performance_comparison",
			},
		},
	}

	c.JSON(http.StatusOK, response)
}

// 新增AI工具处理器

// AIChat 基础AI对话
func (h *Handlers) AIChat(c *gin.Context) {
	var request struct {
		Prompt      string  `json:"prompt" binding:"required"`
		Provider    string  `json:"provider,omitempty"`
		Model       string  `json:"model,omitempty"`
		MaxTokens   int     `json:"max_tokens,omitempty"`
		Temperature float64 `json:"temperature,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	start := time.Now()

	// 为提示词添加语言指令
	enhancedPrompt := h.enhancePromptWithLanguage(request.Prompt)

	args := map[string]interface{}{
		"prompt": enhancedPrompt,
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := h.mcpClient.CallTool(ctx, "ai_chat", args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "AI Chat failed",
			"details":  err.Error(),
			"duration": time.Since(start).String(),
			"tool":     "ai_chat",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tool":     "ai_chat",
		"result":   result,
		"duration": time.Since(start).String(),
	})
}

// AISmartSQL 智能SQL执行（推荐使用，支持自然语言和直接SQL）
func (h *Handlers) AISmartSQL(c *gin.Context) {
	var request struct {
		Prompt string `json:"prompt,omitempty"` // 自然语言查询
		SQL    string `json:"sql,omitempty"`    // 直接SQL执行
		Alias  string `json:"alias,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 必须提供prompt或sql其中之一
	if request.Prompt == "" && request.SQL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "必须提供prompt（自然语言查询）或sql（直接SQL）参数"})
		return
	}

	start := time.Now()
	args := make(map[string]interface{})

	if request.Prompt != "" {
		args["prompt"] = request.Prompt
	}
	if request.SQL != "" {
		args["sql"] = request.SQL
	}
	if request.Alias != "" {
		args["alias"] = request.Alias
	}
	if request.Limit > 0 {
		args["limit"] = request.Limit
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 重新构建统一的参数，使用prompt字段让服务端自动检测类型
	args = make(map[string]interface{})

	if request.Prompt != "" {
		args["prompt"] = request.Prompt
	} else {
		args["prompt"] = request.SQL // SQL也通过prompt传递，让服务端自动检测
	}

	args["analysis_mode"] = "fast" // 默认快速模式

	if request.Alias != "" {
		args["alias"] = request.Alias
	}
	if request.Limit > 0 {
		args["limit"] = request.Limit
	}

	// 调用统一的智能查询工具
	result, err := h.mcpClient.CallTool(ctx, "ai_smart_query", args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Smart SQL execution failed",
			"details":  err.Error(),
			"duration": time.Since(start).String(),
			"tool":     "ai_smart_sql",
		})
		return
	}

	// 解析MCP返回的结果
	var mcpResponse struct {
		Tool     string `json:"tool"`
		Status   string `json:"status"`
		SQL      string `json:"sql"`
		Prompt   string `json:"prompt,omitempty"`
		Alias    string `json:"alias,omitempty"`
		Limit    int    `json:"limit,omitempty"`
		RowCount int    `json:"row_count"`
		AIMode   bool   `json:"ai_mode"`
		Provider string `json:"provider,omitempty"`
		Model    string `json:"model,omitempty"`
		Result   struct {
			Columns  []string                 `json:"columns"`
			Rows     []map[string]interface{} `json:"rows"`
			Limited  bool                     `json:"limited"`
			SQLQuery string                   `json:"sql_query"`
		} `json:"result"`
	}

	// 尝试解析MCP返回的JSON
	var responseData map[string]interface{}
	if err := json.Unmarshal([]byte(result.Content[0].Text), &mcpResponse); err == nil {
		// 构建友好的响应格式
		responseData = map[string]interface{}{
			"tool":      "ai_smart_sql",
			"status":    mcpResponse.Status,
			"sql":       mcpResponse.SQL,
			"ai_mode":   mcpResponse.AIMode,
			"row_count": mcpResponse.RowCount,
			"duration":  time.Since(start).String(),
		}

		if mcpResponse.Prompt != "" {
			responseData["prompt"] = mcpResponse.Prompt
		}
		if mcpResponse.Alias != "" {
			responseData["alias"] = mcpResponse.Alias
		}
		if mcpResponse.Limit > 0 {
			responseData["limit"] = mcpResponse.Limit
		}
		if mcpResponse.Provider != "" {
			responseData["provider"] = mcpResponse.Provider
			responseData["model"] = mcpResponse.Model
		}
		if mcpResponse.Result.Columns != nil {
			responseData["columns"] = mcpResponse.Result.Columns
			responseData["rows"] = mcpResponse.Result.Rows
			responseData["limited"] = mcpResponse.Result.Limited
		}
	} else {
		// 如果解析失败，返回原始结果
		responseData = map[string]interface{}{
			"tool":     "ai_smart_sql",
			"status":   "success",
			"result":   result.Content[0].Text,
			"duration": time.Since(start).String(),
		}
	}

	c.JSON(http.StatusOK, responseData)
}

// AIAnalyzeData 数据分析
func (h *Handlers) AIAnalyzeData(c *gin.Context) {
	var request struct {
		Data         interface{} `json:"data" binding:"required"`
		AnalysisType string      `json:"analysis_type,omitempty"`
		Context      string      `json:"context,omitempty"`
		Provider     string      `json:"provider,omitempty"`
		Model        string      `json:"model,omitempty"`
		Focus        []string    `json:"focus,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	start := time.Now()

	// 将数据转换为JSON字符串，因为MCP服务端期望字符串类型
	dataBytes, err := json.Marshal(request.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid data format",
			"details": err.Error(),
		})
		return
	}

	args := map[string]interface{}{
		"data": string(dataBytes),
	}
	if request.AnalysisType != "" {
		args["analysis_type"] = request.AnalysisType
	}
	if request.Context != "" {
		args["context"] = request.Context
	}
	if request.Provider != "" {
		args["provider"] = request.Provider
	}
	if request.Model != "" {
		args["model"] = request.Model
	}
	if len(request.Focus) > 0 {
		args["focus"] = request.Focus
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := h.mcpClient.CallTool(ctx, "ai_analyze_data", args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Data analysis failed",
			"details":  err.Error(),
			"duration": time.Since(start).String(),
			"tool":     "ai_analyze_data",
		})
		return
	}

	// 解析MCP返回的结果
	var mcpResponse struct {
		Tool         string `json:"tool"`
		Status       string `json:"status"`
		AnalysisType string `json:"analysis_type"`
		Provider     string `json:"provider"`
		Model        string `json:"model"`
		Analysis     string `json:"analysis"`
	}

	// 尝试解析MCP返回的JSON
	var responseData map[string]interface{}
	if err := json.Unmarshal([]byte(result.Content[0].Text), &mcpResponse); err == nil {
		// 构建友好的响应格式
		responseData = map[string]interface{}{
			"tool":          "ai_analyze_data",
			"status":        mcpResponse.Status,
			"analysis_type": mcpResponse.AnalysisType,
			"analysis":      mcpResponse.Analysis,
			"duration":      time.Since(start).String(),
		}

		if mcpResponse.Provider != "" {
			responseData["provider"] = mcpResponse.Provider
		}
		if mcpResponse.Model != "" {
			responseData["model"] = mcpResponse.Model
		}
	} else {
		// 如果解析失败，返回原始结果
		responseData = map[string]interface{}{
			"tool":     "ai_analyze_data",
			"status":   "success",
			"result":   result.Content[0].Text,
			"duration": time.Since(start).String(),
		}
	}

	c.JSON(http.StatusOK, responseData)
}

// AIQueryWithAnalysis 数据查询+分析
func (h *Handlers) AIQueryWithAnalysis(c *gin.Context) {
	var request struct {
		Description  string `json:"description" binding:"required"`
		AnalysisType string `json:"analysis_type,omitempty"`
		TableName    string `json:"table_name,omitempty"`
		Provider     string `json:"provider,omitempty"`
		Model        string `json:"model,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	start := time.Now()
	args := map[string]interface{}{
		"description": request.Description,
	}
	if request.AnalysisType != "" {
		args["analysis_type"] = request.AnalysisType
	}
	if request.TableName != "" {
		args["table_name"] = request.TableName
	}
	if request.Provider != "" {
		args["provider"] = request.Provider
	}
	if request.Model != "" {
		args["model"] = request.Model
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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

	// 解析MCP返回的结果
	var mcpResponse struct {
		Tool         string                 `json:"tool"`
		Status       string                 `json:"status"`
		Description  string                 `json:"description"`
		AnalysisType string                 `json:"analysis_type"`
		QueryResult  interface{}            `json:"query_result"`
		Analysis     map[string]interface{} `json:"analysis"`
	}

	// 尝试解析MCP返回的JSON
	var responseData map[string]interface{}
	if err := json.Unmarshal([]byte(result.Content[0].Text), &mcpResponse); err == nil {
		// 构建友好的响应格式
		responseData = map[string]interface{}{
			"tool":          "ai_query_with_analysis",
			"status":        mcpResponse.Status,
			"description":   mcpResponse.Description,
			"analysis_type": mcpResponse.AnalysisType,
			"duration":      time.Since(start).String(),
		}

		// 处理 analysis 字段，只保留分析文本内容
		if mcpResponse.Analysis != nil {
			// 提取 provider 和 model 到顶层
			if provider, ok := mcpResponse.Analysis["provider"].(string); ok {
				responseData["provider"] = provider
			}
			if model, ok := mcpResponse.Analysis["model"].(string); ok {
				responseData["model"] = model
			}

			// 只保留分析文本内容
			if analysisText, ok := mcpResponse.Analysis["analysis"].(string); ok {
				responseData["analysis"] = analysisText
			}
		}
	} else {
		// 如果解析失败，返回原始结果
		responseData = map[string]interface{}{
			"tool":     "ai_query_with_analysis",
			"status":   "success",
			"result":   result.Content[0].Text,
			"duration": time.Since(start).String(),
		}
	}

	c.JSON(http.StatusOK, responseData)
}

// AISmartInsights 智能洞察 🆕
func (h *Handlers) AISmartInsights(c *gin.Context) {
	var request struct {
		Prompt       string `json:"prompt" binding:"required"`
		Context      string `json:"context,omitempty"`
		InsightLevel string `json:"insight_level,omitempty"`
		TableName    string `json:"table_name,omitempty"`
		Provider     string `json:"provider,omitempty"`
		Model        string `json:"model,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	start := time.Now()

	// 为提示词添加语言指令
	enhancedPrompt := h.enhancePromptWithLanguage(request.Prompt)

	args := map[string]interface{}{
		"prompt": enhancedPrompt,
	}
	if request.Context != "" {
		args["context"] = request.Context
	}
	if request.InsightLevel != "" {
		args["insight_level"] = request.InsightLevel
	}
	if request.TableName != "" {
		args["table_name"] = request.TableName
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

	result, err := h.mcpClient.CallTool(ctx, "ai_smart_insights", args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Smart insights failed",
			"details":  err.Error(),
			"duration": time.Since(start).String(),
			"tool":     "ai_smart_insights",
		})
		return
	}

	// 解析MCP返回的结果
	var mcpResponse struct {
		Tool         string `json:"tool"`
		Status       string `json:"status"`
		Prompt       string `json:"prompt"`
		InsightLevel string `json:"insight_level"`
		Provider     string `json:"provider"`
		Model        string `json:"model"`
		Insights     string `json:"insights"`
		Analysis     string `json:"analysis"`
	}

	// 尝试解析MCP返回的JSON
	var responseData map[string]interface{}
	if err := json.Unmarshal([]byte(result.Content[0].Text), &mcpResponse); err == nil {
		// 构建友好的响应格式
		responseData = map[string]interface{}{
			"tool":          "ai_smart_insights",
			"status":        mcpResponse.Status,
			"prompt":        mcpResponse.Prompt,
			"insight_level": mcpResponse.InsightLevel,
			"duration":      time.Since(start).String(),
		}

		// 优先使用insights字段，如果没有则使用analysis字段
		if mcpResponse.Insights != "" {
			responseData["insights"] = mcpResponse.Insights
		} else if mcpResponse.Analysis != "" {
			responseData["insights"] = mcpResponse.Analysis
		}

		if mcpResponse.Provider != "" {
			responseData["provider"] = mcpResponse.Provider
		}
		if mcpResponse.Model != "" {
			responseData["model"] = mcpResponse.Model
		}
	} else {
		// 如果解析失败，返回原始结果
		responseData = map[string]interface{}{
			"tool":     "ai_smart_insights",
			"status":   "success",
			"result":   result.Content[0].Text,
			"duration": time.Since(start).String(),
		}
	}

	c.JSON(http.StatusOK, responseData)
}

// AISmartQuery 智能查询（综合功能：生成SQL + 执行 + 可选分析）
func (h *Handlers) AISmartQuery(c *gin.Context) {
	start := time.Now()

	var req struct {
		Prompt          string `json:"prompt"`           // 新字段，优先使用
		Description     string `json:"description"`      // 兼容旧字段
		AnalysisMode    string `json:"analysis_mode"`    // "full" 或 "fast"
		IncludeAnalysis bool   `json:"include_analysis"` // 兼容旧字段
		TableName       string `json:"table_name"`
		Alias           string `json:"alias"`
		Limit           int    `json:"limit"`
		Provider        string `json:"provider"`
		Model           string `json:"model"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"tool":    "ai_smart_query",
		})
		return
	}

	// 处理兼容性：优先使用prompt，如果没有则使用description
	var prompt string
	if req.Prompt != "" {
		prompt = req.Prompt
	} else if req.Description != "" {
		prompt = req.Description
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "必须提供prompt或description参数",
			"tool":  "ai_smart_query",
		})
		return
	}

	// 设置默认值
	if req.AnalysisMode == "" {
		if req.IncludeAnalysis {
			req.AnalysisMode = "full"
		} else {
			req.AnalysisMode = "fast"
		}
	}
	if req.TableName == "" {
		req.TableName = "mcp_user"
	}
	if req.Limit == 0 {
		req.Limit = 100
	}
	if req.Provider == "" {
		req.Provider = h.aiConfig.DefaultProvider
	}
	if req.Model == "" {
		req.Model = h.aiConfig.DefaultModel
	}

	// 不在智能查询中添加语言指令，让服务端自动检测SQL类型
	// 语言指令会干扰SQL自动检测功能

	// 构建MCP调用参数
	args := map[string]interface{}{
		"prompt":        prompt,
		"analysis_mode": req.AnalysisMode,
		"table_name":    req.TableName,
		"limit":         req.Limit,
		"provider":      req.Provider,
		"model":         req.Model,
	}

	if req.Alias != "" {
		args["alias"] = req.Alias
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 调用MCP的ai_smart_query工具
	result, err := h.mcpClient.CallTool(ctx, "ai_smart_query", args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Smart query execution failed",
			"details":  err.Error(),
			"tool":     "ai_smart_query",
			"duration": time.Since(start).String(),
		})
		return
	}

	// 解析响应
	type smartQueryResponse struct {
		Tool         string      `json:"tool"`
		Status       string      `json:"status"`
		InputType    string      `json:"input_type"`
		Prompt       string      `json:"prompt"`
		SQL          string      `json:"sql"`
		AnalysisMode string      `json:"analysis_mode"`
		Limit        int         `json:"limit"`
		RowCount     interface{} `json:"row_count"`
		TableName    string      `json:"table_name"`
		Alias        string      `json:"alias"`
		Result       interface{} `json:"result"`
		RawResult    interface{} `json:"raw_result"`
		Columns      interface{} `json:"columns"`
		Rows         interface{} `json:"rows"`
		Limited      interface{} `json:"limited"`
		AIAnalysis   string      `json:"ai_analysis"`
		Error        string      `json:"error"`
	}

	var responseData map[string]interface{}
	var mcpResponse smartQueryResponse

	// 尝试解析MCP返回的JSON
	if err := json.Unmarshal([]byte(result.Content[0].Text), &mcpResponse); err == nil {
		// 构建友好的响应格式
		responseData = map[string]interface{}{
			"tool":          "ai_smart_query",
			"status":        mcpResponse.Status,
			"input_type":    mcpResponse.InputType,
			"prompt":        mcpResponse.Prompt,
			"sql":           mcpResponse.SQL,
			"analysis_mode": mcpResponse.AnalysisMode,
			"limit":         mcpResponse.Limit,
			"row_count":     mcpResponse.RowCount,
			"duration":      time.Since(start).String(),
		}

		// 添加表名（如果是自然语言查询）
		if mcpResponse.TableName != "" {
			responseData["table_name"] = mcpResponse.TableName
		}

		// 添加别名（如果有）
		if mcpResponse.Alias != "" {
			responseData["alias"] = mcpResponse.Alias
		}

		// 添加数据库查询结果 - 这是关键部分
		if mcpResponse.Result != nil {
			responseData["result"] = mcpResponse.Result
		}
		if mcpResponse.RawResult != nil {
			responseData["raw_result"] = mcpResponse.RawResult
		}
		if mcpResponse.Columns != nil {
			responseData["columns"] = mcpResponse.Columns
		}
		if mcpResponse.Rows != nil {
			responseData["rows"] = mcpResponse.Rows
		}
		if mcpResponse.Limited != nil {
			responseData["limited"] = mcpResponse.Limited
		}

		// 添加AI分析结果（如果有）
		if mcpResponse.AIAnalysis != "" {
			responseData["ai_analysis"] = mcpResponse.AIAnalysis
		}

		// 添加错误信息（如果有）
		if mcpResponse.Error != "" {
			responseData["error"] = mcpResponse.Error
			responseData["status"] = "error"
		}
	} else {
		// 如果解析失败，返回原始结果
		responseData = map[string]interface{}{
			"tool":     "ai_smart_query",
			"status":   "success",
			"result":   result.Content[0].Text,
			"duration": time.Since(start).String(),
		}
	}

	c.JSON(http.StatusOK, responseData)
}
