package api

import (
	"log/slog"
	"mkBlog/models"
	"mkBlog/pkg/bloom"
	"mkBlog/pkg/database"
	"mkBlog/utils"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Article struct {
	Author   string `json:"author"`
	Title    string `json:"title"`
	UpdateAt string `json:"update_at"`
	Category string `json:"category"`
	Content  string `json:"content"`
}

func UploadArticle(c *gin.Context) {
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
	// 生成摘要时去除 Markdown 语法
	summary = utils.StripMarkdown(summary)

	arts := models.ArticleSummary{
		Title:    article.Title,
		Category: article.Category,
		Summary:  summary,
	}

	if err := database.GetDatabase().Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "title"}},                        // 冲突字段
			DoUpdates: clause.AssignmentColumns([]string{"content", "author"}), // 更新的字段
		},
	).Create(&artd).Error; err != nil {
		slog.Error("insert article detail failed ", "error: ", err)
	}

	if err := database.GetDatabase().Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "title"}},                          // 冲突字段
			DoUpdates: clause.AssignmentColumns([]string{"summary", "category"}), // 更新的字段
		},
	).Create(&arts).Error; err != nil {
		slog.Error("insert article summary failed ", "error: ", err)
	}

	bloom.GetBloomFilter().Add([]byte(article.Title))
	c.JSON(200, gin.H{"msg": "successfully added article"})
}

// 文章详情
func GetArticleDetail(c *gin.Context) {
	title := strings.TrimSpace(c.Param("title"))
	if title == "" {
		c.JSON(400, gin.H{"msg": "invalid title"})
		return
	}

	// 先检查布隆过滤器，快速判断文章是否存在，减少数据库压力
	if !bloom.GetBloomFilter().Exists([]byte(title)) {
		c.JSON(404, gin.H{"msg": "article not found"})
		return
	}

	var article models.ArticleDetail
	safeTitle := path.Clean(title) // 防止路径遍历攻击
	result := database.GetDatabase().Where("title = ?", safeTitle).First(&article)
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

func GetArticleSummary(c *gin.Context) {
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
		parts := strings.SplitSeq(multiParam, ",")
		for p := range parts {
			t := strings.TrimSpace(p)
			if t != "" {
				filters = append(filters, t)
			}
		}
	} else if single != "" {
		filters = append(filters, single)
	}

	query := database.GetDatabase().Model(&models.ArticleSummary{})
	// 关键词搜索（title/summary 模糊匹配）
	q := strings.TrimSpace(c.Query("q"))
	if q != "" {
		like := "%" + q + "%"
		query = query.Where("title LIKE ? OR summary LIKE ?", like, like)
	}
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

// SearchArticle 使用 MySQL FULLTEXT 在正文 Content 上进行全文搜索
// 路由: GET /api/search?q=keyword&page=1&pageSize=10
// 返回值结构与 GetArticleSummary 保持一致: { articles, total, page, maxPage }
func SearchArticle(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(400, gin.H{"msg": "missing query q"})
		return
	}
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.Query("pageSize"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	// 统计匹配总数（以详情表为准，Title 唯一）
	var total int64
	if err := database.GetDatabase().Raw(countSQL, q).Scan(&total).Error; err != nil {
		c.JSON(500, gin.H{"msg": "server error"})
		return
	}
	// FULLTEXT 在中文环境常因最小分词长度或停用词导致 0 结果；若查询包含 CJK 则回退 LIKE
	usedLike := false
	if total == 0 && utils.ContainsCJK(q) {
		likeTerm := "%" + q + "%"
		if err := database.GetDatabase().Raw(countLikeSQL, likeTerm).Scan(&total).Error; err != nil {
			c.JSON(500, gin.H{"msg": "server error"})
			return
		}
		usedLike = true
	}
	if total == 0 {
		c.JSON(200, gin.H{"articles": []models.ArticleSummary{}, "total": 0, "page": page, "maxPage": 0})
		return
	}
	if (int64(page)-1)*int64(pageSize) >= total {
		c.JSON(404, gin.H{"msg": "no more articles"})
		return
	}

	// 选出摘要列表
	// FULLTEXT 模式按相关性排；LIKE 模式按更新时间排序
	var articles []models.ArticleSummary
	if usedLike {
		likeTerm := "%" + q + "%"
		if err := database.GetDatabase().Raw(listLikeSQL, likeTerm, pageSize, (page-1)*pageSize).Scan(&articles).Error; err != nil {
			c.JSON(500, gin.H{"msg": "server error"})
			return
		}
	} else if err := database.GetDatabase().Raw(listSQL, q, q, pageSize, (page-1)*pageSize).Scan(&articles).Error; err != nil {
		c.JSON(500, gin.H{"msg": "server error"})
		return
	}

	c.JSON(200, gin.H{
		"articles": articles,
		"total":    total,
		"page":     page,
		"maxPage":  (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

func DeleteArticle(c *gin.Context) {
	title := strings.TrimSpace(c.Param("title"))
	if title == "" {
		c.JSON(400, gin.H{"msg": "invalid title"})
		return
	}

	// 先检查布隆过滤器，快速判断文章是否存在，减少数据库压力
	if !bloom.GetBloomFilter().Exists([]byte(title)) {
		c.JSON(404, gin.H{"msg": "article not found"})
		return
	}

	for i := range 3 {
		if err := database.GetDatabase().Where("title = ?", title).Delete(&models.ArticleDetail{}).Error; err == nil {
			if err := database.GetDatabase().Where("title = ?", title).Delete(&models.ArticleSummary{}).Error; err == nil {
				break
			}
		}
		if i == 2 {
			c.JSON(500, gin.H{"msg": "server error"})
			return
		}
		time.Sleep(10 << i * time.Millisecond)
	}
	bloom.GetBloomFilter().Remove([]byte(title))
	c.JSON(200, gin.H{"msg": "successfully deleted article"})
}
