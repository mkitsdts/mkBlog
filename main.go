package main

import (
	"mkBlog/service"
	"os"
)

func main() {
	s := service.InitBlogService()
	if len(os.Args) > 1 {
		if os.Args[1] == "create" {
			if len(os.Args) < 3 {
				panic("请输入文章名")
			}
			var title string
			for i := 2; i < len(os.Args); i++ {
				title += os.Args[i] + " "
			}
			s.CreateArticle(title)
			return
		} else if os.Args[1] == "update" {
			s.UpdateArticle()
			return
		}
	}
	s.Run()
}
