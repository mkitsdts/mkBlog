package main

import (
	"mkBlog/service"
	"mkBlog/utils/medicine"
	"os"
)

func main() {
	service := service.InitBlogService()
	if len(os.Args) > 1 {
		if os.Args[1] == "create" {
			if len(os.Args) < 3 {
				panic("请输入文章名")
			}
			var title string
			for i := 2; i < len(os.Args); i++ {
				title += os.Args[i] + " "
			}
			medicine.CreateArticle(title)
			return
		} else if os.Args[1] == "update" {
			service.UpdateArticle()
			return
		}
	}
	service.Run()
}
