package api

import (
	"log/slog"
	"mkBlog/models"
	"mkBlog/pkg/database"
	"mkBlog/utils"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var comment_count map[string]int
var mtx sync.Mutex

func GetCommentCount(title string) int {
	if count, ok := comment_count[title]; ok {
		return count
	}
	return 0
}

type Comment struct {
	CommentUser string `json:"comment_user"`
	CommentTo   int    `json:"comment_to"`
	Title       string `json:"title"`
	Content     string `json:"content"`
}

func AddComment(c *gin.Context) {
	var comment Comment
	if err := c.BindJSON(&comment); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	mtx.Lock()
	comment_count[comment.Title]++
	mtx.Unlock()
	if !utils.IsLegalComment(comment.Content) {
		c.JSON(400, gin.H{"msg": "illegal comment"})
		return
	}
	for i := range 3 {
		if result := database.GetDatabase().Create(&models.Comment{
			Content:     comment.Content,
			CommentUser: comment.CommentUser,
			CommentTo:   comment.CommentTo,
			Title:       comment.Title,
			Order:       comment_count[comment.Title],
		}); result.Error == nil {
			break
		} else if i == 2 {
			c.JSON(500, gin.H{"msg": "server error"})
			return
		}
		time.Sleep(10 << i * time.Millisecond)
	}
	c.JSON(200, gin.H{"msg": "comment added"})
}

func GetComments(c *gin.Context) {
	title := c.Query("title")
	if title == "" {
		slog.Warn("missing title parameter in GetComments")
		c.JSON(400, gin.H{"msg": "missing title parameter"})
		return
	}
	var comments []models.Comment
	for i := range 3 {
		if result := database.GetDatabase().Where("title = ?", title).Order("`order` ASC").Find(&comments); result.Error == nil {
			break
		} else if i == 2 {
			slog.Warn("failed to fetch comments", "title", title, "error", result.Error)
			c.JSON(500, gin.H{"msg": "server error"})
			return
		}
	}
	c.JSON(200, gin.H{"comments": comments})
}
