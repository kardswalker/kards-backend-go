package game

import (
	"errors"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"kards-backend-go/internal/database"
	"kards-backend-go/internal/models"
	"kards-backend-go/pkg/deckcode"
)

const BotPlayerID uint = 900000001

func (gm *GameManager) CreateBotMatch(playerID, deckID uint) (int64, error) {
	var existingMatchID int64
	var existingMatch *Match
	gm.ActiveMatches.Range(func(key, value interface{}) bool {
		m := value.(*Match)
		if m.PlayerLeft == playerID || m.PlayerRight == playerID {
			existingMatchID = key.(int64)
			existingMatch = m
			return false
		}
		return true
	})
	if existingMatchID != 0 {
		if existingMatch != nil && existingMatch.BotEnabled {
			existingMatch.Lock()
			if existingMatch.PlayerStatusRight != "mulligan_done" || len(existingMatch.RightReplacementCards) == 0 {
				gm.prepareSimpleBotMulligan(existingMatch)
			}
			existingMatch.Unlock()
		}
		return existingMatchID, nil
	}

	var user models.User
	if err := database.DB.First(&user, playerID).Error; err != nil {
		return 0, fmt.Errorf("player not found: %w", err)
	}

	var playerDeck models.Deck
	if err := database.DB.First(&playerDeck, deckID).Error; err != nil {
		return 0, fmt.Errorf("deck not found: %w", err)
	}
	if playerDeck.UserID != playerID {
		return 0, errors.New("deck does not belong to player")
	}

	parsedLeft, err := deckcode.ParseDeckCode(playerDeck.DeckCode)
	if err != nil {
		return 0, fmt.Errorf("invalid player deck code: %w", err)
	}
	parsedRight, err := deckcode.ParseDeckCode(playerDeck.DeckCode)
	if err != nil {
		return 0, fmt.Errorf("invalid bot deck code: %w", err)
	}

	cardsL := gm.CreateMatchCards("left", parsedLeft)
	cardsR := gm.CreateMatchCards("right", parsedRight)

	leftDeck := append([]Card(nil), cardsL[1:]...)
	if err := ShuffleDeckCards(leftDeck, "deck_left"); err != nil {
		return 0, fmt.Errorf("failed to shuffle left deck: %w", err)
	}

	rightDeck := append([]Card(nil), cardsR[1:]...)
	if err := ShuffleDeckCards(rightDeck, "deck_right"); err != nil {
		return 0, fmt.Errorf("failed to shuffle right deck: %w", err)
	}

	newID := atomic.AddInt64(&gm.matchIDCounter, 1)

	match := &Match{
		MatchID:           newID,
		Status:            "pending",
		PlayerLeft:        playerID,
		PlayerRight:       BotPlayerID,
		ActionIDSess:      0,
		MatchType:         "training",
		PlayerStatusLeft:  "not_done",
		PlayerStatusRight: "mulligan_done",
		LvlLoadedLeft:     0,
		LvlLoadedRight:    1,
		WinnerSide:        "",
		WinnerID:          0,
		CurrentTurn:       1,
		CurrentActionID:   0,
		Actions:           []int{},
		ActionsData:       make(map[int]string),
		ActionIndex:       make(map[string]int),
		LeftIsOnline:      true,
		RightIsOnline:     true,
		Notifications:     []interface{}{},
		DeckIDLeft:        deckID,
		DeckIDRight:       deckID,
		LeftDeckData:      parsedLeft,
		RightDeckData:     parsedRight,
		LeftCardsData:     cardsL,
		RightCardsData:    cardsR,

		LeftHandCards:         []Card{},
		RightHandCards:        []Card{},
		LeftDeckCards:         []Card{},
		RightDeckCards:        []Card{},
		LeftReplacementCards:  []Card{},
		RightReplacementCards: []Card{},

		LeftPlayerName:  user.PlayerName,
		LeftPlayerTag:   user.PlayerTag,
		RightPlayerName: "Training Bot",
		RightPlayerTag:  0,

		BotEnabled:       true,
		BotSide:          "right",
		BotLastEndedTurn: 0,
	}

	for i := 0; i < len(leftDeck); i++ {
		if i < 4 {
			leftDeck[i].Location = "hand_left"
			leftDeck[i].LocationNumber = i
			match.LeftHandCards = append(match.LeftHandCards, leftDeck[i])
			continue
		}
		leftDeck[i].Location = "deck_left"
		leftDeck[i].LocationNumber = len(match.LeftDeckCards)
		match.LeftDeckCards = append(match.LeftDeckCards, leftDeck[i])
	}

	for i := 0; i < len(rightDeck); i++ {
		if i < 5 {
			rightDeck[i].Location = "hand_right"
			rightDeck[i].LocationNumber = i
			match.RightHandCards = append(match.RightHandCards, rightDeck[i])
			continue
		}
		rightDeck[i].Location = "deck_right"
		rightDeck[i].LocationNumber = len(match.RightDeckCards)
		match.RightDeckCards = append(match.RightDeckCards, rightDeck[i])
	}
	gm.prepareSimpleBotMulligan(match)

	gm.ActiveMatches.Store(newID, match)
	log.Printf("training match created [%d]: %s vs bot", newID, user.PlayerName)

	return newID, nil
}

func (gm *GameManager) TickSimpleBot(match *Match) {
	if match == nil {
		return
	}

	match.Lock()
	defer match.Unlock()

	if !match.BotEnabled || match.BotSide != "right" {
		return
	}
	if match.Status != "running" || match.ActionIDSess == 0 {
		return
	}
	if match.PlayerStatusRight != "mulligan_done" {
		match.PlayerStatusRight = "mulligan_done"
	}
	if match.CurrentTurn%2 != 0 {
		return
	}
	if match.BotLastEndedTurn == match.CurrentTurn {
		return
	}
	if match.BotPendingTurn != match.CurrentTurn {
		match.BotPendingTurn = match.CurrentTurn
		match.BotTurnReadyAt = time.Now().Add(3 * time.Second)
		return
	}
	if time.Now().Before(match.BotTurnReadyAt) {
		return
	}

	turnToEnd := match.CurrentTurn
	startTurnAction := map[string]interface{}{
		"action_type": "XActionStartOfTurn",
		"player_id":   match.PlayerRight,
		"action_data": map[string]interface{}{
			"side": "right",
		},
		"sub_actions": []interface{}{},
		"turn_number": match.CurrentTurn,
	}
	appendBotAction(match, startTurnAction)

	endTurnAction := map[string]interface{}{
		"action_type": "XActionEndOfTurn",
		"player_id":   match.PlayerRight,
		"action_data": map[string]interface{}{
			"reason": "endTurnButton",
			"side":   "right",
		},
		"sub_actions": []interface{}{},
		"turn_number": match.CurrentTurn,
	}
	appendBotAction(match, endTurnAction)

	match.CurrentTurn++
	match.BotLastEndedTurn = turnToEnd
	match.BotPendingTurn = 0
	match.BotTurnReadyAt = time.Time{}
}

func appendBotAction(match *Match, action map[string]interface{}) {
	actionType, _ := action["action_type"].(string)
	playerID := action["player_id"]
	actionData := action["action_data"]
	turnNumber, _ := action["turn_number"].(int)
	AppendActionForSync(match, actionType, playerID, actionData, turnNumber)
}

func (gm *GameManager) prepareSimpleBotMulligan(match *Match) {
	if match == nil {
		return
	}
	if len(match.RightHandCards) == 0 || len(match.RightDeckCards) == 0 {
		match.PlayerStatusRight = "mulligan_done"
		return
	}

	match.RightReplacementCards = make([]Card, 0, 2)
	replacements := 2
	if replacements > len(match.RightHandCards) {
		replacements = len(match.RightHandCards)
	}

	for i := 0; i < replacements; i++ {
		if len(match.RightDeckCards) == 0 {
			break
		}

		deckIdx, err := CryptoIntn(len(match.RightDeckCards))
		if err != nil {
			break
		}

		oldCard := match.RightHandCards[i]
		newCard := match.RightDeckCards[deckIdx]

		oldCard.Location = "deck_right"
		newCard.Location = "hand_right"
		newCard.LocationNumber, oldCard.LocationNumber = oldCard.LocationNumber, newCard.LocationNumber

		match.RightHandCards[i] = newCard
		match.RightDeckCards[deckIdx] = oldCard
		match.RightReplacementCards = append(match.RightReplacementCards, newCard)
	}

	_ = ShuffleDeckCards(match.RightDeckCards, "deck_right")
	match.PlayerStatusRight = "mulligan_done"
}
