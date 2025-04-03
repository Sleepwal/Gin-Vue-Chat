package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/gin-vue-chat/models"
	"github.com/yourusername/gin-vue-chat/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SendPrivateMessageRequest 发送私聊消息请求
type SendPrivateMessageRequest struct {
	ReceiverID string `json:"receiverId" binding:"required"`
	Content    string `json:"content" binding:"required"`
}

// SendGroupMessageRequest 发送群聊消息请求
type SendGroupMessageRequest struct {
	GroupID string `json:"groupId" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// GetPrivateMessages 获取私聊消息
func GetPrivateMessages(c *gin.Context) {
	userID := c.GetString("userId")
	receiverID := c.Param("userId")

	// 检查接收者是否存在
	_, err := models.GetUserByID(receiverID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 检查是否是好友关系
	friendships, err := models.GetFriendships(userID, "accepted")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}

	// 验证是否为好友
	isFriend := false
	for _, friendship := range friendships {
		if (friendship.UserID == userID && friendship.FriendID == receiverID) ||
			(friendship.UserID == receiverID && friendship.FriendID == userID) {
			isFriend = true
			break
		}
	}

	// 如果不是好友，返回错误
	if !isFriend {
		c.JSON(http.StatusForbidden, gin.H{"error": "您不是该用户的好友"})
		return
	}

	// 获取分页参数
	limit := int64(20) // 默认每页20条
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil && l > 0 {
			limit = l
		}
	}

	skip := int64(0) // 默认从第一条开始
	if skipStr := c.Query("skip"); skipStr != "" {
		if s, err := strconv.ParseInt(skipStr, 10, 64); err == nil && s >= 0 {
			skip = s
		}
	}

	// 获取消息
	messages, err := models.GetPrivateMessages(userID, receiverID, limit, skip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取消息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// SendPrivateMessage 发送私聊消息
func SendPrivateMessage(c *gin.Context) {
	senderID := c.GetString("userId")

	var req SendPrivateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 检查接收者是否存在
	_, err := models.GetUserByID(req.ReceiverID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "接收者不存在"})
		return
	}

	// 检查是否是好友关系
	friendships, err := models.GetFriendships(senderID, "accepted")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}

	// 验证是否为好友
	isFriend := false
	for _, friendship := range friendships {
		if (friendship.UserID == senderID && friendship.FriendID == req.ReceiverID) ||
			(friendship.UserID == req.ReceiverID && friendship.FriendID == senderID) {
			isFriend = true
			break
		}
	}

	if !isFriend {
		c.JSON(http.StatusForbidden, gin.H{"error": "您不是该用户的好友"})
		return
	}

	// 保存消息到MongoDB
	message, err := models.SavePrivateMessage(senderID, req.ReceiverID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存消息失败"})
		return
	}

	// 获取发送者信息
	sender, _ := models.GetUserByID(senderID)

	// 通过WebSocket发送消息给接收者
	wsMessage := map[string]interface{}{
		"type": "private",
		"message": map[string]interface{}{
			"id":        message.ID.Hex(),
			"from":      senderID,
			"to":        req.ReceiverID,
			"content":   req.Content,
			"timestamp": message.Timestamp,
			"sender": map[string]interface{}{
				"id":       sender.ID,
				"username": sender.Username,
				"avatar":   sender.Avatar,
			},
		},
	}

	// 获取WebSocket Hub
	hub := c.MustGet("wsHub").(*websocket.Hub)

	// 发送消息给接收者
	// 将消息转换为JSON字符串，再转换为字节数组
	jsonData, err := json.Marshal(gin.H{"data": wsMessage})
	if err != nil {
		// 记录错误但继续执行，因为这不是致命错误
		log.Printf("消息序列化失败: %v", err)
		return
	}
	hub.SendToUser(req.ReceiverID, jsonData)

	c.JSON(http.StatusOK, gin.H{
		"message": "消息发送成功",
		"data":    message,
	})
}

// GetGroupMessages 获取群聊消息
func GetGroupMessages(c *gin.Context) {
	userID := c.GetString("userId")
	groupID := c.Param("groupId")

	// 检查群组是否存在
	_, err := models.GetGroupByID(groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "群组不存在"})
		return
	}

	// 检查用户是否是群组成员
	// 获取用户所在的群组
	userGroups, err := models.GetUserGroups(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}

	// 验证用户是否在群组中
	isMember := false
	for _, g := range userGroups {
		if g.ID.Hex() == groupID {
			isMember = true
			break
		}
	}

	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "您不是该群组的成员"})
		return
	}

	// 获取分页参数
	limit := int64(20) // 默认每页20条
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil && l > 0 {
			limit = l
		}
	}

	skip := int64(0) // 默认从第一条开始
	if skipStr := c.Query("skip"); skipStr != "" {
		if s, err := strconv.ParseInt(skipStr, 10, 64); err == nil && s >= 0 {
			skip = s
		}
	}

	// 获取消息
	messages, err := models.GetGroupMessages(groupID, limit, skip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取消息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// SendGroupMessage 发送群聊消息
func SendGroupMessage(c *gin.Context) {
	senderID := c.GetString("userId")

	var req SendGroupMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 检查群组是否存在
	_, err := models.GetGroupByID(req.GroupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "群组不存在"})
		return
	}

	// 检查用户是否是群组成员
	// 获取用户所在的群组
	userGroups, err := models.GetUserGroups(senderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}

	// 验证用户是否在群组中
	isMember := false
	for _, g := range userGroups {
		if g.ID.Hex() == req.GroupID {
			isMember = true
			break
		}
	}

	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "您不是该群组的成员"})
		return
	}

	// 保存消息到MongoDB
	message, err := models.SaveGroupMessage(senderID, req.GroupID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存消息失败"})
		return
	}

	// 获取发送者信息
	sender, _ := models.GetUserByID(senderID)

	// 通过WebSocket发送消息给群组所有成员
	wsMessage := map[string]interface{}{
		"type": "group",
		"message": map[string]interface{}{
			"id":        message.ID.Hex(),
			"groupId":   req.GroupID,
			"senderId":  senderID,
			"content":   req.Content,
			"timestamp": message.Timestamp,
			"sender": map[string]interface{}{
				"id":       sender.ID,
				"username": sender.Username,
				"avatar":   sender.Avatar,
			},
		},
	}

	// 获取WebSocket Hub
	hub := c.MustGet("wsHub").(*websocket.Hub)

	// 获取群组所有成员
	members, err := models.GetGroupMembers(req.GroupID)
	if err != nil {
		log.Printf("获取群组成员失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发送消息失败"})
		return
	}

	// 发送消息给所有成员
	for _, member := range members {
		if member.UserID != senderID { // 不需要发送给自己
			jsonData, err := json.Marshal(gin.H{"data": wsMessage})
			if err != nil {
				log.Printf("消息序列化失败: %v", err)
				continue
			}
			hub.SendToUser(member.UserID, jsonData)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "消息发送成功",
		"data":    message,
	})
}

// MarkMessagesAsRead 标记消息为已读
func MarkMessagesAsRead(c *gin.Context) {
	// 获取消息ID列表
	var req struct {
		MessageIDs []string `json:"messageIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 转换字符串ID为ObjectID
	objectIDs := make([]primitive.ObjectID, 0, len(req.MessageIDs))
	for _, idStr := range req.MessageIDs {
		id, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			continue
		}
		objectIDs = append(objectIDs, id)
	}

	if len(objectIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的消息ID"})
		return
	}

	// 标记消息为已读
	err := models.MarkMessagesAsRead(objectIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "标记消息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "消息已标记为已读"})
}
