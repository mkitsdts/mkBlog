package middleware

import (
	"mkBlog/models"
	"mkBlog/pkg/database"

	"github.com/gin-gonic/gin"
)

var blacklistedIPs = make(map[string]struct{})

func Init() {
	// 从数据库加载
	if database.GetDatabase() == nil {
		return
	}
	count := make(map[string]int)
	var blackips []models.SuspectedIP
	if err := database.GetDatabase().Find(&blackips).Error; err == nil {
		for _, b := range blackips {
			count[b.IP]++
			if count[b.IP] > 5 { // 超过5次记录则加入黑名单
				AddToBlacklist(b.IP, "Multiple suspicious activities")
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

func AddToBlacklist(ip string, reason string) {
	blacklistedIPs[ip] = struct{}{}
	database.GetDatabase().Create(&models.BlackIP{
		IP:     ip,
		Reason: reason,
	})
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
