package service

import (
	"fmt"
	"mime"
	"mkBlog/models"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
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
	result := s.RedisClient.Get(c, "article:"+safeTitle).Scan(&article)
	if result != nil {
		c.JSON(404, gin.H{"msg": "文章不存在"})
		return
	}
	// slog()
	c.JSON(200, article)
}

func (s *BlogService) ImageHandler(c *gin.Context) {
	// 安全参数
	title := strings.TrimSpace(c.Param("title"))
	imagePath := strings.TrimSpace(c.Param("path"))

	if title == "" || imagePath == "" {
		c.JSON(400, gin.H{"error": "invalid parameters"})
		return
	}

	safeTitle := path.Clean(title)     // 防止路径遍历攻击
	safeImage := path.Clean(imagePath) // 如过滤 ../ 等字符

	filePath := filepath.Join("resource", safeTitle, safeImage)
	fmt.Println(filePath)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(404, gin.H{"error": "image not found"})
		fmt.Println("image not found")
		return
	}

	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream" // 默认类型
	}

	c.File(filePath)
}

func (s *BlogService) GetArticleSummary(c *gin.Context) {
	result := ""
	err := s.RedisClient.Get(c, "article_summary:").Scan(&result)
	if err != nil {
		c.JSON(400, gin.H{
			"status": "error",
			"msg":    err.Error(),
		})
		return
	}
	c.JSON(200, result)
}

// 申请友链
func (s *BlogService) ApplyFriend(c *gin.Context) {
	var friendApplyment models.FriendApplyment
	c.BindJSON(&friendApplyment)
	result := s.RedisClient.Set(c, "friend_applyment:"+friendApplyment.Email, friendApplyment, 0)
	if result.Err() != nil {
		c.JSON(400, gin.H{
			"status": "error",
			"msg":    result.Err(),
		})
		return
	}
	c.JSON(200, gin.H{
		"status": "success",
	})
}

// 获取友链
func (s *BlogService) GetFriendList(c *gin.Context) {
	var friends []models.Friend
	result := s.RedisClient.Get(c, "friend").Scan(&friends)
	if result.Error != nil {
		c.JSON(400, gin.H{
			"status": "error",
			"msg":    result.Error,
		})
		return
	}
	c.JSON(200, friends)
}
