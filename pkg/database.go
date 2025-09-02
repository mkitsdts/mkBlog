package pkg

import (
	"log/slog"
	"mkBlog/config"
	"mkBlog/models"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	dsn := cfg.MySQL.User + ":" + cfg.MySQL.Password +
		"@tcp(" + cfg.MySQL.Host + ":" + cfg.MySQL.Port + ")/" +
		cfg.MySQL.Name + "?charset=utf8mb4&parseTime=True&loc=Local"
	// 等待数据库启动
	retryTimes := 100
	for i := range retryTimes {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			slog.Info("connected to database", "dsn", dsn)
			break
		}
		time.Sleep(time.Duration(i<<2) * time.Microsecond) // 等待100毫秒后重试
		if i == retryTimes-1 {
			return nil, err
		}
	}
	// 自动迁移
	err = db.AutoMigrate(&models.ArticleSummary{},
		&models.ArticleDetail{},
		&models.Friend{})
	if err != nil {
		return nil, err
	}
	slog.Info("database migration completed")

	// 初始化默认 Hello World 文章（仅在空库时）
	var count int64
	if err := db.Model(&models.ArticleSummary{}).Count(&count).Error; err == nil && count == 0 {
		summary := models.ArticleSummary{
			Title:    "Hello World",
			UpdateAt: time.Now().Format("2006-01-02 15:04:05"),
			Category: "General",
			Summary:  "欢迎使用博客，您可以删除这篇文章或编辑它。",
		}
		detail := models.ArticleDetail{
			Title:    summary.Title,
			CreateAt: summary.UpdateAt,
			UpdateAt: summary.UpdateAt,
			Author:   "system",
			Content:  "# Hello World\n\n欢迎使用博客，您可以删除这篇文章或编辑它。",
		}
		if err := db.Create(&summary).Error; err != nil {
			slog.Warn("failed to insert default article summary", "error", err)
		} else if err := db.Create(&detail).Error; err != nil {
			slog.Warn("failed to insert default article detail", "error", err)
		} else {
			slog.Info("inserted default Hello World article")
		}
	}
	return db, nil
}
