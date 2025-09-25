package api

import (
	"log/slog"
	"mkBlog/models"
	"mkBlog/pkg/database"

	"github.com/gin-gonic/gin"
)

func ApplyFriend(c *gin.Context) {
	var friend models.Friend
	if err := c.BindJSON(&friend); err != nil {
		c.JSON(400, gin.H{"msg": "invalid request body"})
		return
	}
	slog.Info("applying to be friends", "name", friend.Name, "url", friend.URL)
	if friend.Name == "" || friend.URL == "" {
		c.JSON(400, gin.H{"msg": "invalid friend data"})
		return
	}
	result := database.GetDatabase().Create(&friend)
	if result.Error != nil {
		c.JSON(500, gin.H{"msg": "server error"})
		return
	}
	c.JSON(200, gin.H{"msg": "successfully applied to be friends"})
}

func GetFriendList(c *gin.Context) {
	var friends []models.Friend
	result := database.GetDatabase().Find(&friends)
	if result.Error != nil {
		c.JSON(500, gin.H{"msg": "server error"})
		return
	}
	c.JSON(200, friends)
}
