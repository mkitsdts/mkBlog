package models

import "time"

type Friend struct {
	Name     string     `json:"name" gorm:"type:varchar(255);primaryKey"`
	URL      string     `json:"url" gorm:"not null;type:varchar(255)"`
	Avatar   string     `json:"avatar"`
	Desc     string     `json:"desc"`
	CreateAt *time.Time `json:"createAt" gorm:"autoCreateTime;local"`
}
