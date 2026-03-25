package models

type Deck struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"size:50;default:''"` // 卡组名
	CardBack string `gorm:"size:50;default:''"` // 卡背

	MainFaction string `gorm:"size:20;default:''"` // 主国
	AllyFaction string `gorm:"size:20;default:''"` // 盟国

	DeckCode string `gorm:"size:250;default:''"` // 卡组代码
	Favorite bool   `gorm:"default:false"`       // 是否收藏

	// 属于哪个玩家
	UserID uint

	LastPlayed string `gorm:"size:50;default:''"` // 最后使用时间
	CreateDate string `gorm:"size:50;default:''"` // 创建时间
	ModifyDate string `gorm:"size:50;default:''"` // 修改时间
}
