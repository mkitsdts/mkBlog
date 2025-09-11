package service

import (
	"log/slog"
	"mkBlog/models"
	"mkBlog/pkg"
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
	var articles []models.ArticleSummary
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.Query("pageSize"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	single := strings.TrimSpace(c.Query("category"))
	multiParam := strings.TrimSpace(c.Query("categories"))
	var filters []string
	if multiParam != "" {
		parts := strings.Split(multiParam, ",")
		for _, p := range parts {
			t := strings.TrimSpace(p)
			if t != "" {
				filters = append(filters, t)
			}
		}
	} else if single != "" {
		filters = append(filters, single)
	}

	query := s.DB.Model(&models.ArticleSummary{})
	if len(filters) == 1 {
		query = query.Where("category = ?", filters[0])
	} else if len(filters) > 1 {
		query = query.Where("category IN ?", filters)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(500, gin.H{"msg": "server error"})
		return
	}
	if total == 0 {
		c.JSON(200, gin.H{"articles": []models.ArticleSummary{}, "total": 0, "page": page, "maxPage": 0, "categories": filters})
		return
	}
	if (int64(page)-1)*int64(pageSize) >= total {
		c.JSON(404, gin.H{"msg": "no more articles"})
		return
	}
	if err := query.Order("update_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&articles).Error; err != nil {
		c.JSON(500, gin.H{"msg": "server error"})
		return
	}
	c.JSON(200, gin.H{
		"articles":   articles,
		"total":      total,
		"page":       page,
		"maxPage":    (total + int64(pageSize) - 1) / int64(pageSize),
		"categories": filters,
	})
}

// 获取所有分类（去重）
func (s *BlogService) GetCategories(c *gin.Context) {
	var categories []string
	if err := s.DB.Model(&models.ArticleSummary{}).Distinct().Pluck("category", &categories).Error; err != nil {
		c.JSON(500, gin.H{"msg": "server error"})
		return
	}
	c.JSON(200, gin.H{"categories": categories})
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
	result := s.DB.Create(&friend)
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
	c.JSON(200, friends)
}

func (s *BlogService) AddArticle(c *gin.Context) {
	var article Article
	if err := c.BindJSON(&article); err != nil {
		c.JSON(400, gin.H{"msg": "invalid request body"})
		return
	}
	if article.Title == "" || article.Content == "" {
		c.JSON(400, gin.H{"msg": "invalid article data"})
		return
	}

	artd := models.ArticleDetail{
		Title:   article.Title,
		Content: article.Content,
		Author:  article.Author,
	}

	summary := ""
	if len(article.Content) < 100 {
		summary = article.Content
	} else {
		summary = article.Content[:100]
	}

	arts := models.ArticleSummary{
		Title:    article.Title,
		Category: article.Category,
		Summary:  summary,
		UpdateAt: article.UpdateAt,
	}

	if result := s.DB.Create(&artd); result.Error != nil {
		c.JSON(500, gin.H{"msg": "server error"})
		return
	}

	if result := s.DB.Create(&arts); result.Error != nil {
		c.JSON(500, gin.H{"msg": "server error"})
		return
	}
	c.JSON(200, gin.H{"msg": "successfully added article"})
}

func (s *BlogService) AddImage(c *gin.Context) {
	img := &models.Image{}
	if err := c.BindJSON(img); err != nil {
		c.JSON(400, gin.H{"msg": "invalid request body"})
		return
	}
	if img.Title == "" || img.Name == "" || len(img.Data) == 0 {
		c.JSON(400, gin.H{"msg": "invalid image data"})
		return
	}
	if err := pkg.SaveImage(img); err != nil {
		c.JSON(500, gin.H{"msg": "failed to save image"})
		return
	}
	c.JSON(200, gin.H{"msg": "successfully added image"})
}

func (s *BlogService) DeleteArticle(c *gin.Context) {
	type Title struct {
		Title string `json:"title"`
	}
	var title Title
	if err := c.BindJSON(&title); err != nil {
		c.JSON(400, gin.H{"msg": "invalid request body"})
		return
	}

	if err := s.DB.Where("title = ?", title.Title).Delete(&models.ArticleDetail{}).Error; err != nil {
		c.JSON(500, gin.H{"msg": "server error"})
		return
	}

	if err := s.DB.Where("title = ?", title.Title).Delete(&models.ArticleSummary{}).Error; err != nil {
		c.JSON(500, gin.H{"msg": "server error"})
		return
	}

	c.JSON(200, gin.H{"msg": "successfully deleted article"})
}
