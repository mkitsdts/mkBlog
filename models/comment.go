package models

import "time"

type Comment struct {
	ID          uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	Content     string     `json:"content" gorm:"type:text"`
	CommentUser string     `json:"comment_user" gorm:"type:varchar(100)"`
	CommentTo   int        `json:"comment_to_order" gorm:"type:int"`
	Title       string     `json:"title" gorm:"index"` // 关联文章标题
	Order       int        `json:"order" gorm:"default:0"`
	CreatedAt   *time.Time `json:"created_at" gorm:"autoCreateTime;"`
}
