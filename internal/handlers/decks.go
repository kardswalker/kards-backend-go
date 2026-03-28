package handlers

import (
	"fmt"
	"kards-backend-go/internal/database"
	"kards-backend-go/internal/models"
	"kards-backend-go/pkg/utils"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateDeck 创建新卡组
func CreateDeck(c *gin.Context) {
	user, _, ok := utils.GetAuthedPlayer(c)
	if !ok {
		return
	}

	var req utils.DeckRequest
	_ = c.ShouldBindJSON(&req)

	now := utils.GetKardsNow()
	favorite := false
	if req.Favorite != nil {
		favorite = *req.Favorite
	}

	name := req.Name
	if name == "" {
		name = fmt.Sprintf("Deck %d", user.ID)
	}

	deck := models.Deck{
		Name:        name,
		MainFaction: req.MainFaction,
		AllyFaction: req.AllyFaction,
		CardBack:    "cardback_default",
		DeckCode:    req.DeckCode,
		Favorite:    favorite,
		UserID:      user.ID,
		LastPlayed:  now,
		CreateDate:  now,
		ModifyDate:  now,
	}

	if err := database.DB.Create(&deck).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, utils.BuildDeckHeader(deck))
}

// UpdateDeck 更新卡组
func UpdateDeck(c *gin.Context) {
	user, _, ok := utils.GetAuthedPlayer(c)
	if !ok {
		return
	}

	deckID64, err := strconv.ParseUint(c.Param("deck_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid deck_id"})
		return
	}
	deckID := uint(deckID64)

	var req utils.DeckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var deck models.Deck
	if err := database.DB.Where("id = ? AND user_id = ?", deckID, user.ID).First(&deck).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "deck not found"})
		return
	}

	updates := map[string]interface{}{}
	if req.Action == "fill" {
		if req.DeckCode != "" {
			updates["deck_code"] = req.DeckCode
		}
		if req.Name != "" {
			updates["name"] = req.Name
		}
		if req.Favorite != nil {
			updates["favorite"] = *req.Favorite
		}
		updates["modify_date"] = utils.GetKardsNow()

		if err := database.DB.Model(&deck).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.Status(http.StatusNoContent)
}

// ChangeDeck 更改卡组（重命名或更换卡背）
func ChangeDeck(c *gin.Context) {
	user, _, ok := utils.GetAuthedPlayer(c)
	if !ok {
		return
	}

	var req utils.DeckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var deck models.Deck
	if err := database.DB.Where("id = ? AND user_id = ?", req.ID, user.ID).First(&deck).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "deck not found"})
		return
	}

	updates := map[string]interface{}{}
	switch req.Action {
	case "rename":
		if req.Name != "" {
			updates["name"] = req.Name
		}
	case "change_card_back":
		if req.Name != "" {
			updates["card_back"] = req.Name
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action"})
		return
	}
	updates["modify_date"] = utils.GetKardsNow()

	if err := database.DB.Model(&deck).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)

	// 调试日志
	log.Printf("更改卡组成功: player_id=%d, deck_id=%d, action=%s, value=%s", user.ID, req.ID, req.Action, req.Name)
}

// DeleteDeck 删除卡组
func DeleteDeck(c *gin.Context) {
	user, _, ok := utils.GetAuthedPlayer(c)
	if !ok {
		return
	}

	deckID64, err := strconv.ParseUint(c.Param("deck_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid deck_id"})
		return
	}
	deckID := uint(deckID64)

	if err := database.DB.Where("id = ? AND user_id = ?", deckID, user.ID).Delete(&models.Deck{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
