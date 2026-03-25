package handlers

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

type Card struct {
	CardType             string `json:"card_type"`
	Count                int    `json:"count"`
	GoldCardCount        int    `json:"gold_card_count"`
	ID                   int    `json:"id"`
	RecentlyCraftedCount int    `json:"recently_crafted_count"`
}

type Library struct {
	Cards    []Card `json:"cards"`
	NewCards []Card `json:"new_cards"`
}

//go:embed library.json
var libraryJSON []byte

var LibraryData Library

func init() {
	if err := json.Unmarshal(libraryJSON, &LibraryData); err != nil {
		panic(fmt.Sprintf("failed to load embedded library.json: %v", err))
	}

	if LibraryData.Cards == nil {
		LibraryData.Cards = []Card{}
	}
	if LibraryData.NewCards == nil {
		LibraryData.NewCards = []Card{}
	}
}
