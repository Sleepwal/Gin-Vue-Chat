# 聊天应用后端

这是基于Golang的即时聊天应用后端，提供API接口和WebSocket服务。

功能
- [ x ] 用户认证系统，注册、登录
- [ x ] MongoDB存储聊天消息、用户、好友关系和群组信息
- [ x ] WebSocket实现实时消息推送
- [ x ] 添加好友、群组

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

