package handlers

import (
	"kards-backend-go/internal/game"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleMulligan 处理玩家换牌请求
func HandleMulligan(c *gin.Context) {
	var req struct {
		DiscardedCardIDs []int `json:"discarded_card_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid request body"})
		return
	}

	match, _, side, ok := currentUserMatch(c)
	if !ok {
		return
	}

	match.Lock()
	defer match.Unlock()

	var hand *[]game.Card
	var deck *[]game.Card
	var replacements *[]game.Card

	if side == "left" {
		hand, deck = &match.LeftHandCards, &match.LeftDeckCards
		replacements = &match.LeftReplacementCards
	} else {
		hand, deck = &match.RightHandCards, &match.RightDeckCards
		replacements = &match.RightReplacementCards
	}

	*replacements = make([]game.Card, 0)

	// 遍历要换掉的卡
	for _, discardID := range req.DiscardedCardIDs {
		for i, card := range *hand {
			if card.CardID == discardID {
				if len(*deck) == 0 {
					break
				}

				deckIdx, err := game.CryptoIntn(len(*deck))
				if err != nil {
					c.AbortWithStatusJSON(500, gin.H{"error": "failed to generate random replacement"})
					return
				}

				newCard := (*deck)[deckIdx]

				// 交换 Location 和 LocationNumber
				oldCard := card
				oldCard.Location = "deck_" + side
				newCard.Location = "hand_" + side
				newCard.LocationNumber, oldCard.LocationNumber = oldCard.LocationNumber, newCard.LocationNumber

				(*hand)[i] = newCard
				(*deck)[deckIdx] = oldCard

				*replacements = append(*replacements, newCard)
				break
			}
		}
	}

	deckLocation := "deck_" + side
	if err := game.ShuffleDeckCards(*deck, deckLocation); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "failed to reshuffle deck"})
		return
	}

	// 更新玩家状态
	if side == "left" {
		match.PlayerStatusLeft = "mulligan_done"
	} else {
		match.PlayerStatusRight = "mulligan_done"
	}

	c.JSON(http.StatusOK, gin.H{
		"deck":              *deck,
		"replacement_cards": *replacements,
	})
}

// GetMulliganLeft 返回左手玩家的牌库和替换卡
func GetMulliganLeft(c *gin.Context) {
	match, _, _, ok := currentUserMatch(c)
	if !ok {
		return
	}

	match.RLock()
	defer match.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"deck":              match.LeftDeckCards,
		"replacement_cards": match.LeftReplacementCards,
	})
}

// GetMulliganRight 返回右手玩家的牌库和替换卡
func GetMulliganRight(c *gin.Context) {
	match, _, _, ok := currentUserMatch(c)
	if !ok {
		return
	}

	match.RLock()
	defer match.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"deck":              match.RightDeckCards,
		"replacement_cards": match.RightReplacementCards,
	})
}
