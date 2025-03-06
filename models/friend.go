package models

type FriendApplyment struct {
	Name 			string `json:"name" gorm:"primaryKey"`
	Email 			string `json:"email"`
	Website 		string `json:"website"`
	ApplymentTime 	string `json:"applymentTime"`
}

type Friend struct {
	Name 			string `json:"name" gorm:"primaryKey"`
	Email 			string `json:"email"`
	Website 		string `json:"website"`
	ApplymentTime 	string `json:"applymentTime"`
	AgreeTime 		string `json:"agreeTime"`
}