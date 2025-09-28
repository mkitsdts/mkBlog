package middleware

import (
	"mkBlog/models"
	"mkBlog/pkg/database"

	"github.com/gin-gonic/gin"
)

var blacklistedIPs = make(map[string]struct{})

func init() {
	// 从数据库加载
	var blackips []models.BlackIP
	if err := database.GetDatabase().Find(&blackips).Error; err == nil {
		for _, b := range blackips {
			blacklistedIPs[b.IP] = struct{}{}
		}
	}
}

func Blacklist() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if _, exists := blacklistedIPs[ip]; exists {
			c.AbortWithStatusJSON(403, gin.H{"error": "Your IP is blacklisted"})
			return
		}
		c.Next()
	}
}
