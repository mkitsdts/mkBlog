package service

import (
	"mkBlog/models"
	"path"
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
	var articel models.ArticleSummary
	title := strings.TrimSpace(c.Query("title"))
	if err := s.DB.Where("title = ?", title).First(&articel).Error; err != nil {
		c.JSON(400, gin.H{
			"status": "error",
			"msg":    err.Error(),
		})
		return
	}
	c.JSON(200, articel)
}
