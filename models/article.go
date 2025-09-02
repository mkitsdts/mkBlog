package models

type ArticleSummary struct {
	Title    string `json:"title" gorm:"primaryKey"`
	UpdateAt string `json:"updateAt"`
	Category string `json:"category"`
	Summary  string `json:"summary"`
}

type ArticleDetail struct {
	Title    string `json:"title" gorm:"primaryKey"`
	CreateAt string `json:"createAt"`
	UpdateAt string `json:"updateAt"`
	Author   string `json:"author"`
	Content  string `json:"content"`
}
