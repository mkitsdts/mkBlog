package models

import "time"

type ArticleSummary struct {
	Title    string     `json:"title" gorm:"primaryKey"`
	UpdateAt *time.Time `json:"updateAt"`
	Category string     `json:"category"`
	Tags     string     `json:"tags"`
	Summary  string     `json:"summary"`
}

type ArticleDetail struct {
	Title    string     `json:"title" gorm:"primaryKey"`
	CreateAt *time.Time `json:"createAt"`
	UpdateAt *time.Time `json:"updateAt"`
	Author   string     `json:"author" gorm:"type:varchar(255)"`
	Content  string     `json:"content"`
}
