package medicine

import (
	"bufio"
	"fmt"
	"mkBlog/models"
	"os"
	"strings"
)

// 解析markdown文件
func ParseMarkdown(filename string) (models.ArticleSummary,models.ArticleDetail) {
	// 这里有很大优化空间
	file, err := os.Open(filename)
	if(err != nil){
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var article models.ArticleSummary
	fmt.Println("开始解析")
	var count int8 = 0
	for scanner.Scan(){
		text := scanner.Text()
		fmt.Println(text)
		if strings.Contains(text,"title:"){
			article.Title = text[7:]
			fmt.Println("title: ",article.Title)
		}else if strings.Contains(text,"created_time:"){
			article.CreateAt = text[13:]
			fmt.Println("createAt: ",article.CreateAt)
		}else if strings.Contains(text,"updated_time:"){
			article.UpdateAt = text[13:]
			fmt.Println("UpdateAt: ",article.UpdateAt)
		}else if strings.Contains(text,"tags:"){
			article.Tags = text[6:]
			fmt.Println("Tags: ",article.Tags)
		}else if strings.Contains(text,"category:"){
			article.Category = text[10:]
			fmt.Println("Category: ",article.Category)
		}else if strings.Contains(text,"author:"){
			article.Author = text[8:]
			fmt.Println("Author: ",article.Author)
		}else if strings.Contains(text,"--"){
			count++
			if(count == 2) {
				break
			}
		}
	}
	var detail models.ArticleDetail
	for scanner.Scan(){
		detail.Content += scanner.Text()
	}
	return article,detail
}