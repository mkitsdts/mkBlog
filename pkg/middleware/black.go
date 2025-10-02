package middleware

import (
	"mkBlog/models"
	"mkBlog/pkg/database"

	"github.com/gin-gonic/gin"
)

var blacklistedIPs = make(map[string]struct{})

func init() {
	// 从数据库加载
	count := make(map[string]int)
	var blackips []models.SuspectedIP
	if err := database.GetDatabase().Find(&blackips).Error; err == nil {
		for _, b := range blackips {
			count[b.IP]++
			if count[b.IP] > 5 { // 超过5次记录则加入黑名单
				blacklistedIPs[b.IP] = struct{}{}
				// TODO: 记录黑名单变更
				database.GetDatabase().Create(&models.BlackIP{
					IP:     b.IP,
					Reason: "Exceeded suspected IP threshold",
				})
			}
		}
	}
	var blackips2 []models.BlackIP
	if err := database.GetDatabase().Find(&blackips2).Error; err == nil {
		for _, b := range blackips2 {
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
