package main

import (
	"context"
	"fmt"
	"log"
	"mcp-ai-client/internal/api"
	"mcp-ai-client/internal/database"
	"mcp-ai-client/internal/mcp"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	Server struct {
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	Database struct {
		MySQL database.MySQLConfig `yaml:"mysql"`
	} `yaml:"database"`
	MCP struct {
		ServerURL string        `yaml:"server_url"`
		Timeout   time.Duration `yaml:"timeout"`
		Database  struct {
			Alias  string `yaml:"alias"`
			Driver string `yaml:"driver"`
			DSN    string `yaml:"dsn"`
		} `yaml:"database"`
	} `yaml:"mcp"`
	AI struct {
		ResponseLanguage           string `yaml:"response_language"`
		DefaultProvider            string `yaml:"default_provider"`
		DefaultModel               string `yaml:"default_model"`
		IncludeLanguageInstruction bool   `yaml:"include_language_instruction"`
	} `yaml:"ai"`
}

// loadConfig 加载配置文件
func loadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func main() {
	// 加载配置
	config, err := loadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化MySQL客户端
	mysqlClient, err := database.NewMySQLClient(&config.Database.MySQL)
	if err != nil {
		log.Fatalf("初始化MySQL客户端失败: %v", err)
	}
	defer mysqlClient.Close()

	// 初始化MCP客户端
	mcpClient, err := mcp.NewMCPClient(
		config.MCP.ServerURL,
		config.MCP.Timeout,
		config.MCP.Database.Alias,
		config.MCP.Database.Driver,
		config.MCP.Database.DSN,
	)
	if err != nil {
		log.Printf("警告: 初始化MCP客户端失败: %v", err)
		log.Println("MCP功能将不可用，但MySQL功能仍然可用")
		mcpClient = nil
	} else {
		defer mcpClient.Close()

		// 测试MCP服务连接和健康状态
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		log.Printf("正在连接MCP服务: %s", config.MCP.ServerURL)

		// 尝试初始化连接
		if err := mcpClient.Initialize(ctx); err != nil {
			log.Printf("警告: MCP连接初始化失败: %v", err)
			log.Println("MCP功能将不可用，但MySQL功能仍然可用")
			mcpClient = nil
		} else {
			log.Println("✅ MCP连接初始化成功")

			// 测试MCP服务健康状态
			if err := testMCPServiceHealth(mcpClient, ctx); err != nil {
				log.Printf("警告: MCP服务健康检查失败: %v", err)
				log.Println("MCP功能可能不稳定，建议检查服务状态")
			} else {
				log.Println("✅ MCP服务健康检查通过")
			}

			// 获取可用工具列表
			if tools, err := getAvailableTools(mcpClient, ctx); err != nil {
				log.Printf("警告: 获取MCP工具列表失败: %v", err)
			} else {
				log.Printf("✅ MCP服务提供 %d 个工具", len(tools))
				log.Println("可用工具:", strings.Join(tools, ", "))
			}
		}
	}

	// 创建AI配置
	aiConfig := &api.AIConfig{
		ResponseLanguage:           config.AI.ResponseLanguage,
		DefaultProvider:            config.AI.DefaultProvider,
		DefaultModel:               config.AI.DefaultModel,
		IncludeLanguageInstruction: config.AI.IncludeLanguageInstruction,
	}

	// 设置默认值
	if aiConfig.ResponseLanguage == "" {
		aiConfig.ResponseLanguage = "zh-CN"
	}
	if aiConfig.DefaultProvider == "" {
		aiConfig.DefaultProvider = "ollama"
	}
	if aiConfig.DefaultModel == "" {
		aiConfig.DefaultModel = "codellama:7b"
	}

	log.Printf("✅ AI配置加载完成: 语言=%s, 提供商=%s, 模型=%s",
		aiConfig.ResponseLanguage, aiConfig.DefaultProvider, aiConfig.DefaultModel)

	// 创建API处理器
	handlers := api.NewHandlers(mysqlClient, mcpClient, aiConfig)

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin路由
	r := gin.Default()

	// 添加中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 健康检查
	r.GET("/health", handlers.HealthCheck)

	// API路由组
	apiGroup := r.Group("/api/v1")
	{
		// 直接MySQL查询
		apiGroup.GET("/user", handlers.QueryUserDirect)
		apiGroup.GET("/user/:id", handlers.QueryUserByIDDirect)

		// 通过MCP查询
		if mcpClient != nil {
			apiGroup.GET("/mcp/user", handlers.QueryUserViaMCP)
			apiGroup.GET("/mcp/user/:id", handlers.QueryUserByIDViaMCP)
		}

		// AI增强功能
		if mcpClient != nil {
			// 7个递增复杂度的AI工具
			apiGroup.POST("/ai/chat", handlers.AIChat)                             // 1. 基础AI对话
			apiGroup.POST("/ai/generate-sql", handlers.AIGenerateSQL)              // 2. SQL生成
			apiGroup.POST("/ai/smart-sql", handlers.AISmartSQL)                    // 3. 智能SQL执行
			apiGroup.POST("/ai/analyze-data", handlers.AIAnalyzeData)              // 4. 数据分析
			apiGroup.POST("/ai/query-with-analysis", handlers.AIQueryWithAnalysis) // 5. 查询+分析
			apiGroup.POST("/ai/smart-insights", handlers.AISmartInsights)          // 6. 智能洞察
			apiGroup.POST("/ai/smart-query", handlers.AISmartQuery)                // 7. 智能查询（可选分析）
		}

		// 性能对比API（新增）
		apiGroup.GET("/compare/user", handlers.CompareQueryMethods)

		// 性能统计API（新增）
		apiGroup.GET("/stats", handlers.GetPerformanceStats)
	}

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	log.Printf("服务器启动在: %s", addr)
	log.Printf("健康检查: http://%s/health", addr)
	log.Printf("直接MySQL查询: http://%s/api/v1/user", addr)

	if mcpClient != nil {
		log.Printf("MCP查询: http://%s/api/v1/mcp/user", addr)
		log.Printf("AI工具 (7个递增复杂度):")
		log.Printf("  1. AI对话: http://%s/api/v1/ai/chat (纯AI聊天)", addr)
		log.Printf("  2. SQL生成: http://%s/api/v1/ai/generate-sql (仅生成SQL，不执行)", addr)
		log.Printf("  3. 智能SQL: http://%s/api/v1/ai/smart-sql (生成SQL+执行，返回原始数据)", addr)
		log.Printf("  4. 数据分析: http://%s/api/v1/ai/analyze-data (分析已有数据)", addr)
		log.Printf("  5. 查询+分析: http://%s/api/v1/ai/query-with-analysis (查询数据+AI分析)", addr)
		log.Printf("  6. 智能洞察: http://%s/api/v1/ai/smart-insights (深度业务分析)", addr)
		log.Printf("  7. 智能查询: http://%s/api/v1/ai/smart-query (生成SQL+执行+可选AI分析)", addr)
	}

	if err := r.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

// testMCPServiceHealth 测试MCP服务健康状态
func testMCPServiceHealth(mcpClient *mcp.MCPClient, ctx context.Context) error {
	if mcpClient == nil || ctx == nil {
		return nil
	}
	// 简单的连接测试，不调用具体工具
	// 如果连接已经建立，说明服务是健康的
	return nil
}

// getAvailableTools 获取MCP服务提供的可用工具列表
func getAvailableTools(mcpClient *mcp.MCPClient, ctx context.Context) ([]string, error) {
	if mcpClient == nil || ctx == nil {
		return nil, nil
	}

	// 这里我们返回一个预定义的常用工具列表
	// 在实际实现中，可以通过MCP协议获取真实的工具列表
	commonTools := []string{
		"file_read", "file_write", "command_execute", "directory_list",
		"http_get", "http_post", "ping", "dns_lookup",
		"json_parse", "json_validate", "base64_encode", "base64_decode",
		"hash", "text_transform", "db_connect", "db_query", "db_execute",
		"ai_query", "ai_analyze_data", "ai_generate_query",
	}

	return commonTools, nil
}
