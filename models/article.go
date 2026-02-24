package models

import "time"

type ArticleSummary struct {
	Title    string     `json:"title" gorm:"primaryKey"`
	UpdateAt *time.Time `json:"updateAt" gorm:"autoUpdateTime;"`
	Category string     `json:"category" gorm:"type:varchar(255)"`
	Summary  string     `json:"summary" gorm:"type:TEXT"`
}

type ArticleDetail struct {
	Title    string     `json:"title" gorm:"primaryKey"`
	CreateAt *time.Time `json:"createAt" gorm:"autoCreateTime;"`
	UpdateAt *time.Time `json:"updateAt" gorm:"autoUpdateTime;"`
	Author   string     `json:"author" gorm:"type:varchar(255)"`
	Content  string     `json:"content" gorm:"type:TEXT;index:ft_content,fulltext"`
}
