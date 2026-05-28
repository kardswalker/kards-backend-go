package handlers

import (
	"net/http"

	"kards-backend-go/internal/database"
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

	var deck models.Deck
	if err := database.DB.Select("id", "user_id").First(&deck, req.DeckID).Error; err != nil || deck.UserID != user.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deck not found"})
		return
	}

	game.GlobalManager.AddMatchPlayers(user.ID, req.DeckID)

	c.String(http.StatusOK, "OK")
}

func JoinSinglePlayerLobby(c *gin.Context) {
	var req struct {
		PlayerID  uint `json:"player_id"`
		DeckID    uint `json:"deck_id"`
		ExtraData struct {
			MatchType string `json:"match_type"`
		} `json:"extra_data"`
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

	if _, err := game.GlobalManager.CreateBotMatch(user.ID, req.DeckID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.String(http.StatusCreated, "OK")
}
