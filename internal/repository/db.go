package repository

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
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
	// Check if resources already exist
	var count int64
	db.Model(&model.Resource{}).Count(&count)
	if count > 0 {
		return // Resources already seeded
	}
	log.Println("Seeding initial resources...")
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
	var count int64
	db.Model(&model.KnowledgePoint{}).Count(&count)
	if count > 0 {
		// log.Println("Knowledge graph already seeded. Skipping.")
		return
	}
	log.Println("Seeding initial knowledge graph data...")

	tx := db.Begin()
	if tx.Error != nil {
		log.Printf("failed to begin transaction: %v", tx.Error)
		return
	}

	// 1. Clean up old data
	// No longer cleaning up data, this seeding runs only once.

	// 2. Define and insert new categories based on Level 1 nodes
	categories := []model.KnowledgePointCategory{
		{Name: "Go语言基础"},
		{Name: "并发编程"},
		{Name: "标准库"},
		{Name: "工程实践"},
		{Name: "高级特性"},
	}
	for i := range categories {
		if err := tx.Create(&categories[i]).Error; err != nil {
			log.Printf("Failed to create category %s: %v", categories[i].Name, err)
			tx.Rollback()
			return
		}
	}
	catMap := make(map[string]uint)
	for _, c := range categories {
		catMap[c.Name] = c.ID
	}

	// 3. Parse the raw text data
	rawData := `
		节点名称：Go语言基础;节点级别：1级节点;前置节点：无;后置节点:环境搭建与工具链;节点描述:Go语言的入门基石，涵盖环境配置、工具链、基本语法及核心数据类型;
		节点名称：环境搭建与工具链;节点级别：2级节点;前置节点：Go语言基础;后置节点:基本语法;节点描述:安装Go环境，学习go命令（run、build、install、fmt等），配置工作区和环境变量;
		节点名称：安装Go;节点级别：3级节点;前置节点：环境搭建与工具链;后置节点:GOPATH与工作区;节点描述:介绍不同操作系统下Go的安装方法，验证安装，环境变量GOROOT的基本概念;
		节点名称：GOPATH与工作区;节点级别：3级节点;前置节点：安装Go;后置节点:Go Modules;节点描述:传统GOPATH模式下的工作区结构（src、pkg、bin），以及其在现代开发中的演变;
		节点名称：Go Modules;节点级别：3级节点;前置节点：GOPATH与工作区;后置节点:常用go命令;节点描述:Go Modules的引入，mod文件结构，初始化模块，管理依赖;
		节点名称：常用go命令;节点级别：3级节点;前置节点：Go Modules;后置节点:基本语法;节点描述:go run、go build、go install、go fmt、go test等常用命令的使用和区别;
		节点名称：基本语法;节点级别：2级节点;前置节点：环境搭建与工具链;后置节点:数据类型;节点描述:Go程序的基本结构、关键字、运算符、语句结束、注释等基础语法;
		节点名称：程序结构与关键字;节点级别：3级节点;前置节点：基本语法;后置节点:标识符与运算符;节点描述:Go程序的基本组成（包声明、导入、函数等），25个关键字分类及作用;
		节点名称：标识符与运算符;节点级别：3级节点;前置节点：程序结构与关键字;后置节点:字面量与常量;节点描述:标识符命名规则，空白标识符，运算符（算术、关系、逻辑、位、赋值）优先级;
		节点名称：字面量与常量;节点级别：3级节点;前置节点：标识符与运算符;后置节点:变量声明;节点描述:整型、浮点型、字符串字面量，常量的定义（iota枚举）;
		节点名称：变量声明;节点级别：3级节点;前置节点：字面量与常量;后置节点:数据类型;节点描述:var、:=短变量声明，多变量赋值，变量作用域;
		节点名称：数据类型;节点级别：2级节点;前置节点：基本语法;后置节点:复合类型;节点描述:基本数据类型（整型、浮点型、布尔型、字符串）的声明、初始化、操作;
		节点名称：整型与浮点型;节点级别：3级节点;前置节点：数据类型;后置节点:布尔型与字符串;节点描述:有符号/无符号整型（int, int8, uint等），浮点型float32/float64，类型转换;
		节点名称：布尔型与字符串;节点级别：3级节点;前置节点：整型与浮点型;后置节点:复合类型;节点描述:布尔型bool，字符串操作（长度、拼接、切片），字符串不可变性;
		节点名称：控制结构;节点级别：2级节点;前置节点：基本语法;后置节点:函数;节点描述:条件语句（if、switch）、循环语句（for）、延迟执行（defer）等控制流程;
		节点名称：条件语句if;节点级别：3级节点;前置节点：控制结构;后置节点:条件语句switch;节点描述:if语句语法，if-else，if带初始化语句;
		节点名称：条件语句switch;节点级别：3级节点;前置节点：条件语句if;后置节点:循环语句for;节点描述:switch表达式，多case，fallthrough，type switch;
		节点名称：循环语句for;节点级别：3级节点;前置节点：条件语句switch;后置节点:延迟语句defer;节点描述:for循环三种形式（类似于while、传统for、range），break/continue;
		节点名称：延迟语句defer;节点级别：3级节点;前置节点：循环语句for;后置节点:函数;节点描述:defer的执行时机，栈特性，常见用途（资源释放、解锁）;
		节点名称：函数;节点级别：2级节点;前置节点：控制结构;后置节点:错误处理;节点描述:函数定义、参数传递、返回值、匿名函数、闭包、变参函数;
		节点名称：函数定义与参数;节点级别：3级节点;前置节点：函数;后置节点:返回值与命名返回值;节点描述:函数声明语法，参数传递（值传递与引用类型），变长参数;
		节点名称：返回值与命名返回值;节点级别：3级节点;前置节点：函数定义与参数;后置节点:匿名函数与闭包;节点描述:多返回值，命名返回值，return语句的细节;
		节点名称：匿名函数与闭包;节点级别：3级节点;前置节点：返回值与命名返回值;后置节点:错误处理;节点描述:函数字面量，闭包捕获外部变量，应用场景;
		节点名称：错误处理;节点级别：2级节点;前置节点：函数;后置节点:方法;节点描述:错误类型、错误处理机制、panic和recover的使用;
		节点名称：错误值处理;节点级别：3级节点;前置节点：错误处理;后置节点:自定义错误;节点描述:error接口，检查错误（if err != nil），fmt.Errorf;
		节点名称：自定义错误;节点级别：3级节点;前置节点：错误值处理;后置节点:panic与recover;节点描述:实现error接口，哨兵错误，错误包装与unwrap;
		节点名称：panic与recover;节点级别：3级节点;前置节点：自定义错误;后置节点:复合类型;节点描述:panic引发运行时恐慌，recover捕获，defer结合recover处理panic;
		节点名称：复合类型;节点级别：2级节点;前置节点：数据类型;后置节点:结构体;节点描述:数组、切片、映射的声明、初始化、操作和底层原理;
		节点名称：数组;节点级别：3级节点;前置节点：复合类型;后置节点:切片;节点描述:数组声明、初始化，数组长度作为类型一部分，数组的遍历;
		节点名称：切片;节点级别：3级节点;前置节点：数组;后置节点:映射;节点描述:切片创建（make、字面量），切片操作（append、copy、切片表达式），底层数组共享;
		节点名称：映射;节点级别：3级节点;前置节点：切片;后置节点:结构体;节点描述:map声明初始化，增删改查，注意事项（并发不安全）;
		节点名称：结构体;节点级别：2级节点;前置节点：复合类型;函数;后置节点:方法;节点描述:结构体定义、实例化、字段访问、嵌套结构体、标签;
		节点名称：结构体定义与实例化;节点级别：3级节点;前置节点：结构体;后置节点:结构体嵌套与方法;节点描述:type定义结构体，字段标签，实例化方式（&, new）;
		节点名称：结构体嵌套与方法;节点级别：3级节点;前置节点：结构体定义与实例化;后置节点:方法;节点描述:结构体嵌套（匿名字段），提升字段，结构体与JSON;
		节点名称：方法;节点级别：2级节点;前置节点：结构体;函数;后置节点:接口;节点描述:方法的定义、接收者类型（值接收者、指针接收者）、方法值与方法表达式;
		节点名称：方法定义;节点级别：3级节点;前置节点：方法;后置节点:值接收者与指针接收者;节点描述:方法声明语法，方法接收者类型，方法与函数的区别;
		节点名称：值接收者与指针接收者;节点级别：3级节点;前置节点：方法定义;后置节点:方法表达式;节点描述:两种接收者的区别，修改结构体的场景，性能考量;
		节点名称：方法表达式;节点级别：3级节点;前置节点：值接收者与指针接收者;后置节点:接口;节点描述:通过类型或实例调用方法，方法值，方法表达式;
		节点名称：接口;节点级别：2级节点;前置节点：方法;错误处理;后置节点:并发编程;标准库;工程实践;高级特性;节点描述:接口定义、实现、类型断言、空接口、接口组合与多态;
		节点名称：接口定义与实现;节点级别：3级节点;前置节点：接口;后置节点:空接口与类型断言;节点描述:接口定义，隐式实现，接口值动态类型;
		节点名称：空接口与类型断言;节点级别：3级节点;前置节点：接口定义与实现;后置节点:接口组合;节点描述:interface{}用途，类型断言（x.(T)），类型switch;
		节点名称：接口组合;节点级别：3级节点;前置节点：空接口与类型断言;后置节点:并发编程;标准库;工程实践;高级特性;节点描述:接口嵌入，常用接口（io.Reader/Writer）;
		节点名称：并发编程;节点级别：1级节点;前置节点：接口;后置节点:Goroutine;节点描述:Go并发模型概述，介绍goroutine和channel的基础，实现并发编程;
		节点名称：Goroutine;节点级别：2级节点;前置节点：并发编程;后置节点:Channel;节点描述:goroutine的创建、调度、生命周期，理解并发与并行，GMP模型简介;
		节点名称：goroutine创建;节点级别：3级节点;前置节点：Goroutine;后置节点:goroutine调度;节点描述:使用go关键字启动并发任务，匿名goroutine，理解并发基本单位;
		节点名称：goroutine调度;节点级别：3级节点;前置节点：goroutine创建;后置节点:GMP模型;节点描述:goroutine调度器基本概念，系统线程与goroutine的关系;
		节点名称：GMP模型;节点级别：3级节点;前置节点：goroutine调度;后置节点:Channel;节点描述:G-M-P三者的作用，调度策略（抢占、窃取），简要模型;
		节点名称：Channel;节点级别：2级节点;前置节点：Goroutine;后置节点:Select;同步原语;节点描述:channel的类型、创建、发送与接收操作，缓冲与非缓冲channel，关闭channel;
		节点名称：channel创建与类型;节点级别：3级节点;前置节点：Channel;后置节点:发送与接收操作;节点描述:make创建channel，有缓冲/无缓冲chan，类型;
		节点名称：发送与接收操作;节点级别：3级节点;前置节点：channel创建与类型;后置节点:关闭channel;节点描述:<-操作符，发送接收阻塞特性，循环接收（for range）;
		节点名称：关闭channel;节点级别：3级节点;前置节点：发送与接收操作;后置节点:Select;节点描述:close函数，接收方检测关闭，关闭后发送panic;
		节点名称：Select;节点级别：2级节点;前置节点：Channel;后置节点:同步原语;节点描述:select语句的使用，多路复用，超时控制，default分支;
		节点名称：select多路复用;节点级别：3级节点;前置节点：Select;后置节点:超时与default;节点描述:select同时监听多个channel，随机选择可用的case;
		节点名称：超时与default;节点级别：3级节点;前置节点：select多路复用;后置节点:同步原语;节点描述:结合time.After实现超时控制，default分支非阻塞;
		节点名称：同步原语;节点级别：2级节点;前置节点：Goroutine;Channel;后置节点:无;节点描述:sync包中的Mutex、RWMutex、WaitGroup、Once、Cond等同步工具;
		节点名称：Mutex与RWMutex;节点级别：3级节点;前置节点：同步原语;后置节点:WaitGroup;节点描述:互斥锁sync.Mutex，读写锁sync.RWMutex，使用场景;
		节点名称：WaitGroup;节点级别：3级节点;前置节点：Mutex与RWMutex;后置节点:Once与Cond;节点描述:sync.WaitGroup等待一组goroutine完成，Add/Done/Wait;
		节点名称：Once与Cond;节点级别：3级节点;前置节点：WaitGroup;后置节点:原子操作;节点描述:sync.Once确保函数只执行一次，sync.Cond条件变量;
		节点名称：原子操作;节点级别：3级节点;前置节点：Once与Cond;后置节点:无;节点描述:sync/atomic包对基本类型的原子操作，CAS，原子指针;
		节点名称：标准库;节点级别：1级节点;前置节点：接口;后置节点:输入输出;节点描述:Go标准库概览，涵盖常用包如fmt、io、net、json等的使用;
		节点名称：输入输出;节点级别：2级节点;前置节点：标准库;后置节点:网络编程;节点描述:fmt格式化I/O、io.Reader/Writer接口、bufio缓冲I/O、os文件操作;
		节点名称：fmt格式化输出;节点级别：3级节点;前置节点：输入输出;后置节点:io.Reader/Writer;节点描述:fmt包打印函数（Printf, Println），格式化占位符;
		节点名称：io.Reader/Writer;节点级别：3级节点;前置节点：fmt格式化输出;后置节点:文件操作;节点描述:io包的核心接口，实现Reader/Writer的类型，组合接口;
		节点名称：文件操作;节点级别：3级节点;前置节点：io.Reader/Writer;后置节点:bufio;节点描述:os.Open/Close，读写文件，文件权限，目录操作;
		节点名称：bufio缓冲I/O;节点级别：3级节点;前置节点：文件操作;后置节点:网络编程;节点描述:bufio.Reader/Writer，带缓冲的读写，Scanner;
		节点名称：网络编程;节点级别：2级节点;前置节点：输入输出;并发编程;后置节点:编码处理;节点描述:net包进行TCP/UDP编程，net/http构建HTTP客户端和服务端，中间件;
		节点名称：TCP编程;节点级别：3级节点;前置节点：网络编程;后置节点:UDP编程;节点描述:net包Dial/Listen，建立TCP连接，处理多个连接;
		节点名称：UDP编程;节点级别：3级节点;前置节点：TCP编程;后置节点:HTTP客户端;节点描述:UDP连接和无连接通信，使用net.DialUDP和ListenUDP;
		节点名称：HTTP客户端;节点级别：3级节点;前置节点：UDP编程;后置节点:HTTP服务端;节点描述:http.Get/Post，Client结构，设置超时，请求构造;
		节点名称：HTTP服务端;节点级别：3级节点;前置节点：HTTP客户端;后置节点:编码处理;节点描述:http.Handle/HandleFunc，Server结构，中间件模式;
		节点名称：编码处理;节点级别：2级节点;前置节点：输入输出;后置节点:系统操作;节点描述:encoding/json处理JSON数据，encoding/xml处理XML，以及gob、csv等编码;
		节点名称：JSON序列化;节点级别：3级节点;前置节点：编码处理;后置节点:JSON反序列化;节点描述:json.Marshal，结构体标签（json tag），切片/map编码;
		节点名称：JSON反序列化;节点级别：3级节点;前置节点：JSON序列化;后置节点:其他编码;节点描述:json.Unmarshal，解码到结构体，处理动态键;
		节点名称：其他编码;节点级别：3级节点;前置节点：JSON反序列化;后置节点:系统操作;节点描述:encoding/xml、encoding/gob、encoding/csv等简要使用;
		节点名称：系统操作;节点级别：2级节点;前置节点：输入输出;后置节点:测试与基准;节点描述:os包环境变量、进程操作，path/filepath路径处理，time包时间和定时器;
		节点名称：环境变量与进程;节点级别：3级节点;前置节点：系统操作;后置节点:路径处理;节点描述:os.Getenv/Setenv，os.Args，os.Exec，os.Exit;
		节点名称：路径处理;节点级别：3级节点;前置节点：环境变量与进程;后置节点:时间与定时器;节点描述:path/filepath包处理路径（Join, Split, Glob），相对/绝对路径;
		节点名称：时间与定时器;节点级别：3级节点;前置节点：路径处理;后置节点:测试与基准;节点描述:time包（Now, Sleep, Timer, Ticker），时间格式化;
		节点名称：测试与基准;节点级别：2级节点;前置节点：系统操作;函数;后置节点:工程实践;节点描述:testing包编写单元测试、表驱动测试、性能基准测试、示例函数;
		节点名称：单元测试编写;节点级别：3级节点;前置节点：测试与基准;后置节点:表驱动测试;节点描述:testing.T，测试函数格式，Error/Fatal，子测试;
		节点名称：表驱动测试;节点级别：3级节点;前置节点：单元测试编写;后置节点:基准测试;节点描述:使用匿名结构体切片定义测试用例，提升覆盖率;
		节点名称：基准测试;节点级别：3级节点;前置节点：表驱动测试;后置节点:工程实践;节点描述:testing.B运行基准测试，计时器控制，报告内存分配;
		节点名称：工程实践;节点级别：1级节点;前置节点：标准库;后置节点:包管理与模块;节点描述:Go工程化实践，包括模块管理、项目结构、代码规范等;
		节点名称：包管理与模块;节点级别：2级节点;前置节点：工程实践;后置节点:项目结构;节点描述:go mod命令，模块初始化、依赖管理、版本控制，vendor机制;
		节点名称：go mod命令;节点级别：3级节点;前置节点：包管理与模块;后置节点:依赖管理;节点描述:go mod init/tidy/vendor等命令，go.mod文件语法;
		节点名称：依赖管理;节点级别：3级节点;前置节点：go mod命令;后置节点:版本选择;节点描述:添加/更新依赖，替换（replace），排除（exclude）;
		节点名称：版本选择;节点级别：3级节点;前置节点：依赖管理;后置节点:项目结构;节点描述:语义化版本，最小版本选择，go.sum校验;
		节点名称：项目结构;节点级别：2级节点;前置节点：包管理与模块;后置节点:代码规范与格式化;节点描述:典型Go项目布局（cmd、internal、pkg等），包设计原则，避免循环导入;
		节点名称：标准项目布局;节点级别：3级节点;前置节点：项目结构;后置节点:包设计原则;节点描述:cmd、internal、pkg、api等目录的作用，可执行程序组织;
		节点名称：包设计原则;节点级别：3级节点;前置节点：标准项目布局;后置节点:避免循环导入;节点描述:单一职责，内聚性，导出规则，避免过深的包层级;
		节点名称：避免循环导入;节点级别：3级节点;前置节点：包设计原则;后置节点:代码规范与格式化;节点描述:循环导入的检测和解决（接口、共同依赖包）;
		节点名称：代码规范与格式化;节点级别：2级节点;前置节点：项目结构;后置节点:文档生成;节点描述:使用gofmt格式化代码，go vet静态检查，golint等工具，命名规范;
		节点名称：gofmt与go vet;节点级别：3级节点;前置节点：代码规范与格式化;后置节点:命名规范;节点描述:使用gofmt统一代码风格，go vet静态检查;
		节点名称：命名规范;节点级别：3级节点;前置节点：gofmt与go vet;后置节点:文档生成;节点描述:变量、函数、类型、包命名约定（驼峰，首字母大小写）;
		节点名称：文档生成;节点级别：2级节点;前置节点：代码规范与格式化;后置节点:高级特性;节点描述:godoc工具，注释规范（包注释、导出元素注释），生成文档;
		节点名称：godoc与注释;节点级别：3级节点;前置节点：文档生成;后置节点:高级特性;节点描述:注释规范（包注释、导出元素），使用go doc命令，在线文档;
		节点名称：高级特性;节点级别：1级节点;前置节点：工程实践;后置节点:反射;节点描述:Go语言的高级主题，如反射、不安全操作、CGo、汇编等;
		节点名称：反射;节点级别：2级节点;前置节点：高级特性;接口;后置节点:不安全操作;节点描述:reflect包，Type和Value，反射定律，动态调用和修改;
		节点名称：reflect.Type与Value;节点级别：3级节点;前置节点：反射;后置节点:反射定律;节点描述:reflect.TypeOf和reflect.ValueOf，Kind，类型与值分离;
		节点名称：反射定律;节点级别：3级节点;前置节点：reflect.Type与Value;后置节点:动态调用;节点描述:反射的三大定律，从Value获取类型，可设置性;
		节点名称：动态调用与修改;节点级别：3级节点;前置节点：反射定律;后置节点:不安全操作;节点描述:通过反射调用方法，设置结构体字段（可导出），创建实例;
		节点名称：不安全操作;节点级别：2级节点;前置节点：反射;后置节点:CGo;节点描述:unsafe包，Pointer操作，绕过类型系统，大小和偏移量;
		节点名称：unsafe.Sizeof与Alignof;节点级别：3级节点;前置节点：不安全操作;后置节点:Pointer操作;节点描述:unsafe.Sizeof、Alignof、Offsetof，计算内存大小和对齐;
		节点名称：Pointer操作;节点级别：3级节点;前置节点：unsafe.Sizeof与Alignof;后置节点:CGo;节点描述:unsafe.Pointer转换规则，绕过类型系统，使用场景;
		节点名称：CGo;节点级别：2级节点;前置节点：不安全操作;后置节点:汇编;节点描述:cgo工具，在Go中调用C代码，类型转换，构建标签;
		节点名称：CGo基础;节点级别：3级节点;前置节点：CGo;后置节点:类型转换;节点描述:导入"C"，在Go中调用C函数，构建标签cgo;
		节点名称：类型转换;节点级别：3级节点;前置节点：CGo基础;后置节点:汇编;节点描述:C与Go类型对应，字符串、切片等转换，内存管理;
		节点名称：汇编;节点级别：2级节点;前置节点：CGo;后置节点:无;节点描述:Go汇编语言基础，编写汇编函数，与Go交互;
		节点名称：汇编函数;节点级别：3级节点;前置节点：汇编;后置节点:无;节点描述:编写.s文件，Go调用汇编，TEXT指令，栈帧;
	`
	lines := strings.Split(strings.TrimSpace(rawData), "\n")

	type NodeInfo struct {
		Name        string
		Level       string
		Predecessor string
		Description string
		Category    string // This will be inferred from the Level 1 parent
	}

	var parsedNodes []NodeInfo
	for _, line := range lines {
		parts := strings.Split(line, ";")
		if len(parts) < 5 {
			continue
		}

		kvExtract := func(part string) string {
			// Standardize the colon to handle both English and Chinese versions
			standardizedPart := strings.ReplaceAll(part, ":", "：")
			kv := strings.SplitN(standardizedPart, "：", 2)
			if len(kv) == 2 {
				return strings.TrimSpace(kv[1])
			}
			return ""
		}

		node := NodeInfo{
			Name:        kvExtract(parts[0]),
			Level:       kvExtract(parts[1]),
			Predecessor: kvExtract(parts[2]),
			Description: kvExtract(parts[4]),
		}

		if node.Name == "" || node.Level == "" {
			log.Printf("Skipping malformed line: %s", line)
			continue
		}
		parsedNodes = append(parsedNodes, node)
	}

	// 4. Create all points and store in a map for relationship building
	pointMap := make(map[string]*model.KnowledgePoint)
	level1CategoryMap := make(map[string]string)
	for _, n := range parsedNodes {
		if n.Level == "1级节点" {
			level1CategoryMap[n.Name] = n.Name
		}
	}

	// Function to find the ultimate level 1 ancestor
	var findCategory func(string) string
	findCategory = func(nodeName string) string {
		for _, n := range parsedNodes {
			if n.Name == nodeName {
				if n.Level == "1级节点" {
					return n.Name
				}
				if n.Predecessor == "无" || n.Predecessor == "" {
					// If a non-level-1 node has no predecessor, it might be a root of another tree.
					// We need a better way to categorize it. For now, let's check if it's a category itself.
					if _, ok := catMap[n.Name]; ok {
						return n.Name
					}
					return ""
				}
				return findCategory(n.Predecessor)
			}
		}
		return "" // Should not happen
	}

	for _, n := range parsedNodes {
		categoryName := findCategory(n.Name)
		catID, ok := catMap[categoryName]
		if !ok {
			// Fallback for nodes that might not link to our main categories
			if c, ok := catMap[n.Name]; ok {
				catID = c
			} else {
				log.Printf("Warning: Could not find category for node '%s'. Falling back.", n.Name)
				catID = catMap["Go语言基础"] // Default fallback
			}
		}

		level, _ := strconv.Atoi(string(n.Level[0]))

		point := &model.KnowledgePoint{
			Name:        n.Name,
			CategoryID:  catID,
			Description: n.Description,
			Level:       level,
		}
		if err := tx.Create(point).Error; err != nil {
			log.Printf("Failed to create knowledge point %s: %v", n.Name, err)
			tx.Rollback()
			return
		}
		pointMap[n.Name] = point
	}

	// 5. Iterate again to set parent relationships
	for _, n := range parsedNodes {
		if n.Predecessor != "无" && n.Predecessor != "" {
			child, childOk := pointMap[n.Name]
			parent, parentOk := pointMap[n.Predecessor]
			if childOk && parentOk {
				child.ParentID = &parent.ID
				if err := tx.Save(child).Error; err != nil {
					log.Printf("Failed to set parent for %s: %v", n.Name, err)
					tx.Rollback()
					return
				}
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("failed to commit transaction: %v", err)
		tx.Rollback()
	} else {
		log.Println("Successfully seeded new knowledge graph data.")
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
