package main

import (
	"flag"
	"mkBlog/config"
	"mkBlog/models"
	"mkBlog/pkg/bloom"
	"mkBlog/pkg/cache"
	"mkBlog/pkg/log"
	"mkBlog/pkg/middleware"
	"mkBlog/pkg/router"
	"mkBlog/service"
	"os"
)

func Init(debugflag *bool) {
	if err := os.MkdirAll(models.Default_Data_Path, 0755); err != nil {
		return
	}
	if debugflag != nil {
		log.Init(*debugflag)
	} else {
		log.Init(false)
	}

	config.Init()

	bloom.Init()
	cache.Init(models.Default_Static_File_Path)
	middleware.Init()
}

func main() {
	debug := flag.Bool("debug", false, "启用调试模式")

	// 解析命令行参数
	flag.Parse()
	Init(debug)

	if err := router.InitRouter(); err != nil {
		panic("failed to create router: " + err.Error())
	}

	if s, err := service.NewBlogService(); err == nil {
		s.Start()
	}

}
