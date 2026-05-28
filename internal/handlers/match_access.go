package handlers

import (
	"net/http"
	"strconv"

	"kards-backend-go/internal/game"
	"kards-backend-go/internal/models"

	"github.com/gin-gonic/gin"
)

func currentUserMatch(c *gin.Context) (*game.Match, *models.User, string, bool) {
	matchID, err := strconv.ParseInt(c.Param("match_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match_id"})
		return nil, nil, "", false
	}

	user := c.MustGet("user").(*models.User)
	val, ok := game.GlobalManager.ActiveMatches.Load(matchID)
	if !ok {
		c.Status(http.StatusNotFound)
		return nil, nil, "", false
	}

	match := val.(*game.Match)
	match.RLock()
	defer match.RUnlock()

	switch user.ID {
	case match.PlayerLeft:
		return match, user, "left", true
	case match.PlayerRight:
		return match, user, "right", true
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "not a participant"})
		return nil, nil, "", false
	}
}
