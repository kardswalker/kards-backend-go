package handlers

import (
	"net/http"

	"kards-backend-go/internal/game"
	"kards-backend-go/internal/models"

	"github.com/gin-gonic/gin"
)

func JoinLobby(c *gin.Context) {
	var req struct {
		PlayerID uint `json:"player_id"`
		DeckID   uint `json:"deck_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	user := c.MustGet("user").(*models.User)

	if user.ID != req.PlayerID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 调用统一管理器的排队方法
	game.GlobalManager.AddMatchPlayers(user.ID, req.DeckID)

	c.String(http.StatusOK, "OK")
}
