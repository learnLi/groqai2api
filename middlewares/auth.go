package middlewares

import (
	"github.com/gin-gonic/gin"
	"groqai2api/global"
	"strings"
)

func Authorization(c *gin.Context) {
	if global.Authorization != "" {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		// 支持官方apikey调用会跳过Authorization验证
		if global.SupportApikey != "true" && global.Authorization != strings.Replace(authHeader, "Bearer ", "", 1) {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
	}
	c.Next()
}

func AuthSecret(c *gin.Context) {
	if global.OpenAuthSecret != "true" {
		c.JSON(401, gin.H{"error": "未开放功能"})
		c.Abort()
		return
	}
	authHeader := c.GetHeader("Authorization")
	if global.AuthSecret != "" {
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		if global.AuthSecret != strings.Replace(authHeader, "Bearer ", "", 1) {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
	}
	c.Next()
}
