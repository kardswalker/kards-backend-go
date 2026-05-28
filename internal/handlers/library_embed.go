package handlers

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"kards-backend-go/internal/config"
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
	data, err := loadJSONResource(config.LibraryJSONPath, libraryJSON, "library.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(data, &LibraryData); err != nil {
		panic(fmt.Sprintf("failed to parse library.json: %v", err))
	}

	if LibraryData.Cards == nil {
		LibraryData.Cards = []Card{}
	}
	if LibraryData.NewCards == nil {
		LibraryData.NewCards = []Card{}
	}
}
