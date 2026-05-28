package game

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"kards-backend-go/internal/database"
	"kards-backend-go/internal/models"
	"kards-backend-go/pkg/deckcode"
)

type MatchRequest struct {
	UserID uint
	DeckID uint
}

type GameManager struct {
	ActiveMatches sync.Map
	OnlineClients sync.Map

	WaitingQueue []*MatchRequest
	queueMutex   sync.Mutex

	matchIDCounter int64
}

var GlobalManager = &GameManager{
	WaitingQueue: make([]*MatchRequest, 0),
}

func (gm *GameManager) AddMatchPlayers(userID, deckID uint) bool {
	gm.queueMutex.Lock()
	defer gm.queueMutex.Unlock()

	if gm.HasActiveMatch(userID) {
		return true
	}

	for _, req := range gm.WaitingQueue {
		if req.UserID == userID {
			return true
		}
	}

	gm.WaitingQueue = append(gm.WaitingQueue, &MatchRequest{UserID: userID, DeckID: deckID})
	log.Printf("👤 玩家 %d 加入匹配池, 当前人数: %d", userID, len(gm.WaitingQueue))
	return true
}

func (gm *GameManager) HasActiveMatch(userID uint) bool {
	found := false
	gm.ActiveMatches.Range(func(_, value interface{}) bool {
		match := value.(*Match)
		if match.PlayerLeft == userID || match.PlayerRight == userID {
			found = true
			return false
		}
		return true
	})
	return found
}

func (gm *GameManager) IsUserWaiting(userID uint) bool {
	gm.queueMutex.Lock()
	defer gm.queueMutex.Unlock()

	for _, req := range gm.WaitingQueue {
		if req.UserID == userID {
			return true
		}
	}
	return false
}

func (gm *GameManager) StartMatchmaker() {
	log.Println("🚀 撮合系统已在后台启动...")

	for {
		gm.queueMutex.Lock()
		if len(gm.WaitingQueue) >= 2 {
			p1 := gm.WaitingQueue[0]
			p2 := gm.WaitingQueue[1]
			gm.WaitingQueue = gm.WaitingQueue[2:]
			gm.queueMutex.Unlock()

			go func() {
				if err := gm.CreateMatch(p1, p2); err != nil {
					log.Printf("failed to create match: %v", err)
				}
			}()
		} else {
			gm.queueMutex.Unlock()
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func (gm *GameManager) CreateMatch(p1, p2 *MatchRequest) error {
	var uL, uR models.User
	var dL, dR models.Deck
	if err := database.DB.First(&uL, p1.UserID).Error; err != nil {
		return fmt.Errorf("left player not found: %w", err)
	}
	if err := database.DB.First(&uR, p2.UserID).Error; err != nil {
		return fmt.Errorf("right player not found: %w", err)
	}
	if err := database.DB.First(&dL, p1.DeckID).Error; err != nil {
		return fmt.Errorf("left deck not found: %w", err)
	}
	if dL.UserID != p1.UserID {
		return fmt.Errorf("left deck does not belong to player %d", p1.UserID)
	}
	if err := database.DB.First(&dR, p2.DeckID).Error; err != nil {
		return fmt.Errorf("right deck not found: %w", err)
	}
	if dR.UserID != p2.UserID {
		return fmt.Errorf("right deck does not belong to player %d", p2.UserID)
	}

	parsedL, err := deckcode.ParseDeckCode(dL.DeckCode)
	if err != nil {
		return fmt.Errorf("invalid left deck code: %w", err)
	}
	parsedR, err := deckcode.ParseDeckCode(dR.DeckCode)
	if err != nil {
		return fmt.Errorf("invalid right deck code: %w", err)
	}

	cardsL := gm.CreateMatchCards("left", parsedL)
	cardsR := gm.CreateMatchCards("right", parsedR)

	leftDeck := append([]Card(nil), cardsL[1:]...)
	if err := ShuffleDeckCards(leftDeck, "deck_left"); err != nil {
		return fmt.Errorf("failed to shuffle left deck: %w", err)
	}

	rightDeck := append([]Card(nil), cardsR[1:]...)
	if err := ShuffleDeckCards(rightDeck, "deck_right"); err != nil {
		return fmt.Errorf("failed to shuffle right deck: %w", err)
	}

	newID := atomic.AddInt64(&gm.matchIDCounter, 1)
	match := &Match{
		MatchID:           newID,
		Status:            "pending",
		PlayerLeft:        p1.UserID,
		PlayerRight:       p2.UserID,
		ActionIDSess:      0,
		MatchType:         "battle",
		PlayerStatusLeft:  "not_done",
		PlayerStatusRight: "not_done",
		LvlLoadedLeft:     0,
		LvlLoadedRight:    0,
		WinnerSide:        "",
		WinnerID:          0,
		CurrentTurn:       1,
		CurrentActionID:   0,
		Actions:           []int{},
		ActionsData:       make(map[int]string),
		ActionIndex:       make(map[string]int),

		LeftIsOnline:  true,
		RightIsOnline: true,
		Notifications: []interface{}{},

		DeckIDLeft:  p1.DeckID,
		DeckIDRight: p2.DeckID,

		LeftDeckData:   parsedL,
		RightDeckData:  parsedR,
		LeftCardsData:  cardsL,
		RightCardsData: cardsR,

		LeftHandCards:         []Card{},
		RightHandCards:        []Card{},
		LeftDeckCards:         []Card{},
		RightDeckCards:        []Card{},
		LeftReplacementCards:  []Card{},
		RightReplacementCards: []Card{},

		LeftPlayerName:  uL.PlayerName,
		LeftPlayerTag:   uL.PlayerTag,
		RightPlayerName: uR.PlayerName,
		RightPlayerTag:  uR.PlayerTag,
	}

	match.LeftHandCards, match.LeftDeckCards = dealStartingCards(leftDeck, 4, "hand_left", "deck_left")
	match.RightHandCards, match.RightDeckCards = dealStartingCards(rightDeck, 5, "hand_right", "deck_right")

	gm.ActiveMatches.Store(newID, match)
	log.Printf("⚔️ 对战已就绪 [%d]: %s vs %s", newID, uL.PlayerName, uR.PlayerName)
	return nil
}
