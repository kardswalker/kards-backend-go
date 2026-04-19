package game

import "kards-backend-go/internal/database"

func (gm *GameManager) IsPlayerOnline(userID uint) bool {
	_, ok := gm.OnlineClients.Load(userID)
	return ok
}

func (gm *GameManager) SetPlayerOnlineStatus(userID uint, online bool) {
	if database.DB != nil {
		database.DB.Model(struct {
			ID uint
		}{}).Table("users").Where("id = ?", userID).Update("is_online", online)
	}

	gm.ActiveMatches.Range(func(_, value interface{}) bool {
		match := value.(*Match)

		match.Lock()
		if match.PlayerLeft == userID {
			match.LeftIsOnline = online
		}
		if match.PlayerRight == userID {
			match.RightIsOnline = online
		}
		match.Unlock()

		return true
	})
}
