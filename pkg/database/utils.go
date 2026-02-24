package database

import (
	"fmt"
	"log/slog"
	"mkBlog/config"
	"mkBlog/models"
	"path"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func getDSN() string {
	switch strings.ToLower(config.Cfg.Database.Kind) {
	case models.MySQL:
		return config.Cfg.Database.User + ":" + config.Cfg.Database.Password +
			"@tcp(" + config.Cfg.Database.Host + ":" + config.Cfg.Database.Port + ")/" +
			config.Cfg.Database.Name + "?charset=utf8mb4&parseTime=True&loc=UTC"
	case models.Postgres:
		return "host=" + config.Cfg.Database.Host +
			" port=" + config.Cfg.Database.Port +
			" user=" + config.Cfg.Database.User +
			" password=" + config.Cfg.Database.Password +
			" dbname=" + config.Cfg.Database.Name +
			" sslmode=disable TimeZone=UTC"
	case models.SQLite3:
		return path.Join(models.Default_Data_Path, config.Cfg.Database.Host)
	}
	return path.Join(models.Default_Data_Path, models.Default_Data_File_Path)
}

func openDatabase(dsn string) (*gorm.DB, error) {
	switch strings.ToLower(config.Cfg.Database.Kind) {
	case models.MySQL:
		return gorm.Open(mysql.Open(dsn), &gorm.Config{
			NowFunc: func() time.Time { return time.Now().UTC() },
		})
	case models.Postgres:
		return gorm.Open(postgres.Open(dsn), &gorm.Config{
			NowFunc: func() time.Time { return time.Now().UTC() },
		})
	case models.SQLite3:
		return gorm.Open(sqlite.Open(dsn), &gorm.Config{
			NowFunc: func() time.Time { return time.Now().UTC() },
		})
	}
	return nil, fmt.Errorf("unsupported database kind: %s", config.Cfg.Database.Kind)
}

func createFullTextIndex() {
	switch strings.ToLower(config.Cfg.Database.Kind) {
	case models.MySQL:
		db.Exec(createMySQLFullTextIndexSQL)
	case models.Postgres:
		if res := db.Exec(usePostgresExtensionSQL); res.Error != nil {
			slog.Error("failed to start zhparser extension. please ensure it is installed", "error", res.Error)
		}
		slog.Info("start zhparser extension successfully")
		db.Exec(createPostgresChineseDictionarySQL)
		db.Exec(createPostgresDictionaryMappingSQL)
		db.Exec(createPostgresFullTextIndexSQL)
	case models.SQLite3:
		return
	}
}
