package service

import (
	"log/slog"
	"mkBlog/config"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BlogService struct {
	DB     *gorm.DB
	Router *gin.Engine
	Cfg    *config.Config
}

func NewBlogService(db *gorm.DB, r *gin.Engine, cfg *config.Config) *BlogService {
	service := &BlogService{}
	service.DB = db
	service.Router = r
	service.Cfg = cfg

	api := r.Group("/api")
	{
		api.GET("/articles", service.GetArticleSummary)
		api.GET("/article/:title", service.GetArticleDetail)
		api.GET("/categories", service.GetCategories)
		api.GET("/friends", service.GetFriendList)
		api.POST("/friends", service.ApplyFriend)
		if cfg.Auth.Enabled {
			api.PUT("/article/:title", service.AddArticle, service.Auth())
		} else {
			api.PUT("/article/:title", service.AddArticle)
		}
	}
	time.Sleep(2 * time.Second)
	return service
}

func (s *BlogService) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.Request.Header.Get("Authorization")
		if auth == "Bearer "+s.Cfg.Auth.Secret {
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

func (s *BlogService) Run() {
	err := s.Router.Run(":8080")
	if err != nil {
		slog.Error("failed to run server")
	}
}
