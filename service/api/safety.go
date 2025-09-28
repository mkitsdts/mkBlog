package api

import (
	"mkBlog/models"
	"mkBlog/pkg/database"

	"github.com/gin-gonic/gin"
)

type BlackIPReq struct {
	IP     string `json:"ip" binding:"required,ip"`
	Reason string `json:"reason" binding:"max=255"`
}

func BlackIP(c *gin.Context) {
	var req BlackIPReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := database.GetDatabase().Create(&models.BlackIP{
		IP:     req.IP,
		Reason: req.Reason,
	}).Error; err != nil {
		c.JSON(500, gin.H{"error": "failed to block IP"})
		return
	}
}
