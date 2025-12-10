package api

import (
	"log/slog"
	"mkBlog/config"
	"mkBlog/pkg/database"
)

const mysqlListLike = "SELECT s.title, s.update_at, s.category, s.summary FROM article_details d JOIN article_summaries s ON s.title = d.title WHERE MATCH(d.content,d.title) AGAINST (? IN NATURAL LANGUAGE MODE) LIMIT ? OFFSET ?"

const mysqlCountLike = "SELECT COUNT(*) FROM article_details d  WHERE MATCH(d.content,d.title) AGAINST (? IN NATURAL LANGUAGE MODE)"

const postgresListLike = "SELECT s.title, s.update_at, s.category, s.summary FROM article_details d JOIN article_summaries s ON s.title = d.title WHERE to_tsvector('zhparser', COALESCE(d.content,'') || ' ' || COALESCE(d.title,'')) @@ plainto_tsquery('zhparser', ?) ORDER BY s.update_at DESC LIMIT ? OFFSET ?"
const postgresCountLike = "SELECT COUNT(*) FROM article_details d  WHERE to_tsvector('zhparser', COALESCE(d.content,'') || ' ' || COALESCE(d.title,'')) @@ plainto_tsquery('zhparser', ?)"

// LIKE 回退（主要用于未启用倒排索引的检索场景）
const listLikeSQL = "SELECT s.title, s.update_at, s.category, s.summary FROM article_details d JOIN article_summaries s ON s.title = d.title WHERE d.content LIKE ? ORDER BY s.update_at DESC LIMIT ? OFFSET ?"
const countLikeSQL = "SELECT COUNT(*) FROM article_details d WHERE d.content LIKE ?"

var (
	countSQL string
	listSQL  string
)

func Init() {
	// count comment from database
	comment_count = make(map[string]int)
	var rows []struct {
		Title string
		Count int
	}
	database.GetDatabase().Table("article_details AS a").
		Select("a.title, COUNT(c.id) AS count").
		Joins("LEFT JOIN comments c ON a.title = c.title").
		Group("a.title").
		Scan(&rows)
	for _, row := range rows {
		comment_count[row.Title] = row.Count
	}
	switch config.Cfg.Database.Kind {
	case "mysql":
		listSQL = mysqlListLike
		countSQL = mysqlCountLike
	case "postgres":
		listSQL = postgresListLike
		countSQL = postgresCountLike
	default:
		listSQL = listLikeSQL
		countSQL = countLikeSQL
	}
	slog.Info("search SQL initialized",
		"countSQL", countSQL,
		"listSQL", listSQL,
		"countSQL_len", len(countSQL),
		"listSQL_len", len(listSQL))
}
