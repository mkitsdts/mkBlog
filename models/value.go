package models

type HomeInfo struct{
	Articles 	[]ArticleSummary `json:"articles"`
	Count 		int64 `json:"count"`
	CurrentPage int `json:"currentPage"`
}