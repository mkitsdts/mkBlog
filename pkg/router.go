package pkg

import "github.com/gin-gonic/gin"

func NewRouter() (*gin.Engine, error) {
	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1", "192.168.1.100"})
	return r, nil
}
