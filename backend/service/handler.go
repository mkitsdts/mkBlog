package service

import (
	"mkBlog/models"
	"path"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 文章详情
func (s *BlogService) GetArticleDetail(c *gin.Context) {
	title := strings.TrimSpace(c.Param("title"))
	if title == "" {
		c.JSON(400, gin.H{"msg": "请输入文章名"})
		return
	}
	var article models.ArticleDetail
	safeTitle := path.Clean(title) // 防止路径遍历攻击
	result := s.DB.Where("title = ?", safeTitle).First(&article)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"msg": "文章不存在"})
			return
		}
		c.JSON(500, gin.H{"msg": "服务器错误"})
		return
	}
	c.JSON(200, article)
}

func (s *BlogService) GetArticleSummary(c *gin.Context) {
	var articels []models.ArticleSummary
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page < 1 {
		page = 1
	}
	result := s.DB.Find(&articels)
	if result.Error != nil {
		c.JSON(500, gin.H{"msg": "服务器错误"})
		return
	}
	if len(articels) == 0 {
		c.JSON(404, gin.H{"msg": "没有更多文章"})
		return
	}
	c.JSON(200, gin.H{
		"articles": articels[(page-1)*10 : page*10-1],
		"page":     page,
		"maxPage":  (len(articels) + 9) / 10, // 向上取整
		"message":  "获取文章列表成功",
	})
}
