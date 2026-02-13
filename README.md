# GoCodeMentor - 智能编程教学辅助系统

GoCodeMentor 是一款基于 Go 语言和 AI 技术的编程教学辅助平台，旨在通过 AI 赋能教师教学与学生学习，提供作业生成、自动批改、智能答疑及反馈管理等功能。

## 🌟 核心功能

### 👨‍🏫 教师端 (Teacher)
- **班级管理**：创建班级、管理学生名单、查看学生参与度。
- **智能作业生成**：基于 AI 技术，根据知识点和难度自动生成编程作业。
- **作业发布与批改**：支持向多个班级发布作业，系统自动批改并生成评分建议与反馈。
- **学习进度追踪**：实时监控学生作业提交情况及得分分布。
- **智能答疑管理**：查看并参与学生与 AI 助教的对话历史。

### 👨‍🎓 学生端 (Student)
- **加入班级**：通过邀请码加入教师创建的班级。
- **在线答题**：在网页端直接编写代码并提交作业。
- **AI 编程助手**：在答题过程中，可随时咨询 AI 助教获取思路引导（非直接代码答案）。
- **反馈提交**：向系统或教师提交意见与反馈。

### 🔐 管理员端 (Admin)
- **账号管理**：全系统账号概览，支持重置用户密码（默认重置为 `123456`）。
- **全量反馈监控**：监控全系统的用户反馈情况，辅助系统优化。

## 🛠️ 技术栈
- **后端**：Go (Gin 框架)
- **数据库**：PostgreSQL (配合 GORM)
- **AI 集成**：SiliconFlow API (兼容 OpenAI 接口)
- **前端**：HTML5, CSS3, JavaScript (原生, 无需打包)
- **配置管理**：Viper (支持 YAML)
- **工具库**：Excelize (Excel 导入导出), Go-QRCode (二维码生成)

## 🚀 快速开始

### 1. 环境准备
- 安装 [Go](https://golang.org/dl/) (建议 1.20+)
- 安装 [PostgreSQL](https://www.postgresql.org/download/)

### 2. 配置数据库
在 `configs/sql_config.yaml` 中配置您的数据库连接信息：
```yaml
database:
  host: "localhost"
  port: 5432
  user: "your_user"
  password: "your_password"
  dbname: "gocodementor"
  sslmode: "disable"
```

### 3. 运行项目
```bash
# 安装依赖
go mod tidy

# 启动服务器
go run cmd/server/main.go
```
访问地址：`http://localhost:8081`

## 📁 目录结构
- `cmd/server`: 项目入口
- `internal/handler`: 请求处理器 (Controller 层)
- `internal/service`: 业务逻辑层
- `internal/repository`: 数据库访问层 (DAO 层)
- `internal/model`: 数据模型定义
- `web/templates`: 前端 HTML 模板
- `web/static`: 静态资源 (CSS, JS)

## 📝 许可证
本项目采用 MIT 许可证。
