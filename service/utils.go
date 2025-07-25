package service

import (
	"bufio"
	"fmt"
	"log/slog"
	"mkBlog/models"
	"os"
	"strconv"
	"strings"
	"time"
)

// 解析markdown文件
func (s *BlogService) ParseMarkdown(filename string, info os.FileInfo) (models.ArticleSummary, models.ArticleDetail) {
	file, err := os.Open(filename)
	if err != nil {
		slog.Error("failed to open file", "filename", filename, "error", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var article models.ArticleSummary
	var detail models.ArticleDetail
	fmt.Println("开始解析")
	var count int8 = 0
	for scanner.Scan() {
		text := scanner.Text()
		slog.Info("reading line", "line", text)
		if count >= 2 || strings.Contains(text, "---") {
			if count >= 2 {
				detail.Content += scanner.Text() + "\n"
			}
			count++
		} else if strings.Contains(text, "title:") {
			article.Title = strings.TrimSpace(text[6:])
			detail.Title = article.Title
			fmt.Println("title: ", article.Title)
		} else if strings.Contains(text, "created_time:") {
			createdTime, err := time.Parse("2006-01-02", strings.TrimSpace(text[13:]))
			if err != nil {
				slog.Error("failed to parse created_time", "value", strings.TrimSpace(text[13:]), "error", err)
			} else {
				detail.CreateAt = strconv.FormatInt(createdTime.Unix(), 10)
			}
		} else if strings.Contains(text, "tags:") {
			article.Tags = strings.TrimSpace(text[5:])
		} else if strings.Contains(text, "category:") {
			article.Category = strings.TrimSpace(text[9:])
		} else if strings.Contains(text, "author:") {
			detail.Author = strings.TrimSpace(text[7:])
		}
	}
	if info != nil {
		modTime := info.ModTime()
		article.UpdateAt = strconv.FormatInt(modTime.Unix(), 10)
		detail.UpdateAt = strconv.FormatInt(modTime.Unix(), 10)
	}
	if err = scanner.Err(); err != nil {
		slog.Error("failed to read file", "filename", filename, "error", err)
		return models.ArticleSummary{}, models.ArticleDetail{}
	}

	if len(detail.Content) < 72 {
		article.Summary = detail.Content
	} else {
		runes := []rune(detail.Content)
		if len(runes) > 72 {
			article.Summary = string(runes[:72]) + "..."
		} else {
			article.Summary = string(runes)
		}
	}

	return article, detail
}

// 创建文章
func (s *BlogService) CreateArticle(title string) {
	// 生成文章模板
	os.Mkdir("resource/"+title, os.ModePerm)
	filepath := "resource/" + title + ".md"
	file, err := os.Create(filepath)
	if err != nil {
		if os.IsExist(err) {
			fmt.Println("文件已存在")
			return
		}
	}
	defer file.Close()
	content := "---\ntitle: " + title + "\ncreated_time: " + time.Now().Format("2006-01-02 15:04:05") +
		"\ntags: \ncategory: \nauthor: \n---\n"
	file.WriteString(content)
}
