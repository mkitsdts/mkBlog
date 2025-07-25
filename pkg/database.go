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
	return db, nil
}
