package handlers

import (
	"net/http"
	"strconv"

	"kards-backend-go/internal/game"
	"kards-backend-go/pkg/security"

	"github.com/gin-gonic/gin"
)

func HandleActions(c *gin.Context) {
	matchIDStr := c.Param("match_id")
	matchID, _ := strconv.ParseInt(matchIDStr, 10, 64)

	var body struct {
		A string `json:"a"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	val, ok := game.GlobalManager.ActiveMatches.Load(matchID)
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}
	match := val.(*game.Match)

	actionIDSess, data, err := security.DecryptPacket(body.A)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	actionType, _ := data["action_type"].(string)
	actionData := data["action_data"]
	playerID := data["player_id"]

	match.Lock()
	if actionType == "XStartOfGame" {
		match.ActionIDSess = actionIDSess
	}
	actionTurn := match.CurrentTurn
	_, duplicated := game.AppendActionForSync(match, actionType, playerID, actionData, actionTurn)
	if !duplicated && actionType == "XActionEndOfTurn" {
		match.CurrentTurn++
	}
	match.Unlock()

	if !duplicated {
		game.GlobalManager.TickSimpleBot(match)
	}

	c.String(http.StatusCreated, "OK")
}

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

	game.GlobalManager.TickSimpleBot(match)

	match.RLock()
	defer match.RUnlock()

	respData := gin.H{
		"match": gin.H{
			"player_status_left":  match.PlayerStatusLeft,
			"player_status_right": match.PlayerStatusRight,
			"status":              match.Status,
		},
	}

	if match.PlayerLeft == req.OpponentID {
		respData["opponent_polling"] = match.LvlLoadedLeft == 1
	} else if match.PlayerRight == req.OpponentID {
		respData["opponent_polling"] = match.LvlLoadedRight == 1
	}

	var resultActions []string
	if len(match.Actions) > 0 {
		seen := make(map[string]struct{})
		for _, id := range match.Actions {
			if id < req.MinActionID {
				continue
			}
			if actData, exists := match.ActionsData[id]; exists {
				if _, ok := seen[actData]; ok {
					continue
				}
				seen[actData] = struct{}{}
				resultActions = append(resultActions, actData)
			}
		}
	}

	if len(resultActions) > 0 {
		respData["actions"] = resultActions
	}

	c.JSON(http.StatusOK, respData)
}
