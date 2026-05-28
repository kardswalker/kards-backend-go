package game

import (
	"strings"
	"time"

	"kards-backend-go/pkg/security"
)

type MatchEndResult struct {
	WinnerSide string
	WinnerID   uint
	Reason     string
}

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

		result := MatchEndResult{
			WinnerSide: oppositeSide(match, playerID),
			Reason:     reason,
		}

		gm.endMatchLocked(match, playerID, result)

		ended = true
		return false
	})

	return ended
}

func (gm *GameManager) EndMatch(match *Match, playerID uint, result MatchEndResult) bool {
	if match == nil {
		return false
	}

	match.Lock()
	defer match.Unlock()

	if match.PlayerLeft != playerID && match.PlayerRight != playerID {
		return false
	}
	if match.Status == "ending" || match.WinnerSide != "" {
		return true
	}

	gm.endMatchLocked(match, playerID, result)
	return true
}

func (gm *GameManager) endMatchLocked(match *Match, playerID uint, result MatchEndResult) {
	result = normalizeEndResult(match, playerID, result)

	match.CurrentActionID++
	match.CurrentTurn = 1

	endActionID := match.CurrentActionID
	endAction := map[string]interface{}{
		"action_id":   endActionID,
		"action_type": "ActionEndMatch",
		"player_id":   playerID,
		"action_data": map[string]interface{}{
			"winner_id":   result.WinnerID,
			"reason":      result.Reason,
			"winner_side": result.WinnerSide,
		},
		"sub_actions": []interface{}{},
		"turn_number": match.CurrentTurn,
	}

	encrypted := security.EncryptPacket(match.ActionIDSess, endAction)
	match.Actions = append(match.Actions, endActionID)
	match.ActionsData[endActionID] = encrypted

	match.Status = "finished"
	match.WinnerSide = result.WinnerSide
	match.WinnerID = result.WinnerID
	match.PlayerStatusLeft = "mulligan_done"
	match.PlayerStatusRight = "mulligan_done"

	go func(id int64) {
		time.Sleep(10 * time.Second)
		gm.ActiveMatches.Delete(id)
	}(match.MatchID)
}

func normalizeEndResult(match *Match, playerID uint, result MatchEndResult) MatchEndResult {
	result.WinnerSide = normalizeSide(result.WinnerSide)
	result.Reason = strings.TrimSpace(result.Reason)
	if result.Reason == "" {
		result.Reason = "surrender"
	}

	if result.WinnerSide == "" && result.WinnerID != 0 {
		result.WinnerSide = sideForPlayer(match, result.WinnerID)
	}
	if result.WinnerSide == "" {
		result.WinnerSide = oppositeSide(match, playerID)
	}

	switch result.WinnerSide {
	case "left":
		result.WinnerID = match.PlayerLeft
	case "right":
		result.WinnerID = match.PlayerRight
	default:
		result.WinnerID = 0
	}

	return result
}

func normalizeSide(side string) string {
	switch strings.ToLower(strings.TrimSpace(side)) {
	case "left":
		return "left"
	case "right":
		return "right"
	default:
		return ""
	}
}

func sideForPlayer(match *Match, playerID uint) string {
	switch playerID {
	case match.PlayerLeft:
		return "left"
	case match.PlayerRight:
		return "right"
	default:
		return ""
	}
}

func oppositeSide(match *Match, playerID uint) string {
	if match.PlayerLeft == playerID {
		return "right"
	}
	return "left"
}
