package game

import (
	"encoding/json"

	"kards-backend-go/pkg/security"
)

type actionFingerprint struct {
	ActionType string      `json:"action_type"`
	PlayerID   interface{} `json:"player_id"`
	ActionData interface{} `json:"action_data"`
	TurnNumber int         `json:"turn_number"`
}

func buildActionFingerprint(actionType string, playerID interface{}, actionData interface{}, turnNumber int) string {
	fp := actionFingerprint{
		ActionType: actionType,
		PlayerID:   playerID,
		ActionData: actionData,
		TurnNumber: turnNumber,
	}
	b, err := json.Marshal(fp)
	if err != nil {
		return ""
	}
	return string(b)
}

func appendActionLocked(match *Match, actionIDSess int, actionType string, playerID interface{}, actionData interface{}, turnNumber int) (int, bool) {
	if match.ActionIndex == nil {
		match.ActionIndex = make(map[string]int)
	}
	if match.ActionsData == nil {
		match.ActionsData = make(map[int]string)
	}

	fp := buildActionFingerprint(actionType, playerID, actionData, turnNumber)
	if fp != "" {
		if existingID, ok := match.ActionIndex[fp]; ok {
			return existingID, true
		}
	}

	match.CurrentActionID++
	actionID := match.CurrentActionID

	actionToSave := map[string]interface{}{
		"action_id":   actionID,
		"action_type": actionType,
		"player_id":   playerID,
		"action_data": actionData,
		"sub_actions": []interface{}{},
		"turn_number": turnNumber,
	}

	encryptedPacket := security.EncryptPacket(actionIDSess, actionToSave)
	match.Actions = append(match.Actions, actionID)
	match.ActionsData[actionID] = encryptedPacket

	if fp != "" {
		match.ActionIndex[fp] = actionID
	}
	if len(match.ActionIndex) > 4096 {
		match.ActionIndex = make(map[string]int)
	}

	return actionID, false
}

func AppendActionForSync(match *Match, actionType string, playerID interface{}, actionData interface{}, turnNumber int) (int, bool) {
	return appendActionLocked(match, match.ActionIDSess, actionType, playerID, actionData, turnNumber)
}
