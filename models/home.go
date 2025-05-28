package models

type HomeInfo struct{
	Articles 	[]ArticleSummary `json:"articles"`
	MaxPage 	int64 `json:"maxpage"`
	CurrentPage int `json:"currentPage"`
}