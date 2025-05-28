package service

import (
	"log/slog"
	"mkBlog/models"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type BlogService struct {
	DB     *gorm.DB
	Router *gin.Engine
}

func InitBlogService() *BlogService {
	service := &BlogService{}
	dsn := "root:root@tcp(localhost:3306)/mysql?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	service.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil
	}
	// 自动迁移
	err = service.DB.AutoMigrate(&models.ArticleSummary{}, &models.ArticleDetail{})
	if err != nil {
		slog.Error("failed to migrate database")
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

	service.UpdateArticle()

	return service
}

func (s *BlogService) Run() {
	err := s.Router.Run(":8080")
	if err != nil {
		slog.Error("failed to run server")
	}
}

func (s *BlogService) UpdateArticle() {
	err := filepath.Walk("resource", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".md" {
			// 读取文件内容
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			// 解析文件名
			summary := models.ArticleSummary{}
			detail := models.ArticleDetail{}
			summary, detail = s.ParseMarkdown(path, info)
			// 插入数据库
			s.DB.Where("title = ?", summary.Title).FirstOrCreate(&summary)
			s.DB.Where("title = ?", detail.Title).FirstOrCreate(&detail)
			// 更新数据库
			s.DB.Model(&summary).Where("title = ?", summary.Title).Updates(models.ArticleSummary{
				UpdateAt: summary.UpdateAt,
				Category: summary.Category,
				Tags:     summary.Tags,
				Summary:  summary.Summary,
			})
			s.DB.Model(&detail).Where("title = ?", detail.Title).Updates(models.ArticleDetail{
				UpdateAt: detail.UpdateAt,
				CreateAt: detail.CreateAt,
				Author:   detail.Author,
				Content:  detail.Content,
			})
			slog.Info("update article", "title", summary.Title)
			slog.Info("update article", "title", detail.Title)
			// 关闭文件
			file.Close()
		}
		return nil
	},
	)
	if err != nil {
		slog.Error("failed to update article")
	}

}
