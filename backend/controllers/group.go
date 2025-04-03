package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/gin-vue-chat/models"
)

// CreateGroupRequest 创建群组请求
type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
}

// UpdateGroupRequest 更新群组请求
type UpdateGroupRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
}

// AddGroupMemberRequest 添加群组成员请求
type AddGroupMemberRequest struct {
	Username string `json:"username" binding:"required"`
	Role     string `json:"role" binding:"omitempty,oneof=admin member"`
}

// GetGroups 获取用户的群组列表
func GetGroups(c *gin.Context) {
	userID := c.GetString("userId")

	// 获取用户所在的群组
	groups, err := models.GetUserGroups(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取群组列表失败"})
		return
	}

	// 获取用户在每个群组中的角色
	response := make([]gin.H, 0)
	for _, group := range groups {
		// 获取用户在群组中的角色
		members, err := models.GetGroupMembers(group.ID.Hex())
		if err != nil {
			continue
		}

		var role string = "member"
		for _, member := range members {
			if member.UserID == userID {
				role = member.Role
				break
			}
		}

		response = append(response, gin.H{
			"id":          group.ID.Hex(),
			"name":        group.Name,
			"description": group.Description,
			"avatar":      group.Avatar,
			"role":        role,
			"creatorId":   group.CreatorID,
		})
	}

	c.JSON(http.StatusOK, gin.H{"groups": response})
}

// CreateGroup 创建群组
func CreateGroup(c *gin.Context) {
	userID := c.GetString("userId")

	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 创建群组
	group, err := models.CreateGroup(req.Name, req.Description, req.Avatar, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建群组失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "群组创建成功",
		"group": gin.H{
			"id":          group.ID,
			"name":        group.Name,
			"description": group.Description,
			"avatar":      group.Avatar,
			"creatorId":   group.CreatorID,
		},
	})
}

// GetGroupDetail 获取群组详情
func GetGroupDetail(c *gin.Context) {
	userID := c.GetString("userId")
	groupID := c.Param("id")

	// 检查用户是否是群组成员
	members, err := models.GetGroupMembers(groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}

	// 查找当前用户的成员信息
	var membership *models.GroupMember
	for _, member := range members {
		if member.UserID == userID {
			membership = member
			break
		}
	}

	if membership == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "您不是该群组的成员"})
		return
	}

	// 获取群组信息
	group, err := models.GetGroupByID(groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "群组不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"group": gin.H{
			"id":          group.ID.Hex(),
			"name":        group.Name,
			"description": group.Description,
			"avatar":      group.Avatar,
			"creatorId":   group.CreatorID,
			"role":        membership.Role,
		},
	})
}

// UpdateGroup 更新群组信息
func UpdateGroup(c *gin.Context) {
	userID := c.GetString("userId")
	groupID := c.Param("id")

	var req UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 检查用户是否是群组管理员
	members, err := models.GetGroupMembers(groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}

	// 查找当前用户的成员信息
	var isAdmin bool
	for _, member := range members {
		if member.UserID == userID && member.Role == "admin" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "您不是该群组的管理员"})
		return
	}

	// 获取群组信息
	group, err := models.GetGroupByID(groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "群组不存在"})
		return
	}

	// 更新群组信息
	if req.Name != "" {
		group.Name = req.Name
	}
	group.Description = req.Description
	if req.Avatar != "" {
		group.Avatar = req.Avatar
	}

	err = models.UpdateGroup(group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新群组失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "群组更新成功",
		"group": gin.H{
			"id":          group.ID.Hex(),
			"name":        group.Name,
			"description": group.Description,
			"avatar":      group.Avatar,
			"creatorId":   group.CreatorID,
		},
	})
}

// DeleteGroup 删除群组
func DeleteGroup(c *gin.Context) {
	userID := c.GetString("userId")
	groupID := c.Param("id")

	// 检查用户是否是群组创建者
	group, err := models.GetGroupByID(groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "群组不存在"})
		return
	}

	if group.CreatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "您不是该群组的创建者，无法删除"})
		return
	}

	// 删除群组成员 (在MongoDB中是逻辑删除)
	members, err := models.GetGroupMembers(groupID)
	for _, member := range members {
		err = models.RemoveGroupMember(groupID, member.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除群组成员失败"})
			return
		}
	}

	// 删除群组 (在MongoDB中是逻辑删除)
	err = models.DeleteGroup(groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除群组失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "群组已删除"})
}

// GetGroupMembers 获取群组成员列表
func GetGroupMembers(c *gin.Context) {
	userID := c.GetString("userId")
	groupID := c.Param("id")

	// 检查用户是否是群组成员
	members, err := models.GetGroupMembers(groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取群组成员失败"})
		return
	}

	// 检查当前用户是否是群组成员
	var isMember bool
	for _, member := range members {
		if member.UserID == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "您不是该群组的成员"})
		return
	}

	// 构建成员列表响应
	memberList := make([]gin.H, 0)
	for _, member := range members {
		// 获取用户信息
		user, err := models.GetUserByID(member.UserID)
		if err != nil {
			continue // 跳过无法获取的用户
		}

		memberList = append(memberList, gin.H{
			"id":       user.ID.Hex(),
			"username": user.Username,
			"avatar":   user.Avatar,
			"status":   user.Status,
			"role":     member.Role,
		})
	}

	c.JSON(http.StatusOK, gin.H{"members": memberList})
}

// AddGroupMember 添加群组成员
func AddGroupMember(c *gin.Context) {
	userID := c.GetString("userId")
	groupID := c.Param("id")

	var req AddGroupMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 检查用户是否是群组管理员
	members, err := models.GetGroupMembers(groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}

	// 查找当前用户的成员信息
	var isAdmin bool
	for _, member := range members {
		if member.UserID == userID && member.Role == "admin" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "您不是该群组的管理员"})
		return
	}

	// 查找要添加的用户
	user, err := models.GetUserByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 设置角色，默认为普通成员
	role := "member"
	if req.Role != "" {
		role = req.Role
	}

	// 添加用户到群组
	newMember, err := models.AddGroupMember(groupID, user.ID.Hex(), role)
	if err != nil {
		if err.Error() == "用户已经是群组成员" {
			c.JSON(http.StatusConflict, gin.H{"error": "用户已经是群组成员"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "添加群组成员失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "成员添加成功",
		"member": gin.H{
			"id":       user.ID.Hex(),
			"username": user.Username,
			"avatar":   user.Avatar,
			"status":   user.Status,
			"role":     newMember.Role,
		},
	})
}

// RemoveGroupMember 移除群组成员
func RemoveGroupMember(c *gin.Context) {
	userID := c.GetString("userId")
	groupID := c.Param("id")
	memberID := c.Param("userId")

	// 获取群组信息
	group, err := models.GetGroupByID(groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "群组不存在"})
		return
	}

	// 获取群组成员列表
	members, err := models.GetGroupMembers(groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}

	// 检查操作者权限
	var currentUserMembership *models.GroupMember
	var targetMembership *models.GroupMember

	for _, member := range members {
		if member.UserID == userID {
			currentUserMembership = member
		}
		if member.UserID == memberID {
			targetMembership = member
		}
	}

	if currentUserMembership == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "您不是该群组的成员"})
		return
	}

	if targetMembership == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "该用户不是群组成员"})
		return
	}

	// 检查权限：只有管理员可以移除成员，或者用户自己退出
	if currentUserMembership.Role != "admin" && userID != memberID {
		c.JSON(http.StatusForbidden, gin.H{"error": "您没有权限移除该成员"})
		return
	}

	// 群组创建者不能被移除
	if memberID == group.CreatorID && userID != memberID {
		c.JSON(http.StatusForbidden, gin.H{"error": "不能移除群组创建者"})
		return
	}

	// 移除成员
	err = models.RemoveGroupMember(groupID, memberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "移除成员失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "成员已移除"})
}
