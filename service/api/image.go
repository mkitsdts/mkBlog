package api

import (
	"mkBlog/models"
	"mkBlog/pkg"
	"mkBlog/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func AddImage(c *gin.Context) {
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

func DeleteArticle(c *gin.Context) {
	title := strings.TrimSpace(c.Param("title"))
	if title == "" {
		c.JSON(400, gin.H{"msg": "invalid title"})
		return
	}

	if err := pkg.GetDatabase().Where("title = ?", title).Delete(&models.ArticleDetail{}).Error; err != nil {
		c.JSON(500, gin.H{"msg": "server error"})
		return
	}

	if err := pkg.GetDatabase().Where("title = ?", title).Delete(&models.ArticleSummary{}).Error; err != nil {
		c.JSON(500, gin.H{"msg": "server error"})
		return
	}

	c.JSON(200, gin.H{"msg": "successfully deleted article"})
}
