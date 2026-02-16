package database

import (
	"log/slog"
	"mkBlog/config"
	"mkBlog/models"
	"sync"
	"time"

	"gorm.io/gorm"
)

var db *gorm.DB
var once sync.Once

func GetDatabase() *gorm.DB {
	once.Do(func() {
		if err := Init(); err != nil {
			slog.Error("failed to initialize database", "error", err)
		}
	})
	if db == nil {
		return nil
	}
	return db
}

func Init() error {
	dsn := getDSN()
	if dsn == "" {
		panic("unsupported database kind: " + config.Cfg.Database.Kind)
	}
	// 等待数据库启动
	retryTimes := 100
	for i := range retryTimes {
		tmpDB, err := openDatabase(dsn)
		if err == nil {
			db = tmpDB
			slog.Info("connected to database", "dsn", dsn)
			break
		}
		time.Sleep(time.Duration(i<<2) * time.Microsecond) // 指数退避
		slog.Warn("failed to connect database", "dsn", dsn)
		if i == retryTimes-1 {
			return err
		}
	}
	// 自动迁移
	err := db.AutoMigrate(&models.ArticleSummary{},
		&models.ArticleDetail{},
		&models.Friend{},
		&models.Comment{},
		&models.BlackIP{},
		&models.SuspectedIP{},
	)
	if err != nil {
		return err
	}
	slog.Info("database migration completed")

	// 初始化默认 Hello World 文章（仅在空库时）
	var count int64
	if err := db.Model(&models.ArticleSummary{}).Count(&count).Error; err == nil && count == 0 {
		summary := models.ArticleSummary{
			Title:    "Hello World",
			Category: "General",
			Summary:  "欢迎使用博客，您可以删除这篇文章或编辑它。",
		}
		detail := models.ArticleDetail{
			Title:   summary.Title,
			Author:  "system",
			Content: "# Hello World\n\n欢迎使用博客，您可以删除这篇文章或编辑它。",
		}
		if err := db.Create(&summary).Error; err != nil {
			slog.Warn("failed to insert default article summary", "error", err)
		} else if err := db.Create(&detail).Error; err != nil {
			slog.Warn("failed to insert default article detail", "error", err)
		} else {
			slog.Info("inserted default Hello World article")
		}
	}

	// FULLTEXT 索引由 GORM tag 自动处理（models.ArticleDetail.Content 上的 index:ft_content,class:FULLTEXT,option:WITH PARSER ngram）
	createFullTextIndex()
	return nil
}
