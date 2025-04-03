# 聊天应用后端

这是基于Golang的即时聊天应用后端，提供API接口和WebSocket服务。

功能
- [ x ] 用户认证系统，注册、登录
- [ x ] MongoDB存储聊天消息、用户、好友关系和群组信息

好友：
- [ x ] 添加好友
- [ x ] 发送消息给好友

群组：
- [ x ] 添加群组
- [ x ] 邀请好友加入群组
- [ x ] 在群组里发送消息

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

