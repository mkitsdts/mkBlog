package models

import "time"

type SuspectedIP struct {
	ID        uint32     `json:"id" gorm:"primaryKey;autoIncrement"`
	IP        string     `json:"ip" gorm:"primaryKey"`
	CreatedAt *time.Time `json:"createdAt" gorm:"autoCreateTime;"`
	Reason    string     `json:"reason"`
}

type BlackIP struct {
	IP        string     `json:"ip" gorm:"primaryKey"`
	CreatedAt *time.Time `json:"createdAt" gorm:"autoCreateTime;"`
	Reason    string     `json:"reason"`
}
