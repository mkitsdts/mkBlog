package main

import (
	"mkBlog/pkg"
	"mkBlog/service"
)

func main() {
	if err := pkg.InitRouter(); err != nil {
		panic("failed to create router: " + err.Error())
	}

	if s, err := service.NewBlogService(); err == nil {
		s.Start()
	}

}
