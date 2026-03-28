package models

import (
	"time"

	"gorm.io/gorm"
)

// Item 代表玩家物品
type Item struct {
	ItemID  string                 `json:"item_id"`
	Slot    string                 `json:"slot,omitempty"`    // 装备槽位
	Faction string                 `json:"faction,omitempty"` // 阵营
	Cnt     int                    `json:"cnt,omitempty"`     // 数量
	Details map[string]interface{} `json:"details,omitempty"`
}

// User 玩家结构体
type User struct {
	ID         uint   `gorm:"primaryKey"`
	Username   string `gorm:"size:50;uniqueIndex"` // 已有唯一索引
	Password   string `gorm:"size:50"`
	PlayerName string `gorm:"size:50"`
	PlayerJWT  string `gorm:"type:text"`
	PlayerTag  int    `gorm:"index"`

	// 国家等级与经验
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

	// 资产与道具
	Diamonds        int    `gorm:"default:0"`
	Gold            int    `gorm:"default:0"`
	DraftAdmissions int    `gorm:"default:0"`
	DoubleXpEndDate string `gorm:"size:50"`

	// 物品和装备（存储为JSON）
	ItemsJSON    string `gorm:"type:text;serializer:json"` // JSON格式的物品列表
	EquippedJSON string `gorm:"type:text;serializer:json"` // JSON格式的装备列表

	// 状态
	IsOnline   bool `gorm:"default:false"`
	SeasonWins int  `gorm:"default:0"`
	CreatedAt  time.Time

	// GORM 关联 Deck
	Decks []Deck `gorm:"foreignKey:UserID"`

	// 内存字段，不存数据库
	Items         []Item `gorm:"-" json:"items"`          // 玩家拥有物品
	EquippedItems []Item `gorm:"-" json:"equipped_items"` // 当前装备物品
}

// 可选：方便自动迁移创建表
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	return
}
