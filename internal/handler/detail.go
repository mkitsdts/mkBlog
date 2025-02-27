package handler

import (
	"github.com/gin-gonic/gin"
	"mkBlog/models"
	"mkBlog/internal/database"
)

// 文章详情
func DetailHandler(c *gin.Context) {
	id := c.Param("id")
	var article models.ArticleDetail
	result := database.Db.First(&article, id)
	if result.Error != nil {
		c.JSON(404, gin.H{"msg": "文章不存在"})
		return
	}
	c.JSON(200, article)
}