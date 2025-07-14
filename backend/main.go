package main

import (
	"mkBlog/config"
	"mkBlog/pkg"
	"mkBlog/service"
	"os"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
	db, err := pkg.NewDatabase(cfg)
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}
	r, err := pkg.NewRouter()
	if err != nil {
		panic("failed to create router: " + err.Error())
	}

	s := service.NewBlogService(db, r)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "create":
			if len(os.Args) < 3 {
				panic("请输入文章名")
			}
			var title string
			for i := 2; i < len(os.Args); i++ {
				title += os.Args[i] + " "
			}
			s.CreateArticle(title)
			return
		case "update":
			s.UpdateArticle()
			return
		}
	}
	s.Run()
}
