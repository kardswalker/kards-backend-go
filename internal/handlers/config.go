package handlers

import "github.com/gin-gonic/gin"

func GetConfig(c *gin.Context) {
	c.JSON(200, gin.H{
		"xserver_closed":        "Server in Go!\n服务器已切换至 Go 引擎",
		"xserver_closed_header": "Backend Migration",
		"forgot_password_url":   "https://www.kards.com/auth/recovery?lang={lang}",
	})
}
