package config

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

// Config 应用配置
type Config struct {
	// 服务器配置
	Server struct {
		Port string
		Mode string // development, production
	}

	// 数据库配置
	Database struct {
		Type     string // sqlite, mysql, postgres
		Path     string // SQLite数据库文件路径
		Host     string
		Port     string
		User     string
		Password string
		Name     string
	}

	// MongoDB配置
	MongoDB struct {
		URI      string
		Database string
	}

	// JWT配置
	JWT struct {
		Secret    string
		ExpireDur time.Duration
	}

	// 跨域配置
	CORS struct {
		AllowOrigins []string
	}
}

// AppConfig 全局配置实例
var AppConfig Config

// InitConfig 初始化配置
func InitConfig() {
	// 设置默认配置
	setDefaultConfig()

	// 从环境变量加载配置
	loadFromEnv()

	// 确保数据目录存在
	ensureDataDir()

	log.Println("配置初始化完成")
}

// 设置默认配置
func setDefaultConfig() {
	// 服务器配置
	AppConfig.Server.Port = "8080"
	AppConfig.Server.Mode = "development"

	// 数据库配置 - 默认使用MySQL
	AppConfig.Database.Type = "mysql"
	AppConfig.Database.Host = "localhost"
	AppConfig.Database.Port = "3306"
	AppConfig.Database.User = "root"
	AppConfig.Database.Password = "password"
	AppConfig.Database.Name = "gin_vue_chat"

	// MongoDB配置
	AppConfig.MongoDB.URI = "mongodb://localhost:27017"
	AppConfig.MongoDB.Database = "chat"

	// JWT配置
	AppConfig.JWT.Secret = "your-secret-key-change-in-production"
	AppConfig.JWT.ExpireDur = 24 * time.Hour

	// CORS配置
	AppConfig.CORS.AllowOrigins = []string{"http://localhost:3000"}
}

// 从环境变量加载配置
func loadFromEnv() {
	// 服务器配置
	if port := os.Getenv("SERVER_PORT"); port != "" {
		AppConfig.Server.Port = port
	}
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		AppConfig.Server.Mode = mode
	}

	// 数据库配置
	if dbType := os.Getenv("DB_TYPE"); dbType != "" {
		AppConfig.Database.Type = dbType
	}
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		AppConfig.Database.Path = dbPath
	}
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		AppConfig.Database.Host = dbHost
	}
	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		AppConfig.Database.Port = dbPort
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		AppConfig.Database.User = dbUser
	}
	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		AppConfig.Database.Password = dbPassword
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		AppConfig.Database.Name = dbName
	}

	// MongoDB配置
	if mongoURI := os.Getenv("MONGO_URI"); mongoURI != "" {
		AppConfig.MongoDB.URI = mongoURI
	}
	if mongoDB := os.Getenv("MONGO_DB"); mongoDB != "" {
		AppConfig.MongoDB.Database = mongoDB
	}

	// JWT配置
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		AppConfig.JWT.Secret = jwtSecret
	}
}

// 确保数据目录存在
func ensureDataDir() {
	if AppConfig.Database.Type == "sqlite" {
		dataDir := filepath.Dir(AppConfig.Database.Path)
		if _, err := os.Stat(dataDir); os.IsNotExist(err) {
			os.MkdirAll(dataDir, 0755)
		}
	}
}
