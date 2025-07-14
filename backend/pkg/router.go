package pkg

import "github.com/gin-gonic/gin"

func NewRouter() (*gin.Engine, error) {
	r := gin.Default()
	return r, nil
}
