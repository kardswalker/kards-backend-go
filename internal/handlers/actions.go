package handlers

import (
	"kards-backend-go/internal/game"
	"kards-backend-go/pkg/security"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 玩家执行操作（出牌、攻击等）时调用
func HandleActions(c *gin.Context) {
	matchIDStr := c.Param("match_id")
	matchID, _ := strconv.ParseInt(matchIDStr, 10, 64)

	// 获取加密包
	var body struct {
		A string `json:"a"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// 1. 获取比赛实例
	val, ok := game.GlobalManager.ActiveMatches.Load(matchID)
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}
	match := val.(*game.Match)

	// 2. 解密动作包 (使用 pkg/security)
	actionIDSess, data, err := security.DecryptPacket(body.A)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	actionType, _ := data["action_type"].(string)

	match.Lock()
	defer match.Unlock()

	// 统一逻辑：自增 ActionID，记录历史，加密后存入 ActionsData
	match.CurrentActionID++
	currentID := match.CurrentActionID

	if actionType == "XStartOfGame" {
		match.ActionIDSess = actionIDSess
	}
	if actionType == "XActionEndOfTurn" {
		match.CurrentTurn++
	}

	// 构造回传/同步的数据结构
	actionToSave := map[string]interface{}{
		"action_id":   currentID,
		"action_type": actionType,
		"player_id":   data["player_id"],
		"action_data": data["action_data"],
		"sub_actions": []interface{}{},
		"turn_number": match.CurrentTurn,
	}

	// 加密动作并存入 Map (供对手 PUT 轮询)
	encryptedPacket := security.EncryptPacket(match.ActionIDSess, actionToSave)
	match.Actions = append(match.Actions, currentID)
	match.ActionsData[currentID] = encryptedPacket

	c.String(http.StatusCreated, "OK")
}

// 玩家轮询对手的动作
func PollActions(c *gin.Context) {
	matchIDStr := c.Param("match_id")
	matchID, _ := strconv.ParseInt(matchIDStr, 10, 64)

	var req struct {
		OpponentID  uint `json:"opponent_id"`
		MinActionID int  `json:"min_action_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	val, ok := game.GlobalManager.ActiveMatches.Load(matchID)
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}
	match := val.(*game.Match)

	match.RLock()
	defer match.RUnlock()

	respData := gin.H{
		"match": gin.H{
			"player_status_left":  match.PlayerStatusLeft,
			"player_status_right": match.PlayerStatusRight,
			"status":              match.Status,
		},
	}

	// 判断对手在线状态
	if match.PlayerLeft == req.OpponentID {
		respData["opponent_polling"] = match.LvlLoadedLeft == 1
	} else if match.PlayerRight == req.OpponentID {
		respData["opponent_polling"] = match.LvlLoadedRight == 1
	}

	// 动作同步逻辑
	var resultActions []string
	startIndex := req.MinActionID - 1
	if startIndex < 0 {
		startIndex = 0
	}

	if startIndex < len(match.Actions) {
		for i := startIndex; i < len(match.Actions); i++ {
			id := match.Actions[i]
			if actData, exists := match.ActionsData[id]; exists {
				resultActions = append(resultActions, actData)
			}
		}
	}

	if len(resultActions) > 0 {
		respData["actions"] = resultActions
	}

	c.JSON(http.StatusOK, respData)
}
