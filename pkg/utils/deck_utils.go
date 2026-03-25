package utils

import (
	"net/http"
	"strconv"

	"kards-backend-go/internal/game"
	"kards-backend-go/internal/models"

	"github.com/gin-gonic/gin"
)

// DeckRequest 用于接收前端卡组请求
type DeckRequest struct {
	Action      string `json:"action"`
	Name        string `json:"name"`
	MainFaction string `json:"main_faction"`
	AllyFaction string `json:"ally_faction"`
	CardBack    string `json:"card_back"`
	DeckCode    string `json:"deck_code"`
	Favorite    *bool  `json:"favorite"`
}

// 获取当前请求的用户并校验权限
func GetAuthedPlayer(c *gin.Context) (*models.User, uint, bool) {
	val, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return nil, 0, false
	}
	user := val.(*models.User)
	playerID64, err := strconv.ParseUint(c.Param("player_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player_id"})
		return nil, 0, false
	}
	playerID := uint(playerID64)
	if user.ID != playerID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return nil, 0, false
	}
	return user, playerID, true
}

// BuildDeckHeader 构建返回给客户端的卡组信息
func BuildDeckHeader(deck models.Deck) game.DeckHeader {
	return game.DeckHeader{
		Name:        deck.Name,
		MainFaction: deck.MainFaction,
		AllyFaction: deck.AllyFaction,
		CardBack:    deck.CardBack,
		DeckCode:    deck.DeckCode,
		Favorite:    deck.Favorite,
		ID:          deck.ID,
		PlayerID:    deck.UserID,
		LastPlayed:  deck.LastPlayed,
		CreateDate:  deck.CreateDate,
		ModifyDate:  deck.ModifyDate,
	}
}
