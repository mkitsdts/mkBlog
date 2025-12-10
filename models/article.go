package models

import "time"

type ArticleSummary struct {
	Title    string     `json:"title" gorm:"primaryKey"`
	UpdateAt *time.Time `json:"updateAt" gorm:"autoUpdateTime;"`
	Category string     `json:"category"`
	Summary  string     `json:"summary"`
}

type ArticleDetail struct {
	Title    string     `json:"title" gorm:"primaryKey"`
	CreateAt *time.Time `json:"createAt" gorm:"autoCreateTime;"`
	UpdateAt *time.Time `json:"updateAt" gorm:"autoUpdateTime;"`
	Author   string     `json:"author"`
	// 使用 GORM 创建 FULLTEXT 索引并指定 ngram 分词器（MySQL 5.7.6+/8.0）
	Content string `json:"content" gorm:"type:TEXT;index:ft_content`
}
