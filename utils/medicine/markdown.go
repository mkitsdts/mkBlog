package medicine

import (
	"bufio"
	"fmt"
	"mkBlog/models"
	"os"
	"strings"
	"time"
)

// 解析markdown文件
func ParseMarkdown(filename string, info os.FileInfo) (models.ArticleSummary,models.ArticleDetail) {
	file, err := os.Open(filename)
	if(err != nil){
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var article models.ArticleSummary
	var detail models.ArticleDetail
	fmt.Println("开始解析")
	var count int8 = 0
	for scanner.Scan(){
		text := scanner.Text()
		if count >= 2 || strings.Contains(text,"---") {
			if count >= 2 {
				detail.Content += scanner.Text() + "\n"
			}
			count++
		}else if strings.Contains(text,"title:") {
			article.Title = text[7:]
			detail.Title = text[7:]
			fmt.Println("title: ",article.Title)
		}else if strings.Contains(text,"created_time:"){
			detail.CreateAt = text[13:]
			fmt.Println("createAt: ",detail.CreateAt)
		}else if strings.Contains(text,"tags:"){
			article.Tags = text[6:]
			fmt.Println("Tags: ",article.Tags)
		}else if strings.Contains(text,"category:"){
			article.Category = text[10:]
			fmt.Println("Category: ",article.Category)
		}else if strings.Contains(text,"author:"){
			detail.Author = text[8:]
			fmt.Println("Author: ",detail.Author)
		}
	}
	article.UpdateAt = info.ModTime().Format("2006-01-02 15:04:05")
	detail.UpdateAt = info.ModTime().Format("2006-01-02 15:04:05")
	if err = scanner.Err(); err != nil {
		panic(err)
	}
	
	if(len(detail.Content) < 72){
		article.Summary = detail.Content
	}else{
		article.Summary = detail.Content[:72] + "..."
	}

	return article,detail
}

// 创建文章
func CreateArticle(title string) {
	// 生成文章模板
	os.Mkdir("resource/"+title,os.ModePerm)
	filepath := "resource/" + title + ".md"
	file, err := os.Create(filepath)
	if(err != nil){
		if(os.IsExist(err)){
			fmt.Println("文件已存在")
			return
		}
	}
	defer file.Close()
	content := "---\ntitle: " + title  + "\ncreated_time: " + time.Now().Format("2006-01-02 15:04:05") + 
	"\ntags: \ncategory: \nauthor: \n---\n"
	file.WriteString(content)
}