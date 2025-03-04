package handler

import (
	"github.com/gin-gonic/gin"
	"mkBlog/models"
	"mkBlog/internal/database"
	"strconv"
	"fmt"
)

// 主界面
func HomeHandler(c *gin.Context) {
	page , err:= strconv.Atoi(c.Query("page"))
	if err != nil {
		page = 1
	}
	offset := (page - 1) * models.MaxCountEachPage
	var articles []models.ArticleSummary
	database.Db.Limit(models.MaxCountEachPage).Offset(offset).Find(&articles)
	fmt.Println(articles)
	var info models.HomeInfo
	database.Db.Model(&models.ArticleSummary{}).Count(&info.MaxPage)
	info.MaxPage = (info.MaxPage + models.MaxCountEachPage - 1) / models.MaxCountEachPage
	info.Articles = articles
	info.CurrentPage = page
	c.JSON(200, info)
}