package main

import (
	"mkBlog/config"
	"mkBlog/pkg/bloom"
	"mkBlog/pkg/cache"
	"mkBlog/pkg/log"
	"mkBlog/pkg/middleware"
	"mkBlog/pkg/router"
	tlscert "mkBlog/pkg/tls_cert"
	"mkBlog/service"
)

func Init() {
	log.Init()
	config.Init()
	tlscert.Init()
	bloom.Init()
	cache.Init("./static")
	middleware.Init()
}

func main() {
	Init()

	if err := router.InitRouter(); err != nil {
		panic("failed to create router: " + err.Error())
	}

	if s, err := service.NewBlogService(); err == nil {
		s.Start()
	}

}
