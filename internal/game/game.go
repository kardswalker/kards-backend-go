package game

import (
	"kards-backend-go/internal/database"
	"kards-backend-go/internal/models"
)

// DeckHeader 客户端返回卡组结构
type DeckHeader struct {
	ID          uint   `json:"id"`
	PlayerID    uint   `json:"player_id"`
	Name        string `json:"name"`
	MainFaction string `json:"main_faction"`
	AllyFaction string `json:"ally_faction"`
	CardBack    string `json:"card_back"`
	DeckCode    string `json:"deck_code"`
	Favorite    bool   `json:"favorite"`
	LastPlayed  string `json:"last_played"`
	CreateDate  string `json:"create_date"`
	ModifyDate  string `json:"modify_date"`
}

// GetDecksForUser 高效返回用户卡组列表
func GetDecksForUser(userID uint) []DeckHeader {
	var decks []models.Deck
	if err := database.DB.Where("user_id = ?", userID).Find(&decks).Error; err != nil {
		return []DeckHeader{}
	}

	result := make([]DeckHeader, len(decks))
	for i, d := range decks {
		result[i] = DeckHeader{
			ID:          d.ID,
			PlayerID:    d.UserID,
			Name:        d.Name,
			MainFaction: d.MainFaction,
			AllyFaction: d.AllyFaction,
			CardBack:    d.CardBack,
			DeckCode:    d.DeckCode,
			Favorite:    d.Favorite,
			LastPlayed:  d.LastPlayed,
			CreateDate:  d.CreateDate,
			ModifyDate:  d.ModifyDate,
		}
	}
	return result
}
