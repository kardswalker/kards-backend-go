package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kards-backend-go/internal/adminui"
	"kards-backend-go/internal/config"
	"kards-backend-go/internal/database"
	"kards-backend-go/internal/game"
	"kards-backend-go/internal/handlers"
	"kards-backend-go/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Printf("server initializing")
	if err := config.PromptInitialSetup(); err != nil {
		log.Fatalf("failed to initialize config: %v", err)
	}

	database.InitDB()
	go game.GlobalManager.StartMatchmaker()
	go game.GlobalManager.StartWSServer()

	r := gin.New()
	r.RemoveExtraSlash = true
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.FixResponseHeaders())

	r.GET("/.com/config", handlers.GetConfig)
	r.GET("/", handlers.GetRoot)
	r.POST("/session", handlers.HandleSession)
	adminui.RegisterRoutes(r)
	r.NoRoute(func(c *gin.Context) {
		c.Header("Connection", "close")
		c.JSON(http.StatusNotFound, gin.H{})
	})

	auth := r.Group("/")
	auth.Use(middleware.JWTAuth())
	{
		auth.GET("/players/:player_id/library", handlers.GetLibrary)
		auth.GET("/matches/v2/reconnect", handlers.Reconnect)
		auth.POST("/lobbyplayers", handlers.JoinLobby)
		auth.POST("/singleplayerlobby", handlers.JoinSinglePlayerLobby)
		auth.PUT("/matches/v2/:match_id/mulligan", handlers.HandleMulligan)
		auth.GET("/matches/v2/:match_id/mulligan-left", handlers.GetMulliganLeft)
		auth.GET("/matches/v2/:match_id/mulligan-right", handlers.GetMulliganRight)
		auth.GET("/items/:player_id", handlers.GetItems)
		auth.POST("/items/:player_id", handlers.ChangeItem)
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
			player.PUT("/decks", handlers.ChangeDeck)
			player.DELETE("/decks/:deck_id", handlers.DeleteDeck)
			player.PUT("/player/:player_id/decks", handlers.ChangeDeck)
		}
	}

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", config.Port),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("http server listening on :%d", config.Port)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
