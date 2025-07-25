package service

import (
	"log/slog"
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
		c.JSON(400, gin.H{"msg": "invalid title"})
		return
	}
	var article models.ArticleDetail
	safeTitle := path.Clean(title) // 防止路径遍历攻击
	result := s.DB.Where("title = ?", safeTitle).First(&article)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"msg": "article not found"})
			return
		}
		c.JSON(500, gin.H{"msg": "server error"})
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
		c.JSON(500, gin.H{"msg": "server error"})
		return
	}
	if len(articels) == 0 {
		c.JSON(404, gin.H{"msg": "no more articles"})
		return
	}
	c.JSON(200, gin.H{
		"articles": articels[(page-1)*10 : page*10-1],
		"page":     page,
		"maxPage":  (len(articels) + 9) / 10, // 向上取整
		"message":  "successfully retrieved article summaries",
	})
}

func (s *BlogService) ApplyFriend(c *gin.Context) {
	var friend models.Friend
	if err := c.BindJSON(&friend); err != nil {
		c.JSON(400, gin.H{"msg": "invalid request body"})
		return
	}
	slog.Info("applying to be friends", "name", friend.Name, "url", friend.URL)
	if friend.Name == "" || friend.URL == "" {
		c.JSON(400, gin.H{"msg": "invalid friend data"})
		return
	}
	result := s.DB.FirstOrCreate(&friend)
	if result.Error != nil {
		c.JSON(500, gin.H{"msg": "server error"})
		return
	}
	c.JSON(200, gin.H{"msg": "successfully applied to be friends"})
}

func (s *BlogService) GetFriendList(c *gin.Context) {
	var friends []models.Friend
	result := s.DB.Find(&friends)
	if result.Error != nil {
		c.JSON(500, gin.H{"msg": "server error"})
		return
	}
	if len(friends) == 0 {
		c.JSON(200, gin.H{"msg": "no more friends"})
		return
	}
	c.JSON(200, gin.H{
		"friends": friends,
		"message": "successfully retrieved friend list",
	})
}
