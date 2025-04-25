package service

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func (s *BlogService) UpdateArticle() {
	err := filepath.Walk("resource", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".md" {

		}
		return nil
	},
	)
	if err != nil {
		slog.Error("failed to update article")
	}

}

func InitBlogService() *BlogService {
	service := &BlogService{}
	service.RedisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6789",
		DB:   0, // use default DB
	})

	_, err := service.RedisClient.Ping(context.Background()).Result()
	if err != nil {
		return nil
	}
	service.Router = gin.Default()
	service.Router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // 允许前端域名
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204) // 处理预检请求
			return
		}
		c.Next()
	})
	service.Router.GET("/", service.GetArticleSummary)
	service.Router.GET("/home", service.GetArticleSummary)
	service.Router.GET("/articles/:title", service.GetArticleDetail)
	service.Router.GET("/images/:title/:path", service.ImageHandler)
	service.Router.GET("/friend", service.GetFriendList)
	service.Router.POST("/friend/apply", service.ApplyFriend)
	return service
}

func (s *BlogService) Run() {
	err := s.Router.Run(":8080")
	if err != nil {
		slog.Error("failed to run server")
	}
}
