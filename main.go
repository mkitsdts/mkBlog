package main

import (
	"mkBlog/config"
	"mkBlog/pkg"
	"mkBlog/service"
)

func main() {

	if err := config.LoadConfig(); err != nil {
		panic("failed to load config: " + err.Error())
	}
	if err := pkg.InitDatabase(); err != nil {
		panic("failed to connect to database: " + err.Error())
	}
	if err := pkg.InitRouter(); err != nil {
		panic("failed to create router: " + err.Error())
	}

	if s, err := service.NewBlogService(); err == nil {
		s.Start()
	}

}
