package main

import (
	"mkBlog/config"
	"mkBlog/pkg"
	"mkBlog/service"
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

	s := service.NewBlogService(db, r, cfg)
	s.Run()
}
