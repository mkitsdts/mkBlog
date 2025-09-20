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
	// 使用 GORM 创建 FULLTEXT 索引并指定 ngram 分词器（MySQL 5.7.6+/8.0）
	// 索引名与后续原生 SQL 检测保持一致：ft_content
	Content string `json:"content" gorm:"type:LONGTEXT;index:ft_content,class:FULLTEXT,option:WITH PARSER ngram"`
}
