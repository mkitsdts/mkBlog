package models

type ArticleSummary struct{
	Title string `json:"title" gorm:"primaryKey"`
	CreateAt string `json:"createAt"`
	UpdateAt string `json:"updateAt"`
	Author string `json:"author"`
	Category string `json:"category"`
	Tags string `json:"tags"`
}

type ArticleDetail struct{
	Title string `json:"title" gorm:"primaryKey"`
	Content string `json:"content"`
}