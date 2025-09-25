package api

import (
	"mkBlog/models"
	"mkBlog/utils"

	"github.com/gin-gonic/gin"
)

func UploadImage(c *gin.Context) {
	img := &models.Image{}
	if err := c.BindJSON(img); err != nil {
		c.JSON(400, gin.H{"msg": "invalid request body"})
		return
	}
	if img.Title == "" || img.Name == "" || len(img.Data) == 0 {
		c.JSON(400, gin.H{"msg": "invalid image data"})
		return
	}
	if err := utils.SaveImage(img); err != nil {
		c.JSON(500, gin.H{"msg": "failed to save image"})
		return
	}
	c.JSON(200, gin.H{"msg": "successfully added image"})
}
