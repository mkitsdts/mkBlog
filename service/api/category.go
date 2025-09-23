package api

import (
	"mkBlog/models"
	"mkBlog/pkg"

	"github.com/gin-gonic/gin"
)

// 获取数据库中的所有分类
func GetCategories(c *gin.Context) {
	var categories []string
	if err := pkg.GetDatabase().Model(&models.ArticleSummary{}).Distinct().Pluck("category", &categories).Error; err != nil {
		c.JSON(500, gin.H{"msg": "server error"})
		return
	}
	c.JSON(200, gin.H{"categories": categories})
}
