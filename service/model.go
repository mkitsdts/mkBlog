package service

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type BlogService struct {
	RedisClient *redis.Client
	Router      *gin.Engine
}
