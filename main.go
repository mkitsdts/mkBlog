package main

import (
	"mkBlog/internal/routers"
	"mkBlog/internal/database"
	"mkBlog/internal/service"
)

func main() {
	database.InitDatabase()
	service.UpdateArticle()
	routers.InitRouter()
	routers.Router.Run(":8080")
}
