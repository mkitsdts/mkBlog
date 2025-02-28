package handler

import (
	"github.com/gin-gonic/gin"
	"mkBlog/models"
	"mkBlog/internal/database"
	"fmt"
)

// 搜索
func SearchHandler(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(400, gin.H{"msg": "请输入关键字"})
		return
	}
	var articles []models.ArticleSummary
	database.Db.Where("title LIKE ?", "%"+keyword+"%").Find(&articles)
	fmt.Println(articles)
	c.JSON(200, articles)
}