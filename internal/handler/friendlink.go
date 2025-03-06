package handler

import (
	"mkBlog/internal/database"
	"github.com/gin-gonic/gin"
	"mkBlog/models"
)

// 申请友链
func ApplyFriend(c *gin.Context) {
	var friendApplyment models.FriendApplyment
	c.BindJSON(&friendApplyment)
	result := database.ApplyFriend(friendApplyment)
	if result != nil {
		c.JSON(400, gin.H{
			"status": "error",
			"msg": result.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"status": "success",
	})
}

// 获取友链
func GetFriend(c *gin.Context) {
	var friends []models.Friend
	result := database.Db.Find(&friends)
	if result.Error != nil {
		c.JSON(400, gin.H{
			"status": "error",
			"msg": result.Error.Error(),
		})
		return
	}
	c.JSON(200, friends)
}
