package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/gin-vue-chat/models"
)

// AddFriendRequest 添加好友请求
type AddFriendRequest struct {
	FriendId string `json:"friendId" binding:"required"`
}

// GetFriends 获取好友列表
func GetFriends(c *gin.Context) {
	userID := c.GetString("userId")

	// 查询已接受的好友关系
	friendships, err := models.GetFriendships(userID, "accepted")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取好友列表失败"})
		return
	}

	// 构建好友列表响应
	friends := make([]gin.H, 0)
	for _, friendship := range friendships {
		var friendID string
		if friendship.UserID == userID {
			friendID = friendship.FriendID
		} else {
			friendID = friendship.UserID
		}

		// 获取好友信息
		friend, err := models.GetUserByID(friendID)
		if err != nil {
			continue // 跳过无法获取的好友
		}

		friends = append(friends, gin.H{
			"id":       friend.ID.Hex(),
			"username": friend.Username,
			"avatar":   friend.Avatar,
			"status":   friend.Status,
		})
	}

	c.JSON(http.StatusOK, gin.H{"friends": friends})
}

// AddFriend 添加好友
func AddFriend(c *gin.Context) {
	userID := c.GetString("userId")

	var req AddFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 查找要添加的用户 - 前端发送的是用户名，不是ID
	friend, err := models.GetUserByUsername(req.FriendId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 不能添加自己为好友
	if friend.ID.Hex() == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能添加自己为好友"})
		return
	}

	// 添加好友关系
	friendship, err := models.AddFriend(userID, friend.ID.Hex())
	if err != nil {
		if err.Error() == "好友关系已存在" {
			c.JSON(http.StatusConflict, gin.H{"error": "已经是好友或好友请求已发送"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "添加好友失败: " + err.Error()})
		}
		return
	}

	// 简化处理，直接设为已接受
	friendship.Status = "accepted"
	err = models.UpdateFriendship(friendship)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新好友关系失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "好友添加成功",
		"friend": gin.H{
			"id":       friend.ID.Hex(),
			"username": friend.Username,
			"avatar":   friend.Avatar,
			"status":   friend.Status,
		},
	})
}

// RemoveFriend 删除好友
func RemoveFriend(c *gin.Context) {
	userID := c.GetString("userId")
	friendID := c.Param("id")

	// 获取所有好友关系
	friendships, err := models.GetFriendships(userID, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取好友关系失败"})
		return
	}

	// 查找与指定好友的关系
	var targetFriendship *models.Friendship
	for _, friendship := range friendships {
		if (friendship.UserID == userID && friendship.FriendID == friendID) ||
			(friendship.UserID == friendID && friendship.FriendID == userID) {
			targetFriendship = friendship
			break
		}
	}

	if targetFriendship == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "好友关系不存在"})
		return
	}

	// 删除好友关系
	err = models.DeleteFriendship(targetFriendship.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除好友失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "好友已删除"})
}
