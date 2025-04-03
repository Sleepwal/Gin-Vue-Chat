package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/gin-vue-chat/models"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

// UpdateProfileRequest 更新用户资料请求
type UpdateProfileRequest struct {
	Email  string `json:"email" binding:"omitempty,email"`
	Avatar string `json:"avatar"`
	Status string `json:"status" binding:"omitempty,oneof=online offline away"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// GetUserProfile 获取用户资料
func GetUserProfile(c *gin.Context) {
	userID := c.GetString("userId")

	user, err := models.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       user.ID.Hex(),
			"username": user.Username,
			"email":    user.Email,
			"avatar":   user.Avatar,
			"status":   user.Status,
		},
	})
}

// UpdateUserProfile 更新用户资料
func UpdateUserProfile(c *gin.Context) {
	userID := c.GetString("userId")

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	user, err := models.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 更新字段
	if req.Email != "" && req.Email != user.Email {
		// 检查邮箱是否已被其他用户使用
		collection := models.MongoDatabase.Collection("users")
		var existingUser models.User
		err = collection.FindOne(context.Background(), bson.M{
			"email":   req.Email,
			"_id":     bson.M{"$ne": user.ID},
			"deleted": false,
		}).Decode(&existingUser)

		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "邮箱已被使用"})
			return
		}

		user.Email = req.Email
	}

	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	if req.Status != "" {
		user.Status = req.Status
	}

	err = models.UpdateUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户资料失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户资料已更新",
		"user": gin.H{
			"id":       user.ID.Hex(),
			"username": user.Username,
			"email":    user.Email,
			"avatar":   user.Avatar,
			"status":   user.Status,
		},
	})
}

// ChangePassword 修改密码
func ChangePassword(c *gin.Context) {
	userID := c.GetString("userId")

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	user, err := models.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 验证旧密码
	if !user.CheckPassword(req.OldPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "旧密码不正确"})
		return
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 更新密码
	user.Password = string(hashedPassword)
	err = models.UpdateUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新密码失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码已更新"})
}
