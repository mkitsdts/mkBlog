package service

import (
	"log/slog"
	"mkBlog/models"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BlogService struct {
	DB     *gorm.DB
	Router *gin.Engine
}

func NewBlogService(db *gorm.DB, r *gin.Engine) *BlogService {
	service := &BlogService{}
	service.DB = db
	service.Router = r
	service.Router.GET("/", service.GetArticleSummary)
	service.Router.GET("/article/:title", service.GetArticleDetail)
	service.Router.GET("/friend", service.GetFriendList)
	service.Router.POST("/apply", service.ApplyFriend)
	service.UpdateArticle()
	time.Sleep(2 * time.Second) // 等待文件更新完成
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
			slog.Info("processing file", "path", path)
			// 读取文件内容
			file, err := os.Open(path)
			if err != nil {
				slog.Error("failed to open file", "path", path, "error", err)
				return nil
			}
			defer file.Close()

			// 解析文件内容
			summary, detail := s.ParseMarkdown(path, info)
			slog.Info("updating article", "title", summary.Title)

			// 使用事务处理数据库操作
			tx := s.DB.Begin()
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
				}
			}()

			// 处理 ArticleSummary
			var existingSummary models.ArticleSummary
			result := tx.Where("title = ?", summary.Title).First(&existingSummary)
			if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
				slog.Error("failed to query summary", "title", summary.Title, "error", result.Error)
				tx.Rollback()
				return nil
			}

			if result.Error == gorm.ErrRecordNotFound {
				// 创建新记录
				if err := tx.Create(&summary).Error; err != nil {
					slog.Error("failed to create summary", "title", summary.Title, "error", err)
					tx.Rollback()
					return nil
				}
			} else {
				// 更新现有记录
				if err := tx.Model(&existingSummary).Updates(summary).Error; err != nil {
					slog.Error("failed to update summary", "title", summary.Title, "error", err)
					tx.Rollback()
					return nil
				}
			}

			// 处理 ArticleDetail
			var existingDetail models.ArticleDetail
			result = tx.Where("title = ?", detail.Title).First(&existingDetail)
			if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
				slog.Error("failed to query detail", "title", detail.Title, "error", result.Error)
				tx.Rollback()
				return nil
			}

			if result.Error == gorm.ErrRecordNotFound {
				// 创建新记录
				if err := tx.Create(&detail).Error; err != nil {
					slog.Error("failed to create detail", "title", detail.Title, "error", err)
					tx.Rollback()
					return nil
				}
			} else {
				// 更新现有记录
				if err := tx.Model(&existingDetail).Updates(detail).Error; err != nil {
					slog.Error("failed to update detail", "title", detail.Title, "error", err)
					tx.Rollback()
					return nil
				}
			}

			// 提交事务
			if err := tx.Commit().Error; err != nil {
				slog.Error("failed to commit transaction", "title", summary.Title, "error", err)
				return nil
			}

			slog.Info("successfully updated article", "title", summary.Title)
			return nil
		}
		slog.Info("processed file", "path", path)
		return nil
	})
	if err != nil {
		slog.Error("failed to update articles", "error", err)
	}
}
