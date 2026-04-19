package game

import (
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

	for _, req := range gm.WaitingQueue {
		if req.UserID == userID {
			return true
		}
	}

	gm.WaitingQueue = append(gm.WaitingQueue, &MatchRequest{UserID: userID, DeckID: deckID})
	log.Printf("👤 玩家 %d 加入匹配池, 当前人数: %d", userID, len(gm.WaitingQueue))
	return true
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

			go gm.CreateMatch(p1, p2)
		} else {
			gm.queueMutex.Unlock()
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func (gm *GameManager) CreateMatch(p1, p2 *MatchRequest) {
	newID := atomic.AddInt64(&gm.matchIDCounter, 1)

	var uL, uR models.User
	var dL, dR models.Deck
	database.DB.First(&uL, p1.UserID)
	database.DB.First(&uR, p2.UserID)
	database.DB.First(&dL, p1.DeckID)
	database.DB.First(&dR, p2.DeckID)

	parsedL, _ := deckcode.ParseDeckCode(dL.DeckCode)
	parsedR, _ := deckcode.ParseDeckCode(dR.DeckCode)

	cardsL := gm.CreateMatchCards("left", parsedL)
	cardsR := gm.CreateMatchCards("right", parsedR)

	leftDeck := append([]Card(nil), cardsL[1:]...)
	if err := ShuffleDeckCards(leftDeck, "deck_left"); err != nil {
		log.Printf("failed to shuffle left deck for match %d: %v", newID, err)
		return
	}

	rightDeck := append([]Card(nil), cardsR[1:]...)
	if err := ShuffleDeckCards(rightDeck, "deck_right"); err != nil {
		log.Printf("failed to shuffle right deck for match %d: %v", newID, err)
		return
	}

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

	gm.ActiveMatches.Store(newID, match)
	log.Printf("⚔️ 对战已就绪 [%d]: %s vs %s", newID, uL.PlayerName, uR.PlayerName)
}
