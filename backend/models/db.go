package models

import (
	"context"
	"log"
	"time"

	"github.com/yourusername/gin-vue-chat/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB 全局MongoDB客户端
var MongoDB *mongo.Client

// MongoDatabase MongoDB数据库实例
var MongoDatabase *mongo.Database

// InitDB 初始化数据库连接 - 现在只初始化MongoDB
func InitDB() {
	// 初始化MongoDB
	InitMongoDB()
}

// InitMongoDB 初始化MongoDB连接
func InitMongoDB() {
	// 创建MongoDB客户端
	clientOptions := options.Client().ApplyURI(config.AppConfig.MongoDB.URI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	MongoDB, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("无法连接到MongoDB: %v", err)
	}

	// 检查连接
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = MongoDB.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB连接检查失败: %v", err)
	}

	// 获取数据库实例
	MongoDatabase = MongoDB.Database(config.AppConfig.MongoDB.Database)

	log.Println("成功连接到MongoDB")
}
