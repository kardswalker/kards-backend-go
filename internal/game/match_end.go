package game

import (
	"time"

	"kards-backend-go/pkg/security"
)

func (gm *GameManager) EndMatchBySurrender(playerID uint, reason string) bool {
	ended := false

	gm.ActiveMatches.Range(func(_, value interface{}) bool {
		match := value.(*Match)

		match.Lock()
		defer match.Unlock()

		if match.PlayerLeft != playerID && match.PlayerRight != playerID {
			return true
		}

		if match.Status == "ending" || match.WinnerSide != "" {
			ended = true
			return false
		}

		winnerSide := "left"
		if match.PlayerLeft == playerID {
			winnerSide = "right"
		}

		match.CurrentActionID++
		match.CurrentTurn = 1

		endActionID := match.CurrentActionID
		endAction := map[string]interface{}{
			"action_id":   endActionID,
			"action_type": "ActionEndMatch",
			"player_id":   playerID,
			"action_data": map[string]interface{}{
				"reason":      reason,
				"winner_side": winnerSide,
			},
			"sub_actions": []interface{}{},
			"turn_number": match.CurrentTurn,
		}

		encrypted := security.EncryptPacket(match.ActionIDSess, endAction)
		match.Actions = append(match.Actions, endActionID)
		match.ActionsData[endActionID] = encrypted

		match.Status = "finished"
		match.WinnerSide = winnerSide

		switch winnerSide {
		case "left":
			match.WinnerID = match.PlayerLeft
		case "right":
			match.WinnerID = match.PlayerRight
		default:
			match.WinnerID = 0
		}

		match.PlayerStatusLeft = "mulligan_done"
		match.PlayerStatusRight = "mulligan_done"

		go func(id int64) {
			time.Sleep(10 * time.Second)
			gm.ActiveMatches.Delete(id)
		}(match.MatchID)

		ended = true
		return false
	})

	return ended
}
