package handlers

import (
	"math/rand"
	"net/http"
	"time"

	"kards-backend-go/internal/database"
	"kards-backend-go/internal/models"

	"github.com/gin-gonic/gin"
)

func UpdatePlayer(c *gin.Context) {
	// 1. 从中间件获取当前登录的用户对象
	val, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"title": "401 Unauthorized", "description": "Warning"})
		return
	}
	user := val.(*models.User)

	// 2. 解析请求体
	var body struct {
		Action string      `json:"action"`
		Value  interface{} `json:"value"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 3. 执行起名逻辑
	if body.Action == "set-name" {
		newName, ok := body.Value.(string)
		if !ok || newName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid name value"})
			return
		}

		// 更新用户名
		user.PlayerName = newName
		rand.Seed(time.Now().UnixNano())
		user.PlayerTag = rand.Intn(9000) + 1000

		// 保存到数据库 (对应 TS 的 User.prototype.store.call(user))
		database.DB.Save(&user)
	}

	// 4. 构造响应体
	responseBody := gin.H{
		"player_name": user.PlayerName,
		"player_tag":  user.PlayerTag,
	}

	c.JSON(http.StatusOK, responseBody)
}
