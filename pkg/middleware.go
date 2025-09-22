package pkg

import (
	"net/http"
	"strings"
	"sync"
	"time"

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

var limiter map[string]int
var mux sync.Mutex

func init() {
	limiter = make(map[string]int)
}

// 限流中间件
func RateLimit(maxRequests int, windowSeconds int) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		mux.Lock()
		count := limiter[clientIP]
		limiter[clientIP] = count + 1
		mux.Unlock()
		// 在窗口期后减少计数
		go func() {
			time.AfterFunc(time.Duration(windowSeconds)*time.Second, func() {
				mux.Lock()
				limiter[clientIP]--
				if limiter[clientIP] <= 0 {
					delete(limiter, clientIP)
				}
				mux.Unlock()
			})
		}()
		if count >= maxRequests {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"msg": "too many requests"})
			return
		}
	}
}
