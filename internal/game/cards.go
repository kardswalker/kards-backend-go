package game

import (
	"kards-backend-go/pkg/deckcode"
)

type Card struct {
	CardID         int    `json:"card_id"`
	Name           string `json:"name"`
	Faction        string `json:"faction,omitempty"`
	IsGold         bool   `json:"is_gold"`
	Location       string `json:"location"`
	LocationNumber int    `json:"location_number"`
}

func (gm *GameManager) CreateMatchCards(side string, parsed *deckcode.ParsedDeck) []Card {
	var cards []Card
	var startID int
	var deckLoc string
	var hqLoc string

	if side == "left" {
		startID = 1
		hqLoc = "board_hqleft"
		deckLoc = "deck_left"
	} else {
		startID = 41
		hqLoc = "board_hqright"
		deckLoc = "deck_right"
	}

	hqCode := parsed.HQ
	if len(hqCode) > 2 {
		hqCode = hqCode[:2]
	}

	hqName := "card_location_unknown"
	if info, ok := DeckCodeIDsTable[hqCode]; ok {
		hqName = info.Card
	}

	cards = append(cards, Card{
		CardID:         startID,
		Name:           hqName,
		Faction:        parsed.MainCountry,
		IsGold:         false,
		Location:       hqLoc,
		LocationNumber: 0,
	})

	cardCounter := startID + 1
	locCounter := 0
	for code, count := range parsed.Cards {
		name := "card_unknown"
		if info, ok := DeckCodeIDsTable[code]; ok {
			name = info.Card
		}
		for i := 0; i < count; i++ {
			cards = append(cards, Card{
				CardID:         cardCounter,
				Name:           name,
				IsGold:         false,
				Location:       deckLoc,
				LocationNumber: locCounter,
			})
			cardCounter++
			locCounter++
		}
	}
	return cards
}
