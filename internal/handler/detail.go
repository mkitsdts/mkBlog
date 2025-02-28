package handler

import (
	"github.com/gin-gonic/gin"
	"mkBlog/models"
	"mkBlog/internal/database"
	"strings"
	"fmt"
	"path"
	"path/filepath"
	"mime"
	"os"
)

// 文章详情
func DetailHandler(c *gin.Context) {
	title := c.Param("title")
	var article models.ArticleDetail
	result := database.Db.First(&article, title)
	if result.Error != nil {
		c.JSON(404, gin.H{"msg": "文章不存在"})
		return
	}
	fmt.Println(article)
	c.JSON(200, article)
}

func ImageHandler(c *gin.Context) {
	// 安全参数
    title := strings.TrimSpace(c.Param("title"))
    imagePath := strings.TrimSpace(c.Param("path"))

    if title == "" || imagePath == "" {
        c.JSON(400, gin.H{"error": "invalid parameters"})
        return
    }

    safeTitle := path.Clean(title)       // 防止路径遍历攻击
    safeImage := path.Clean(imagePath)   	// 如过滤 ../ 等字符
    
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