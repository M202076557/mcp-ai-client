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

// QueryUser 查询mcp_user表
func (c *MySQLClient) QueryUser() ([]map[string]interface{}, error) {
	query := "SELECT * FROM `mcp_user` LIMIT 100"
	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询mcp_user表失败: %v", err)
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

// QueryUserByID 根据ID查询mcp_user表
func (c *MySQLClient) QueryUserByID(id int) (map[string]interface{}, error) {
	query := "SELECT * FROM `mcp_user` WHERE id = ?"
	row := c.db.QueryRow(query, id)

	// 对于单行查询，我们需要知道列的结构
	// 这里我们使用一个通用的查询来获取列信息
	columnsQuery := "SELECT * FROM `mcp_user` LIMIT 1"
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

// GetUserCount 获取mcp_user表记录数
func (c *MySQLClient) GetUserCount() (int, error) {
	query := "SELECT COUNT(*) FROM `mcp_user`"
	var count int
	err := c.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("获取mcp_user表记录数失败: %v", err)
	}
	return count, nil
}

// GetUserSchema 获取mcp_user表结构
func (c *MySQLClient) GetUserSchema() ([]map[string]interface{}, error) {
	query := "DESCRIBE `mcp_user`"
	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("获取mcp_user表结构失败: %v", err)
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
