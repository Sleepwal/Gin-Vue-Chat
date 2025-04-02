# 聊天应用后端

这是基于Golang的即时聊天应用后端，提供API接口和WebSocket服务。

功能
- [ x ] 用户认证系统, 支持用户注册、登录和
- MongoDB存储聊天消息、用户、好友关系和群组信息
- WebSocket实现实时消息推送
- 完整的API接口（用户、好友、群组、消息）

## 技术栈

- Golang
- Gin框架
- Gorm
- MongoDB
- WebSocket

## 开发指南

```bash
go mod tidy
go run main.go
```

