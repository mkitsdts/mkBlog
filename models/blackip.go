package models

import "time"

type BlackIP struct {
	IP        string     `json:"ip" gorm:"primaryKey"`
	CreatedAt *time.Time `json:"createdAt" gorm:"autoCreateTime;"`
	Reason    string     `json:"reason"`
}
