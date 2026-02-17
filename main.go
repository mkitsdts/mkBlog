package main

import (
	"mkBlog/config"
	"mkBlog/models"
	"mkBlog/pkg/bloom"
	"mkBlog/pkg/cache"
	"mkBlog/pkg/log"
	"mkBlog/pkg/middleware"
	"mkBlog/pkg/router"
	"mkBlog/service"
	"os"
	"path/filepath"
)

func Init() {
	dir := filepath.Dir(models.Default_Config_File_Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}
	log.Init()
	config.Init()

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
