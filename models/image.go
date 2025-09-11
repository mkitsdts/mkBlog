package models

type Image struct {
	Title string `json:"title"`
	Data  []byte `json:"data"` // base64 encoded image data
	Name  string `json:"name"` // original file name
}
