package repository

import (
	"fmt"
	"log"
	"os"
	"time"

	"GoCodeMentor/internal/model"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config 结构体，用于映射配置文件
// 确保每个字段都有 mapstructure 标签
type SqlConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Dbname   string `mapstructure:"dbname"`
	Sslmode  string `mapstructure:"sslmode"`
	Timezone string `mapstructure:"timezone"`
}

func InitDB() (*gorm.DB, error) {
	// 加载sql配置
	config, err := LoadSqlConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// 先尝试连接目标数据库
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		config.Host, config.Username, config.Password, config.Dbname, config.Port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	// 如果连接失败（可能是数据库不存在），尝试创建数据库
	if err != nil {
		fmt.Printf("数据库 %s 不存在，正在自动创建...\n", config.Dbname)

		// 连接到默认的 postgres 数据库（这个一定存在）
		defaultDsn := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%s sslmode=disable",
			config.Host, config.Username, config.Password, config.Port)

		defaultDB, err := gorm.Open(postgres.Open(defaultDsn), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("连接默认数据库失败: %v", err)
		}

		// 创建目标数据库（使用原始 SQL）
		sql := fmt.Sprintf("CREATE DATABASE %s", config.Dbname)
		if err := defaultDB.Exec(sql).Error; err != nil {
			// 如果错误是数据库已存在，忽略错误；否则返回错误
			// 注意：不同的 PostgreSQL 驱动错误信息可能不同
			fmt.Printf("创建数据库时出错（可能已存在）: %v\n", err)
		} else {
			fmt.Printf("数据库 %s 创建成功！\n", config.Dbname)
		}

		// 关闭默认连接
		sqlDB, _ := defaultDB.DB()
		sqlDB.Close()

		// 重新连接目标数据库
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			return nil, fmt.Errorf("连接新创建的数据库失败: %v", err)
		}
	}

	// 自动创建表结构
	// fmt.Println("正在创建数据表...")
	err = db.AutoMigrate(
		&model.User{},
		&model.Class{},
		&model.ChatSession{},
		&model.ChatMessage{},
		&model.Assignment{},
		&model.Question{},
		&model.Submission{},
		&model.Feedback{},
		&model.AssignmentClass{},
		&model.ResourceLike{}, // 新增资源点赞模型
		&model.Resource{},
		&model.KnowledgePoint{},
		&model.KnowledgePointCategory{},
	)
	if err != nil {
		return nil, err
	}

	// 初始化推荐资源数据
	seedInitialResources(db)

	// 初始化知识图谱数据
	seedInitialKnowledgeGraph(db)

	// 初始化管理员账号
	initAdminUser(db)

	return db, nil
}

// seedInitialResources seeds the database with a predefined list of resources
// if the resources table is empty.
func seedInitialResources(db *gorm.DB) {
	resources := []model.Resource{
		// Official
		{ResourceID: "go-tour", Title: "A Tour of Go", URL: "https://go.dev/tour/", Description: "官方的 Go 语言入门教程，通过在线编程环境交互式学习，是入门首选。", Category: "official", IconURL: "https://www.google.com/s2/favicons?sz=64&domain=go.dev"},
		{ResourceID: "go-doc", Title: "Go Documentation", URL: "https://go.dev/doc/", Description: "最权威的 Go 语言官方文档，包含语言规范、标准库和工具文档。", Category: "official", IconURL: "https://www.google.com/s2/favicons?sz=64&domain=go.dev"},
		{ResourceID: "go-blog", Title: "The Go Blog", URL: "https://go.dev/blog/", Description: "Go 团队官方博客，发布关于语言更新、最佳实践和内部实现的深度文章。", Category: "official", IconURL: "https://www.google.com/s2/favicons?sz=64&domain=go.dev"},

		// Practice
		{ResourceID: "awesome-go", Title: "Awesome Go", URL: "https://github.com/avelino/awesome-go", Description: "一个由社区维护的、非常全面的Go语言框架、库和软件列表，是探索Go生态的绝佳入口。", Category: "practice", IconURL: "https://www.google.com/s2/favicons?sz=64&domain=github.com"},
		{ResourceID: "project-layout", Title: "Standard Go Project Layout", URL: "https://github.com/golang-standards/project-layout", Description: "Go应用程序的通用项目结构布局模板，提供了一套推荐的目录组织方式。", Category: "practice", IconURL: "https://www.google.com/s2/favicons?sz=64&domain=github.com"},
		{ResourceID: "gin-web-framework", Title: "Gin Web Framework", URL: "https://github.com/gin-gonic/gin", Description: "一个高性能的Go Web框架，以其Radix树路由和中间件支持而闻名，本项目就使用了该框架。", Category: "practice", IconURL: "https://www.google.com/s2/favicons?sz=64&domain=github.com"},
		{ResourceID: "gorm", Title: "GORM", URL: "https://github.com/go-gorm/gorm", Description: "Go 语言中最受欢迎的 ORM 库之一，提供了强大且易于使用的 API 来操作数据库。", Category: "practice", IconURL: "https://www.google.com/s2/favicons?sz=64&domain=github.com"},
		{ResourceID: "cobra", Title: "Cobra", URL: "https://github.com/spf13/cobra", Description: "一个用于创建强大的现代CLI应用程序的库，被许多知名项目（如Kubernetes）使用。", Category: "practice", IconURL: "https://www.google.com/s2/favicons?sz=64&domain=github.com"},

		// Community
		{ResourceID: "dave-cheney-blog", Title: "Dave Cheney's Blog", URL: "https://dave.cheney.net/", Description: "Go语言核心贡献者之一的博客，包含大量关于Go性能、设计和最佳实践的深度好文。", Category: "community", IconURL: "https://www.google.com/s2/favicons?sz=64&domain=dave.cheney.net"},
		{ResourceID: "studygolang", Title: "Go语言中文网", URL: "https://studygolang.com/", Description: "国内最活跃的Go语言社区之一，提供新闻、教程、招聘信息和活跃的论坛。", Category: "community", IconURL: "https://www.google.com/s2/favicons?sz=64&domain=studygolang.com"},
		{ResourceID: "go-by-example", Title: "Go by Example", URL: "https://gobyexample.com/", Description: "通过简洁、带注释的示例代码来学习Go语言的网站，非常适合快速查阅。", Category: "community", IconURL: "https://www.google.com/s2/favicons?sz=64&domain=gobyexample.com"},
		{ResourceID: "ardan-labs-blog", Title: "Ardan Labs Blog", URL: "https://www.ardanlabs.com/blog/", Description: "由知名Go培训机构Ardan Labs维护的博客，内容深入且专业，适合进阶学习。", Category: "community", IconURL: "https://www.google.com/s2/favicons?sz=64&domain=ardanlabs.com"},

		// CS
		{ResourceID: "system-design-primer", Title: "System Design Primer", URL: "https://github.com/donnemartin/system-design-primer", Description: "系统设计的终极指南。学习如何设计可扩展、高可用的系统，是后端工程师的必读材料。", Category: "cs", IconURL: "https://www.google.com/s2/favicons?sz=64&domain=github.com"},
		{ResourceID: "missing-semester", Title: "The Missing Semester of Your CS Education", URL: "https://missing.csail.mit.edu/", Description: "MIT开设的“计算机教育中缺失的一课”，涵盖命令行、Vim、Git等程序员必备的强大工具。", Category: "cs", IconURL: "https://www.google.com/s2/favicons?sz=64&domain=missing.csail.mit.edu"},
		{ResourceID: "build-your-own-x", Title: "Build your own X", URL: "https://github.com/codecrafters-io/build-your-own-x", Description: "一个通过从零开始构建技术（如数据库、Git、Docker）来学习的绝佳项目集合。", Category: "cs", IconURL: "https://www.google.com/s2/favicons?sz=64&domain=github.com"},
		{ResourceID: "developer-roadmap", Title: "Developer Roadmap", URL: "https://roadmap.sh/", Description: "为开发者提供的技能树和学习路线图，涵盖前后端、DevOps等多个方向。", Category: "cs", IconURL: "https://www.google.com/s2/favicons?sz=64&domain=roadmap.sh"},
	}

	for _, r := range resources {
		// Use FirstOrCreate with Assign to ensure data is always up-to-date with the code definition.
		// This will create the resource if it doesn't exist, or update it if it does.
		if err := db.Where(model.Resource{ResourceID: r.ResourceID}).Assign(r).FirstOrCreate(&model.Resource{}).Error; err != nil {
			log.Printf("Failed to seed resource %s: %v", r.Title, err)
		}
	}
}

func seedInitialKnowledgeGraph(db *gorm.DB) {
	// 1. Create Categories
	categories := []model.KnowledgePointCategory{
		{Name: "基础语法"},
		{Name: "并发编程"},
		{Name: "Web开发"},
		{Name: "区块链应用"},
	}
	for i := range categories {
		db.FirstOrCreate(&categories[i], model.KnowledgePointCategory{Name: categories[i].Name})
	}
	// Create a map for easy lookup
	catMap := make(map[string]uint)
	for _, c := range categories {
		catMap[c.Name] = c.ID
	}

	// 2. Define Points and their relationships
	pointsData := map[string]struct {
		Category   string
		ParentName string
	}{
		"Go基础":      {Category: "基础语法"},
		"变量与数据类型":   {Category: "基础语法", ParentName: "Go基础"},
		"控制流":       {Category: "基础语法", ParentName: "Go基础"},
		"函数":        {Category: "基础语法", ParentName: "Go基础"},
		"结构体与方法":    {Category: "基础语法", ParentName: "Go基础"},
		"并发编程":      {Category: "并发编程", ParentName: "Go基础"},
		"Goroutine": {Category: "并发编程", ParentName: "并发编程"},
		"Channel":   {Category: "并发编程", ParentName: "并发编程"},
		"sync包":     {Category: "并发编程", ParentName: "并发编程"},
		"Web开发":     {Category: "Web开发", ParentName: "Go基础"},
		"net/http":  {Category: "Web开发", ParentName: "Web开发"},
		"Gin框架":     {Category: "Web开发", ParentName: "Web开发"},
		"区块链应用":     {Category: "区块链应用"},
		"智能合约":      {Category: "区块链应用", ParentName: "结构体与方法"}, // Cross-category
		"分布式系统":     {Category: "区块链应用", ParentName: "并发编程"},   // Cross-category
	}

	// 3. Create all point objects and store them in a map for relationship building
	pointMap := make(map[string]*model.KnowledgePoint)
	for name, data := range pointsData {
		point := &model.KnowledgePoint{
			Name:       name,
			CategoryID: catMap[data.Category],
		}
		db.FirstOrCreate(point, model.KnowledgePoint{Name: name})
		pointMap[name] = point
	}

	// 4. Iterate again to set parent relationships
	for name, data := range pointsData {
		if data.ParentName != "" {
			child := pointMap[name]
			parent := pointMap[data.ParentName]
			if child != nil && parent != nil {
				child.ParentID = &parent.ID
				db.Save(child)
			}
		}
	}
}

func initAdminUser(db *gorm.DB) {
	var count int64
	db.Model(&model.User{}).Where("role = ?", "admin").Count(&count)
	if count == 0 {
		fmt.Println("正在初始化管理员账号...")
		// 检查用户名 admin 是否已存在
		var existingUser model.User
		if err := db.Where("username = ?", "admin").First(&existingUser).Error; err != nil {
			// 如果用户名 admin 不存在，则创建一个
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
			admin := &model.User{
				ID:        uuid.New().String(),
				Username:  "admin",
				Password:  string(hashedPassword),
				Name:      "系统管理员",
				Role:      "admin",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := db.Create(admin).Error; err != nil {
				fmt.Printf("初始化管理员账号失败: %v\n", err)
			} else {
				fmt.Println("管理员账号 (admin/admin123) 初始化成功！")
			}
		} else {
			// 如果 admin 用户已存在但角色不是 admin，则更新角色
			existingUser.Role = "admin"
			db.Save(&existingUser)
			fmt.Println("已将现有用户 admin 的角色更新为管理员")
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// LoadSqlConfig 加载配置文件
func LoadSqlConfig() (*SqlConfig, error) {
	v := viper.New()

	v.SetConfigFile("./configs/sql_config.yaml") // 注意：路径是相对于执行 `go run` 命令的目录

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config file not found: %w", err)
	}

	log.Printf("Successfully loaded config file: %s", v.ConfigFileUsed())

	var config SqlConfig
	// 使用 Unmarshal 填充结构体
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	// 调试：打印加载后的配置，确认是否为空
	log.Printf("Loaded configuration: %+v", config)

	return &config, nil
}
