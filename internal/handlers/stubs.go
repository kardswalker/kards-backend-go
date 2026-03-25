package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetPacks(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}

func GetLibrary(c *gin.Context) {
	c.JSON(http.StatusOK, LibraryData)
}

func Heartbeat(c *gin.Context) {
	c.Header("Connection", "keep-alive")
	c.Header("Keep-Alive", "timeout=5")
	c.JSON(http.StatusOK, gin.H{})
}

func HandleFP(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"custom": "your_value",
	})
}

func HandleFriends(c *gin.Context) {
	c.Status(http.StatusOK)
}
