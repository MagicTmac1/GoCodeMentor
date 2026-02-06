package repository

import (
	"fmt"
	"log"
	"os"

	"GoCodeMentor/internal/model"

	"github.com/spf13/viper"
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
	)
	if err != nil {
		return nil, err
	}

	return db, nil
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
