package middleware

import (
	"mkBlog/models"
	"mkBlog/pkg/database"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

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
		limiter[clientIP]++
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
		if limiter[clientIP] >= maxRequests {
			go database.GetDatabase().Create(&models.SuspectedIP{
				IP:     clientIP,
				Reason: "Rate limit exceeded",
			})
			// 超过限流阈值，拒绝请求
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"msg": "too many requests"})
			return
		}
	}
}
