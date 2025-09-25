package main

import (
	"mkBlog/pkg/router"
	"mkBlog/service"
)

func main() {
	if err := router.InitRouter(); err != nil {
		panic("failed to create router: " + err.Error())
	}

	if s, err := service.NewBlogService(); err == nil {
		s.Start()
	}

}
