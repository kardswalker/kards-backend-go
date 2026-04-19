package game

import (
	"kards-backend-go/pkg/deckcode"
	"sync"
	"time"
)

// Match 代表一场对局
type Match struct {
	sync.RWMutex

	MatchID      int64
	PlayerLeft   uint
	PlayerRight  uint
	ActionIDSess int
	Status       string
	MatchType    string

	PlayerStatusLeft  string
	PlayerStatusRight string
	LvlLoadedLeft     int
	LvlLoadedRight    int

	WinnerSide string
	WinnerID   uint

	CurrentTurn     int
	CurrentActionID int
	Actions         []int
	ActionsData     map[int]string
	ActionIndex     map[string]int

	LeftIsOnline  bool
	RightIsOnline bool
	Notifications []interface{}

	DeckIDLeft  uint
	DeckIDRight uint

	LeftDeckData  *deckcode.ParsedDeck
	RightDeckData *deckcode.ParsedDeck

	LeftCardsData  []Card
	RightCardsData []Card

	LeftHandCards  []Card
	RightHandCards []Card
	LeftDeckCards  []Card
	RightDeckCards []Card

	LeftReplacementCards  []Card
	RightReplacementCards []Card

	LeftPlayerName  string
	LeftPlayerTag   int
	RightPlayerName string
	RightPlayerTag  int

	BotEnabled       bool
	BotSide          string
	BotLastEndedTurn int
	BotPendingTurn   int
	BotTurnReadyAt   time.Time
}
