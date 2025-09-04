package service

import (
	"fmt"
	"log"
	"mcp-ai-client/internal/database"
	"strconv"
	"strings"
	"time"
)

// UserService 传统用户服务
type UserService struct {
	mysqlClient *database.MySQLClient
	userTable   string // 用户表名
}

// NewUserService 创建用户服务
func NewUserService(mysqlClient *database.MySQLClient, userTable string) *UserService {
	return &UserService{
		mysqlClient: mysqlClient,
		userTable:   userTable,
	}
}

// User 用户结构体
type User struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	Department string  `json:"department"`
	Age        int     `json:"age"`
	Salary     float64 `json:"salary"`
}

// GetAllUsers 获取所有用户 - 传统方法
func (s *UserService) GetAllUsers() ([]User, error) {
	start := time.Now()
	log.Printf("🔍 [传统查询] 开始查询所有用户...")

	// 直接调用数据库
	data, err := s.mysqlClient.QueryUser(s.userTable)
	if err != nil {
		log.Printf("❌ [传统查询] 查询失败: %v", err)
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}

	// 转换数据格式（兼容 MySQL 返回的多种类型：int64、[]uint8、string 等）
	users := make([]User, 0, len(data))
	for _, row := range data {
		user := User{}
		// ID
		if id64, ok := row["id"].(int64); ok {
			user.ID = int(id64)
		} else if b, ok := row["id"].([]uint8); ok {
			if s := string(b); s != "" {
				if v, err := strconv.Atoi(s); err == nil {
					user.ID = v
				}
			}
		}
		// Name
		if name, ok := row["name"].(string); ok {
			user.Name = name
		} else if b, ok := row["name"].([]uint8); ok {
			user.Name = string(b)
		}
		// Email
		if email, ok := row["email"].(string); ok {
			user.Email = email
		} else if b, ok := row["email"].([]uint8); ok {
			user.Email = string(b)
		}
		// Department
		if d, ok := row["department"].(string); ok {
			user.Department = d
		} else if b, ok := row["department"].([]uint8); ok {
			user.Department = string(b)
		}
		// Age
		if age64, ok := row["age"].(int64); ok {
			user.Age = int(age64)
		} else if b, ok := row["age"].([]uint8); ok {
			if s := string(b); s != "" {
				if v, err := strconv.Atoi(s); err == nil {
					user.Age = v
				}
			}
		}
		// Salary (DECIMAL 常为 []uint8 或 string)
		if b, ok := row["salary"].([]uint8); ok {
			if salaryStr := string(b); salaryStr != "" {
				if salaryFloat, err := strconv.ParseFloat(salaryStr, 64); err == nil {
					user.Salary = salaryFloat
				}
			}
		} else if s, ok := row["salary"].(string); ok {
			if salaryFloat, err := strconv.ParseFloat(s, 64); err == nil {
				user.Salary = salaryFloat
			}
		}
		users = append(users, user)
	}

	duration := time.Since(start)
	log.Printf("✅ [传统查询] 查询完成，共找到 %d 个用户，耗时: %v", len(users), duration)

	return users, nil
}

// GetUserByID 根据ID获取用户 - 传统方法
func (s *UserService) GetUserByID(id int) (*User, error) {
	start := time.Now()
	log.Printf("🔍 [传统查询] 开始查询用户 ID: %d", id)

	// 直接调用数据库
	data, err := s.mysqlClient.QueryUserByID(id, s.userTable)
	if err != nil {
		log.Printf("❌ [传统查询] 查询失败: %v", err)
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}

	// 转换数据格式
	user := &User{}
	if userId, ok := data["id"].(int64); ok {
		user.ID = int(userId)
	}
	if name, ok := data["name"].(string); ok {
		user.Name = name
	}
	if email, ok := data["email"].(string); ok {
		user.Email = email
	}
	if department, ok := data["department"].(string); ok {
		user.Department = department
	}
	if age, ok := data["age"].(int64); ok {
		user.Age = int(age)
	}
	if salary, ok := data["salary"].([]uint8); ok {
		// MySQL DECIMAL类型返回的是[]uint8，需要转换为字符串再转为float64
		if salaryStr := string(salary); salaryStr != "" {
			if salaryFloat, err := strconv.ParseFloat(salaryStr, 64); err == nil {
				user.Salary = salaryFloat
			}
		}
	}

	duration := time.Since(start)
	log.Printf("✅ [传统查询] 查询完成，用户: %s，耗时: %v", user.Name, duration)

	return user, nil
}

// SearchUsers 搜索用户 - 传统方法
func (s *UserService) SearchUsers(keyword string) ([]User, error) {
	start := time.Now()
	log.Printf("🔍 [传统查询] 开始搜索用户，关键词: %s", keyword)

	// 获取所有用户并过滤（简单实现）
	allUsers, err := s.GetAllUsers()
	if err != nil {
		return nil, err
	}

	// 简单的关键词匹配
	var filteredUsers []User
	for _, user := range allUsers {
		if contains(user.Name, keyword) || contains(user.Email, keyword) || contains(user.Department, keyword) {
			filteredUsers = append(filteredUsers, user)
		}
	}

	duration := time.Since(start)
	log.Printf("✅ [传统查询] 搜索完成，找到 %d 个匹配用户，耗时: %v", len(filteredUsers), duration)

	return filteredUsers, nil
}

// GetUserStats 获取用户统计 - 传统方法
func (s *UserService) GetUserStats() (map[string]interface{}, error) {
	start := time.Now()
	log.Printf("🔍 [传统查询] 开始统计用户数据...")

	users, err := s.GetAllUsers()
	if err != nil {
		return nil, err
	}

	// 计算统计信息
	totalUsers := len(users)
	ageSum := 0
	salarySum := 0.0
	departments := make(map[string]int)
	emailDomains := make(map[string]int)

	for _, user := range users {
		ageSum += user.Age
		salarySum += user.Salary

		// 统计部门分布
		if user.Department != "" {
			departments[user.Department]++
		}

		// 统计邮箱域名
		if user.Email != "" {
			if emailParts := strings.Split(user.Email, "@"); len(emailParts) == 2 {
				domain := emailParts[1]
				emailDomains[domain]++
			}
		}
	}

	avgAge := 0.0
	avgSalary := 0.0
	if totalUsers > 0 {
		avgAge = float64(ageSum) / float64(totalUsers)
		avgSalary = salarySum / float64(totalUsers)
	}

	stats := map[string]interface{}{
		"total_users":    totalUsers,
		"average_age":    avgAge,
		"average_salary": avgSalary,
		"departments":    departments,
		"email_domains":  emailDomains,
		"query_method":   "traditional_database",
		"query_time":     time.Since(start).String(),
	}

	duration := time.Since(start)
	log.Printf("✅ [传统查询] 统计完成，总用户: %d，平均年龄: %.1f，平均薪资: %.2f，耗时: %v", totalUsers, avgAge, avgSalary, duration)

	return stats, nil
}

// GetUserStatsWithTable 获取指定表的用户统计 - 传统方法
func (s *UserService) GetUserStatsWithTable(tableName string) (map[string]interface{}, error) {
	start := time.Now()
	log.Printf("🔍 [传统查询] 开始统计用户数据，表: %s...", tableName)

	// 直接调用数据库查询指定表
	data, err := s.mysqlClient.QueryUser(tableName)
	if err != nil {
		log.Printf("❌ [传统查询] 查询失败: %v", err)
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}

	// 转换数据格式并计算统计
	totalUsers := len(data)
	ageSum := 0
	salarySum := 0.0
	departments := make(map[string]int)
	emailDomains := make(map[string]int)
	userCount := 0

	for _, row := range data {
		// 解析年龄
		if ageVal, ok := row["age"]; ok && ageVal != nil {
			if age, ok := ageVal.(int64); ok {
				ageSum += int(age)
				userCount++
			}
		}

		// 解析薪资
		if salaryVal, ok := row["salary"]; ok && salaryVal != nil {
			if salary, ok := salaryVal.([]uint8); ok {
				if salaryStr := string(salary); salaryStr != "" {
					if salaryFloat, err := strconv.ParseFloat(salaryStr, 64); err == nil {
						salarySum += salaryFloat
					}
				}
			}
		}

		// 统计部门分布
		if deptVal, ok := row["department"]; ok && deptVal != nil {
			if dept, ok := deptVal.(string); ok && dept != "" {
				departments[dept]++
			}
		}

		// 统计邮箱域名
		if emailVal, ok := row["email"]; ok && emailVal != nil {
			if email, ok := emailVal.(string); ok && email != "" {
				if emailParts := strings.Split(email, "@"); len(emailParts) == 2 {
					domain := emailParts[1]
					emailDomains[domain]++
				}
			}
		}
	}

	avgAge := 0.0
	avgSalary := 0.0
	if userCount > 0 {
		avgAge = float64(ageSum) / float64(userCount)
		avgSalary = salarySum / float64(userCount)
	}

	stats := map[string]interface{}{
		"total_users":    totalUsers,
		"average_age":    avgAge,
		"average_salary": avgSalary,
		"departments":    departments,
		"email_domains":  emailDomains,
		"table_name":     tableName,
		"query_method":   "traditional_database",
		"query_time":     time.Since(start).String(),
	}

	duration := time.Since(start)
	log.Printf("✅ [传统查询] 统计完成，表: %s，总用户: %d，平均年龄: %.1f，平均薪资: %.2f，耗时: %v", tableName, totalUsers, avgAge, avgSalary, duration)

	return stats, nil
}

// contains 简单的字符串包含检查（忽略大小写）
func contains(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

// 性能测试方法
func (s *UserService) BenchmarkQuery(iterations int) map[string]interface{} {
	log.Printf("🚀 [性能测试] 开始传统查询性能测试，迭代次数: %d", iterations)

	start := time.Now()
	var totalDuration time.Duration
	successCount := 0

	for i := 0; i < iterations; i++ {
		iterStart := time.Now()
		_, err := s.GetAllUsers()
		iterDuration := time.Since(iterStart)
		totalDuration += iterDuration

		if err == nil {
			successCount++
		}

		if i%10 == 0 {
			log.Printf("📊 [性能测试] 已完成 %d/%d 次查询", i+1, iterations)
		}
	}

	totalTime := time.Since(start)
	avgDuration := totalDuration / time.Duration(iterations)

	result := map[string]interface{}{
		"method":           "traditional_database",
		"total_iterations": iterations,
		"success_count":    successCount,
		"total_time":       totalTime.String(),
		"average_time":     avgDuration.String(),
		"queries_per_sec":  float64(iterations) / totalTime.Seconds(),
		"success_rate":     float64(successCount) / float64(iterations) * 100,
	}

	log.Printf("✅ [性能测试] 传统查询测试完成")
	log.Printf("📈 总耗时: %v, 平均耗时: %v, QPS: %.2f",
		totalTime, avgDuration, result["queries_per_sec"])

	return result
}
