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
		MySQL  database.MySQLConfig `yaml:"mysql"`
		Tables struct {
			UserTable string `yaml:"user_table"`
		} `yaml:"tables"`
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
	// 默认运行统一服务模式 - 既支持传统API又支持MCP协议
	runUnifiedServerMode()
}

// printUsage 显示使用说明
func printUsage() {
	fmt.Println("MCP AI Client - 统一的MCP客户端")
	fmt.Println()
	fmt.Println("HTTP服务器模式 (默认):")
	fmt.Println("  ./bin/mcp-ai-client                    # 启动HTTP API服务器")
	fmt.Println()
	fmt.Println("AI工具演示模式:")
	fmt.Println("  ./bin/mcp-ai-client demo               # 运行所有AI工具演示")
	fmt.Println("  ./bin/mcp-ai-client chat               # AI对话演示")
	fmt.Println("  ./bin/mcp-ai-client file               # AI文件管理演示")
	fmt.Println("  ./bin/mcp-ai-client data               # AI数据处理演示")
	fmt.Println("  ./bin/mcp-ai-client api                # AI网络请求演示")
	fmt.Println("  ./bin/mcp-ai-client db                 # AI数据库查询演示")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  ./bin/mcp-ai-client                    # 启动HTTP服务器")
	fmt.Println("  ./bin/mcp-ai-client demo               # 运行AI演示")
}

// runAIClientMode 运行AI客户端演示模式 (保留用于向后兼容)
func runAIClientMode(command string) {
	log.Printf("⚠️  AI演示模式已集成到统一服务中，请使用: ./bin/mcp-ai-client")
	log.Printf("然后访问: http://localhost:8080/demo/ 查看演示")
}

// runUnifiedServerMode 运行统一服务模式 - 同时支持传统API和MCP协议
func runUnifiedServerMode() {
	log.Println("🚀 启动统一服务 - 支持传统API + MCP协议")

	// 加载配置
	config, err := loadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 1. 初始化MySQL客户端 (传统数据库服务)
	log.Println("🔗 初始化MySQL数据库连接...")
	mysqlClient, err := database.NewMySQLClient(&config.Database.MySQL)
	if err != nil {
		log.Fatalf("初始化MySQL客户端失败: %v", err)
	}
	defer mysqlClient.Close()
	log.Println("✅ MySQL连接成功")

	// 2. 初始化MCP客户端 (AI增强服务)
	log.Println("🤖 初始化MCP AI客户端...")
	var mcpClient *mcp.MCPClient
	mcpClient, err = mcp.NewMCPClient(config.MCP.ServerURL, config.MCP.Timeout)
	if err != nil {
		log.Printf("⚠️  MCP客户端初始化失败: %v", err)
		log.Println("📍 继续启动服务，但MCP增强功能将不可用")
		mcpClient = nil
	} else {
		defer mcpClient.Close()

		// 测试MCP连接
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := mcpClient.Initialize(ctx); err != nil {
			log.Printf("⚠️  MCP连接失败: %v", err)
			log.Println("📍 继续启动服务，但MCP增强功能将不可用")
			mcpClient = nil
		} else {
			log.Println("✅ MCP连接成功")
		}
	}

	// 3. 创建服务层
	log.Println("⚙️  初始化服务层...")

	// MCP服务状态检查
	var mcpServiceAvailable bool
	if mcpClient != nil {
		mcpServiceAvailable = true
		log.Println("✅ MCP服务已就绪")
	} else {
		mcpServiceAvailable = false
		log.Println("⚠️  MCP服务不可用")
	}

	// 4. 创建AI配置
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
		aiConfig.DefaultModel = "llama2:7b"
	}

	log.Printf("✅ AI配置: 语言=%s, 提供商=%s, 模型=%s",
		aiConfig.ResponseLanguage, aiConfig.DefaultProvider, aiConfig.DefaultModel)

	// 5. 创建统一API处理器
	log.Println("🌐 初始化API处理器...")

	// 配置数据库相关设置
	dbConfig := &api.DatabaseConfig{
		UserTable: config.Database.Tables.UserTable,
	}
	if dbConfig.UserTable == "" {
		dbConfig.UserTable = "mcp_user" // 默认表名
	}

	handlers := api.NewHandlers(mysqlClient, mcpClient, aiConfig, dbConfig)

	if mcpServiceAvailable {
		log.Println("✅ 统一API处理器已就绪 (传统 + MCP)")
	} else {
		log.Println("⚠️  统一API处理器功能受限（仅传统功能）")
	}

	// 6. 设置HTTP服务器
	log.Println("🌍 配置HTTP服务器...")
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 添加中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// ===== 根路径 - 服务概览 =====
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service":     "MCP AI Client - 统一服务",
			"description": "同时支持传统HTTP API和MCP AI增强协议",
			"version":     "1.0.0",
			"capabilities": gin.H{
				"traditional_database": true,
				"mcp_ai_enhanced":      mcpClient != nil,
				"unified_services":     true,
			},
			"api_groups": gin.H{
				"health":       "/health",
				"traditional":  "/api/v1/traditional/*",
				"mcp_enhanced": "/api/v1/mcp/*",
				"comparison":   "/api/v1/comparison/*",
				"demo":         "/demo/*",
			},
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// 健康检查
	r.GET("/health", handlers.HealthCheck)

	// ===== API路由设置 =====
	// 传统API路由
	traditionalV1 := r.Group("/api/v1/traditional")
	{
		traditionalV1.GET("/users", handlers.GetUsersTraditional)
		traditionalV1.GET("/users/:id", handlers.GetUserByIDTraditional)
		traditionalV1.GET("/search/users", handlers.SearchUsersTraditional)
		traditionalV1.GET("/stats/users", handlers.GetUserStatsTraditional)
	}

	// MCP增强API路由
	if mcpServiceAvailable {
		mcpV1 := r.Group("/api/v1/mcp")
		{
			mcpV1.POST("/chat", handlers.MCPChatHandler)
			mcpV1.GET("/analyze", handlers.MCPAnalyzeHandler)
			mcpV1.POST("/query", handlers.MCPQueryHandler)
		}
		log.Println("✅ MCP增强API路由已配置")
	}

	// 比较和能力展示API
	comparisonV1 := r.Group("/api/v1/comparison")
	{
		comparisonV1.GET("/services", handlers.CompareServicesHandler)
		comparisonV1.GET("/capabilities", handlers.GetServiceCapabilitiesHandler)
	}

	// 向后兼容API
	legacyV1 := r.Group("/api/v1")
	{
		legacyV1.GET("/user", handlers.QueryUserDirect)
		legacyV1.GET("/query", handlers.AIGenerateSQL)
	}

	log.Println("✅ 所有API路由已配置")

	// 7. 启动服务器
	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)

	log.Println("🎉 统一服务启动完成!")
	log.Println(strings.Repeat("=", 60))
	log.Printf("📍 服务地址: http://%s", addr)
	log.Printf("🔍 健康检查: http://%s/health", addr)
	log.Printf("📖 服务概览: http://%s/", addr)
	log.Printf("🎯 演示页面: http://%s/demo/", addr)
	log.Println()

	log.Println("📋 可用API端点:")
	log.Println("┌─ 传统API (Traditional)")
	log.Printf("│  ├─ 用户列表: GET %s/api/v1/traditional/users", addr)
	log.Printf("│  ├─ 用户详情: GET %s/api/v1/traditional/users/:id", addr)
	log.Printf("│  ├─ 用户搜索: GET %s/api/v1/traditional/search/users?keyword=xxx", addr)
	log.Printf("│  └─ 用户统计: GET %s/api/v1/traditional/stats/users", addr)
	log.Println("│")

	if mcpClient != nil {
		log.Println("├─ MCP增强API (AI Enhanced)")
		log.Printf("│  ├─ AI查询: POST %s/api/v1/mcp/query/users", addr)
		log.Printf("│  ├─ AI分析: GET %s/api/v1/mcp/analyze/users?type=xxx", addr)
		log.Printf("│  ├─ AI报告: GET %s/api/v1/mcp/report/users?type=xxx", addr)
		log.Printf("│  └─ 智能搜索: GET %s/api/v1/mcp/search/smart?q=xxx", addr)
		log.Println("│")

		log.Println("└─ 对比分析 (Comparison)")
		log.Printf("   ├─ 方法对比: GET %s/api/v1/comparison/methods", addr)
		log.Printf("   └─ 能力展示: GET %s/api/v1/comparison/capabilities", addr)
	} else {
		log.Println("└─ 注意: MCP服务不可用，仅提供传统API功能")
	}
	log.Println()

	log.Println("💡 使用建议:")
	log.Println("  • 简单查询使用传统API（速度快）")
	if mcpClient != nil {
		log.Println("  • 复杂分析使用MCP增强API（功能强）")
		log.Println("  • 比较不同方法的性能和结果")
	} else {
		log.Println("  • 启用MCP服务器以获得AI增强功能")
	}
	log.Println(strings.Repeat("=", 60))

	if err := r.Run(addr); err != nil {
		log.Fatalf("❌ 启动服务器失败: %v", err)
	}
}
