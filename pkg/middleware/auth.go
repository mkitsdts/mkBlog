package middleware

import (
	"net/http"
	"strings"

	"mkBlog/config"

	"github.com/gin-gonic/gin"
)

// 简单的 Bearer Token 鉴权（用于后台写接口）
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 可通过 config.Auth.Enabled 开关控制是否启用鉴权
		auth := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(auth, prefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "missing bearer token"})
			return
		}
		token := strings.TrimPrefix(auth, prefix)
		if token != config.Cfg.Auth.Secret {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"msg": "invalid token"})
			return
		}
		c.Next()
	}
}
