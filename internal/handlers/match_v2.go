package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"kards-backend-go/internal/config"
	"kards-backend-go/internal/database"
	"kards-backend-go/internal/game"
	"kards-backend-go/internal/models"
	"kards-backend-go/pkg/security"

	"github.com/gin-gonic/gin"
)

func getStringField(src interface{}, names ...string) string {
	if src == nil {
		return ""
	}

	v := reflect.ValueOf(src)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	if !v.IsValid() {
		return ""
	}

	for _, name := range names {
		f := v.FieldByName(name)
		if f.IsValid() && f.Kind() == reflect.String {
			return f.String()
		}
	}

	return ""
}

func firstCard(cards []game.Card) interface{} {
	if len(cards) == 0 {
		return nil
	}
	return cards[0]
}

func normalizeCardBack(cardBack string) string {
	if cardBack == "" {
		return "cardback_default"
	}
	return cardBack
}

func GetMatchInfo(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	var match *game.Match
	game.GlobalManager.ActiveMatches.Range(func(key, value interface{}) bool {
		m := value.(*game.Match)
		if m.PlayerLeft == user.ID || m.PlayerRight == user.ID {
			match = m
			return false
		}
		return true
	})

	if match == nil {
		c.JSON(http.StatusOK, nil)
		return
	}

	match.RLock()
	defer match.RUnlock()

	localSide := "left"
	localPlayerID := match.PlayerLeft
	if user.ID == match.PlayerRight {
		localSide = "right"
		localPlayerID = match.PlayerRight
	}

	var leftDeck models.Deck
	leftCardBack := "cardback_default"
	if err := database.DB.Select("card_back").First(&leftDeck, "id = ?", match.DeckIDLeft).Error; err == nil {
		leftCardBack = normalizeCardBack(leftDeck.CardBack)
	}

	var rightDeck models.Deck
	rightCardBack := "cardback_default"
	if err := database.DB.Select("card_back").First(&rightDeck, "id = ?", match.DeckIDRight).Error; err == nil {
		rightCardBack = normalizeCardBack(rightDeck.CardBack)
	}

	var leftUser models.User
	if err := database.DB.First(&leftUser, "id = ?", match.PlayerLeft).Error; err == nil && leftUser.EquippedJSON != "" {
		json.Unmarshal([]byte(leftUser.EquippedJSON), &leftUser.EquippedItems)
	}
	if leftUser.EquippedItems == nil {
		leftUser.EquippedItems = []models.Item{}
	}

	var rightUser models.User
	if err := database.DB.First(&rightUser, "id = ?", match.PlayerRight).Error; err == nil && rightUser.EquippedJSON != "" {
		json.Unmarshal([]byte(rightUser.EquippedJSON), &rightUser.EquippedItems)
	}
	if rightUser.EquippedItems == nil {
		rightUser.EquippedItems = []models.Item{}
	}

	equipmentLeft := make([]string, len(leftUser.EquippedItems))
	for i, item := range leftUser.EquippedItems {
		equipmentLeft[i] = item.ItemID
	}

	equipmentRight := make([]string, len(rightUser.EquippedItems))
	for i, item := range rightUser.EquippedItems {
		equipmentRight[i] = item.ItemID
	}

	resp := gin.H{
		"local_subactions": true,
		"match_and_starting_data": gin.H{
			"match": gin.H{
				"action_player_id":    localPlayerID,
				"action_side":         localSide,
				"actions":             match.Actions,
				"actions_url":         fmt.Sprintf("http://%s:%d/matches/v2/%d/actions", config.Host, config.Port, match.MatchID),
				"current_action_id":   match.CurrentActionID,
				"current_turn":        match.CurrentTurn,
				"deck_id_left":        match.DeckIDLeft,
				"deck_id_right":       match.DeckIDRight,
				"left_is_online":      match.LeftIsOnline,
				"match_id":            match.MatchID,
				"match_type":          match.MatchType,
				"match_url":           fmt.Sprintf("http://%s:%d/matches/v2/%d", config.Host, config.Port, match.MatchID),
				"modify_date":         time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
				"notifications":       match.Notifications,
				"player_id_left":      match.PlayerLeft,
				"player_id_right":     match.PlayerRight,
				"player_status_left":  match.PlayerStatusLeft,
				"player_status_right": match.PlayerStatusRight,
				"right_is_online":     match.RightIsOnline,
				"start_side":          "left",
				"status":              match.Status,
				"winner_id":           match.WinnerID,
				"winner_side":         match.WinnerSide,
			},
			"starting_data": gin.H{
				"ally_faction_left":    getStringField(match.LeftDeckData, "AllyCountry", "AllyFaction"),
				"ally_faction_right":   getStringField(match.RightDeckData, "AllyCountry", "AllyFaction"),
				"card_back_left":       leftCardBack,
				"card_back_right":      rightCardBack,
				"starting_hand_left":   match.LeftHandCards,
				"starting_hand_right":  match.RightHandCards,
				"deck_left":            match.LeftDeckCards,
				"deck_right":           match.RightDeckCards,
				"equipment_left":       equipmentLeft,
				"equipment_right":      equipmentRight,
				"is_ai_match":          false,
				"left_player_name":     match.LeftPlayerName,
				"left_player_officer":  false,
				"left_player_tag":      match.LeftPlayerTag,
				"location_card_left":   firstCard(match.LeftCardsData),
				"location_card_right":  firstCard(match.RightCardsData),
				"player_id_left":       match.PlayerLeft,
				"player_id_right":      match.PlayerRight,
				"player_stars_left":    20,
				"player_stars_right":   20,
				"right_player_name":    match.RightPlayerName,
				"right_player_officer": false,
				"right_player_tag":     match.RightPlayerTag,
			},
		},
		"action_player_id": localPlayerID,
		"action_side":      localSide,
	}

	c.JSON(http.StatusOK, resp)
}

func GetMatchStatus(c *gin.Context) {
	matchID, _ := strconv.ParseInt(c.Param("match_id"), 10, 64)
	user := c.MustGet("user").(*models.User)

	val, ok := game.GlobalManager.ActiveMatches.Load(matchID)
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}
	match := val.(*game.Match)

	match.Lock()
	if user.ID == match.PlayerLeft {
		match.LvlLoadedLeft = 1
	} else if user.ID == match.PlayerRight {
		match.LvlLoadedRight = 1
	}

	if match.LvlLoadedLeft == 1 && match.LvlLoadedRight == 1 {
		match.Status = "running"
	}
	currentStatus := match.Status
	match.Unlock()

	c.String(http.StatusOK, currentStatus)
}

func EndMatch(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	matchID, _ := strconv.ParseInt(c.Param("match_id"), 10, 64)
	val, ok := game.GlobalManager.ActiveMatches.Load(matchID)
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	match := val.(*game.Match)

	var body struct {
		A string `json:"a"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if body.A == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty encrypted payload"})
		return
	}
	if len(body.A) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "encrypted payload too short"})
		return
	}

	actionIDSess, data, err := security.DecryptPacket(body.A)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("decrypt error: %v", err)})
		return
	}

	action, _ := data["action"].(string)
	if action != "end-match" {
		c.Status(http.StatusNoContent)
		return
	}

	valData, ok := data["value"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value field in payload"})
		return
	}

	match.Lock()
	match.ActionIDSess = actionIDSess
	match.Unlock()

	_ = valData
	game.GlobalManager.EndMatchBySurrender(user.ID, "surrender")
	c.String(http.StatusOK, "OK")
}

func Reconnect(c *gin.Context) {
	c.JSON(http.StatusOK, "running")
}
