package handler

import (
	"github.com/gin-gonic/gin"
	"mkBlog/models"
	"mkBlog/internal/database"
	"fmt"
)

// 主界面
func HomeHandler(c *gin.Context) {
	var articles []models.ArticleSummary
	database.Db.Find(&articles)
	fmt.Println(articles)
	c.JSON(200, articles)
}