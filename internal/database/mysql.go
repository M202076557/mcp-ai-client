package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Database  string `yaml:"database"`
	Charset   string `yaml:"charset"`
	ParseTime bool   `yaml:"parse_time"`
	Loc       string `yaml:"loc"`
}

// MySQLClient MySQL客户端
type MySQLClient struct {
	db *sql.DB
}

// NewMySQLClient 创建MySQL客户端
func NewMySQLClient(config *MySQLConfig) (*MySQLClient, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%v&loc=%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.Charset,
		config.ParseTime,
		config.Loc,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接MySQL失败: %v", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("MySQL连接测试失败: %v", err)
	}

	log.Println("MySQL连接成功")
	return &MySQLClient{db: db}, nil
}

// Close 关闭数据库连接
func (c *MySQLClient) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// QueryUser 查询指定用户表
func (c *MySQLClient) QueryUser(tableName string) ([]map[string]interface{}, error) {
	if tableName == "" {
		tableName = "mcp_user" // 默认表名
	}
	query := fmt.Sprintf("SELECT * FROM `%s` LIMIT 100", tableName)
	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询%s表失败: %v", tableName, err)
	}
	defer rows.Close()

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("获取列信息失败: %v", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		// 创建值的切片
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// 扫描行数据
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("扫描行数据失败: %v", err)
		}

		// 构建结果map
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val != nil {
				row[col] = val
			} else {
				row[col] = nil
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历结果集失败: %v", err)
	}

	return results, nil
}

// QueryUserByID 根据ID查询指定用户表
func (c *MySQLClient) QueryUserByID(id int, tableName string) (map[string]interface{}, error) {
	if tableName == "" {
		tableName = "mcp_user" // 默认表名
	}
	query := fmt.Sprintf("SELECT * FROM `%s` WHERE id = ?", tableName)
	row := c.db.QueryRow(query, id)

	// 对于单行查询，我们需要知道列的结构
	// 这里我们使用一个通用的查询来获取列信息
	columnsQuery := fmt.Sprintf("SELECT * FROM `%s` LIMIT 1", tableName)
	columnsRow, err := c.db.Query(columnsQuery)
	if err != nil {
		return nil, fmt.Errorf("获取列信息失败: %v", err)
	}
	defer columnsRow.Close()

	columns, err := columnsRow.Columns()
	if err != nil {
		return nil, fmt.Errorf("获取列信息失败: %v", err)
	}

	// 创建值的切片
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// 扫描行数据
	if err := row.Scan(valuePtrs...); err != nil {
		return nil, fmt.Errorf("扫描行数据失败: %v", err)
	}

	// 构建结果map
	result := make(map[string]interface{})
	for i, col := range columns {
		val := values[i]
		if val != nil {
			result[col] = val
		} else {
			result[col] = nil
		}
	}

	return result, nil
}

// GetUserCount 获取指定用户表记录数
func (c *MySQLClient) GetUserCount(tableName string) (int, error) {
	if tableName == "" {
		tableName = "mcp_user" // 默认表名
	}
	query := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)
	var count int
	err := c.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("获取%s表记录数失败: %v", tableName, err)
	}
	return count, nil
}

// GetUserSchema 获取指定用户表结构
func (c *MySQLClient) GetUserSchema(tableName string) ([]map[string]interface{}, error) {
	if tableName == "" {
		tableName = "mcp_user" // 默认表名
	}
	query := fmt.Sprintf("DESCRIBE `%s`", tableName)
	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("获取%s表结构失败: %v", tableName, err)
	}
	defer rows.Close()

	var schema []map[string]interface{}
	columns, _ := rows.Columns()
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val != nil {
				row[col] = val
			} else {
				row[col] = nil
			}
		}
		schema = append(schema, row)
	}

	return schema, nil
}
