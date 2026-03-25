package main

import (
	"fmt"
	"log"

	"kards-backend-go/internal/config"
	"kards-backend-go/internal/database"
	"kards-backend-go/internal/game"
	"kards-backend-go/internal/handlers"
	"kards-backend-go/internal/middleware"

	"github.com/gin-gonic/gin"

	"net/http"
)

func main() {
	log.Printf("🔧 正在初始化 Kards Server")
	database.InitDB()
	go game.GlobalManager.StartMatchmaker()
	go game.GlobalManager.StartWSServer()
	r := gin.Default()
	r.Use(middleware.FixResponseHeaders())
	r.Use(gin.Recovery())
	r.GET("/.com/config", handlers.GetConfig)
	r.GET("/", handlers.GetRoot)
	r.POST("/session", handlers.HandleSession)
	r.NoRoute(func(c *gin.Context) {
		c.Header("Connection", "close")
		c.JSON(http.StatusNotFound, gin.H{})
	})

	auth := r.Group("/")
	auth.Use(middleware.JWTAuth())
	{
		auth.GET("/items/:player_id", handlers.GetItems)
		auth.GET("/players/:player_id/library", handlers.GetLibrary)
		auth.GET("/matches/v2/reconnect", handlers.Reconnect)
		auth.POST("/lobbyplayers", handlers.JoinLobby)
		auth.PUT("/matches/v2/:match_id/mulligan", handlers.HandleMulligan)
		auth.GET("/matches/v2/:match_id/mulligan-left", handlers.GetMulliganLeft)
		auth.GET("/matches/v2/:match_id/mulligan-right", handlers.GetMulliganRight)
		auth.GET("/players/:player_id/items", handlers.GetItems)
		auth.PUT("/players/:player_id/items/change", handlers.ChangeItem)
		auth.GET("/matches/v2/", handlers.GetMatchInfo)
		match := auth.Group("/matches/v2/:match_id")
		{
			match.GET("", handlers.GetMatchStatus)
			match.PUT("", handlers.EndMatch)
			match.POST("/actions", handlers.HandleActions)
			match.PUT("/actions", handlers.PollActions)
			match.POST("/mulligan", handlers.HandleMulligan)
			match.GET("/mulligan/left", handlers.GetMulliganLeft)
			match.GET("/mulligan/right", handlers.GetMulliganRight)
		}

		player := auth.Group("/players/:player_id")
		{
			player.PUT("", handlers.UpdatePlayer)
			player.PUT("/heartbeat", handlers.Heartbeat)
			player.GET("/packs", handlers.GetPacks)
			player.POST("/friends", handlers.HandleFriends)
			player.POST("/decks", handlers.CreateDeck)
			player.PUT("/decks/:deck_id", handlers.UpdateDeck)
			player.DELETE("/decks/:deck_id", handlers.DeleteDeck)
		}
	}

	log.Printf("🚀 HTTP 服务已启动! 监听端口: %s", config.Port)
	if err := r.Run(fmt.Sprintf(":%d", config.Port)); err != nil {
		log.Fatalf("❌ 无法启动服务器: %v", err)
	}
}
