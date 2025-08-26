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
	log.Println("🚀 启动MCP AI Client - 简化版")
	log.Println("📋 功能: 5类AI增强工具 + 基础数据库查询")

	// 加载配置
	config, err := loadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 1. 初始化MySQL客户端 (基础数据库服务)
	log.Println("🔗 初始化MySQL数据库连接...")
	mysqlClient, err := database.NewMySQLClient(&config.Database.MySQL)
	if err != nil {
		log.Fatalf("初始化MySQL客户端失败: %v", err)
	}
	defer mysqlClient.Close()
	log.Println("✅ MySQL连接成功")

	// 2. 初始化MCP客户端 (AI增强服务)
	log.Println("🤖 初始化MCP AI客户端...")
	mcpClient, err := mcp.NewMCPClient(config.MCP.ServerURL, config.MCP.Timeout)
	if err != nil {
		log.Fatalf("MCP客户端初始化失败: %v", err)
	}
	defer mcpClient.Close()

	// 测试MCP连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := mcpClient.Initialize(ctx); err != nil {
		log.Fatalf("MCP连接失败: %v", err)
	}
	log.Println("✅ MCP连接成功")

	// 3. 创建AI配置
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

	// 4. 创建数据库配置
	dbConfig := &api.DatabaseConfig{
		UserTable: config.Database.Tables.UserTable,
	}
	if dbConfig.UserTable == "" {
		dbConfig.UserTable = "mcp_user" // 默认表名
	}

	// 5. 创建API处理器
	log.Println("🌐 初始化API处理器...")
	handlers := api.NewHandlers(mysqlClient, mcpClient, aiConfig, dbConfig)
	log.Println("✅ API处理器已就绪")

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
			"service":     "MCP AI Client - 简化版",
			"description": "5类AI增强工具 + 基础数据库查询",
			"version":     "2.0.0",
			"features": []string{
				"AI对话 (ai_chat)",
				"AI文件管理 (ai_file_manager)",
				"AI数据处理 (ai_data_processor)",
				"AI网络请求 (ai_api_client)",
				"AI数据库查询 (ai_query_with_analysis)",
				"基础数据库查询",
			},
			"api_groups": gin.H{
				"health":    "/health",
				"ai_tools":  "/api/v1/ai/*",
				"database":  "/api/v1/db/*",
			},
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// 健康检查
	r.GET("/health", handlers.HealthCheck)

	// ===== AI工具API路由 (5.1-5.5) =====
	aiV1 := r.Group("/api/v1/ai")
	{
		// 5.1 基础AI对话
		aiV1.POST("/chat", handlers.MCPChatHandler)
		
		// 5.2 AI智能文件管理
		aiV1.POST("/file-manager", handlers.MCPFileManagerHandler)
		
		// 5.3 AI智能数据处理
		aiV1.POST("/data-processor", handlers.MCPDataProcessorHandler)
		
		// 5.4 AI智能网络请求
		aiV1.POST("/api-client", handlers.MCPAPIClientHandler)
		
		// 5.5 AI智能数据库查询
		aiV1.POST("/query-with-analysis", handlers.MCPQueryWithAnalysisHandler)
	}

	// ===== 基础数据库查询API =====
	dbV1 := r.Group("/api/v1/db")
	{
		// 基础用户查询
		dbV1.GET("/users", handlers.GetUsersTraditional)
	}

	log.Println("✅ 所有API路由已配置")

	// 7. 启动服务器
	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)

	log.Println("🎉 MCP AI Client 简化版启动完成!")
	log.Println(strings.Repeat("=", 60))
	log.Printf("📍 服务地址: http://%s", addr)
	log.Printf("🔍 健康检查: http://%s/health", addr)
	log.Printf("📖 服务概览: http://%s/", addr)
	log.Println()

	log.Println("📋 可用API端点:")
	log.Println("┌─ AI增强工具 (5.1-5.5)")
	log.Printf("│  ├─ 5.1 AI对话: POST %s/api/v1/ai/chat", addr)
	log.Printf("│  ├─ 5.2 文件管理: POST %s/api/v1/ai/file-manager", addr)
	log.Printf("│  ├─ 5.3 数据处理: POST %s/api/v1/ai/data-processor", addr)
	log.Printf("│  ├─ 5.4 网络请求: POST %s/api/v1/ai/api-client", addr)
	log.Printf("│  └─ 5.5 数据库查询: POST %s/api/v1/ai/query-with-analysis", addr)
	log.Println("│")
	log.Println("└─ 基础数据库查询")
	log.Printf("   └─ 用户列表: GET %s/api/v1/db/users", addr)
	log.Println()

	log.Println("💡 使用说明:")
	log.Println("  • AI工具: 使用POST请求调用AI增强功能")
	log.Println("  • 数据库: 使用GET请求进行基础数据查询")
	log.Println("  • 所有AI工具都支持自然语言交互")
	log.Println(strings.Repeat("=", 60))

	if err := r.Run(addr); err != nil {
		log.Fatalf("❌ 启动服务器失败: %v", err)
	}
}
