package service

import (
	"log/slog"
	"mkBlog/config"
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
		api.PUT("/article/:title", service.AddArticle)
	}

	time.Sleep(2 * time.Second)
	return service
}

func (s *BlogService) Run() {
	err := s.Router.Run(":8080")
	if err != nil {
		slog.Error("failed to run server")
	}
}
