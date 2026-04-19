package models

import (
	"time"

	"gorm.io/gorm"
)

// Item 代表玩家物品
type Item struct {
	ItemID  string                 `json:"item_id"`
	Slot    string                 `json:"slot,omitempty"`
	Faction string                 `json:"faction,omitempty"`
	Cnt     int                    `json:"cnt,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// User 玩家结构体
type User struct {
	ID         uint   `gorm:"primaryKey"`
	Username   string `gorm:"size:50;uniqueIndex"`
	Password   string `gorm:"size:50"`
	PlayerName string `gorm:"size:50"`
	PlayerJWT  string `gorm:"type:text"`
	PlayerTag  int    `gorm:"index"`

	BritainLevel        int `gorm:"default:500"`
	BritainLevelClaimed int `gorm:"default:500"`
	BritainXp           int `gorm:"default:0"`
	GermanyLevel        int `gorm:"default:500"`
	GermanyLevelClaimed int `gorm:"default:500"`
	GermanyXp           int `gorm:"default:0"`
	JapanLevel          int `gorm:"default:500"`
	JapanLevelClaimed   int `gorm:"default:500"`
	JapanXp             int `gorm:"default:0"`
	SovietLevel         int `gorm:"default:500"`
	SovietLevelClaimed  int `gorm:"default:500"`
	SovietXp            int `gorm:"default:0"`
	UsaLevel            int `gorm:"default:500"`
	UsaLevelClaimed     int `gorm:"default:500"`
	UsaXp               int `gorm:"default:0"`

	Diamonds        int    `gorm:"default:0"`
	Gold            int    `gorm:"default:0"`
	DraftAdmissions int    `gorm:"default:0"`
	DoubleXpEndDate string `gorm:"size:50"`

	ItemsJSON           string `gorm:"type:text;serializer:json"`
	EquippedJSON        string `gorm:"type:text;serializer:json"`
	PurchasedOffersJSON string `gorm:"type:text"`

	IsOnline   bool `gorm:"default:false"`
	SeasonWins int  `gorm:"default:0"`
	CreatedAt  time.Time

	Decks []Deck `gorm:"foreignKey:UserID"`

	Items         []Item `gorm:"-" json:"items"`
	EquippedItems []Item `gorm:"-" json:"equipped_items"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	return
}
