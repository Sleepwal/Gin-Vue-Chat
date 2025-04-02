package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/yourusername/gin-vue-chat/config"
)

// JWTAuth 是JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 尝试从URL参数获取token（用于WebSocket连接）
			token := c.Query("token")
			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证令牌"})
				c.Abort()
				return
			}
			authHeader = "Bearer " + token
		}

		// 检查token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证格式无效"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
			}
			return []byte(config.AppConfig.JWT.Secret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌: " + err.Error()})
			c.Abort()
			return
		}

		// 验证token
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// 设置用户ID到上下文
			userID, ok := claims["userId"].(string)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户ID"})
				c.Abort()
				return
			}

			c.Set("userId", userID)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			c.Abort()
			return
		}
	}
}
