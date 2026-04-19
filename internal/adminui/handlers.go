package adminui

import (
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/base64"
	"io/fs"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"kards-backend-go/internal/config"
	"kards-backend-go/internal/database"
	"kards-backend-go/internal/game"
	"kards-backend-go/internal/models"

	"github.com/gin-gonic/gin"
)

const (
	adminCookieName = "admin_session"
	adminSessionTTL = 24 * time.Hour
)

var (
	serverStartedAt = time.Now()
	sessionStore    = newAdminSessionStore()
)

//go:embed static/*.html
var staticHTML embed.FS

type adminSessionStore struct {
	mu       sync.RWMutex
	sessions map[string]time.Time
}

type adminUserItem struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	PlayerName string `json:"player_name"`
	PlayerTag  int    `json:"player_tag"`
	Gold       int    `json:"gold"`
	Diamonds   int    `json:"diamonds"`
	SeasonWins int    `json:"season_wins"`
	IsOnline   bool   `json:"is_online"`
}

func newAdminSessionStore() *adminSessionStore {
	return &adminSessionStore{sessions: make(map[string]time.Time)}
}

func (s *adminSessionStore) create() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(buf)
	s.mu.Lock()
	s.sessions[token] = time.Now().Add(adminSessionTTL)
	s.mu.Unlock()
	return token, nil
}

func (s *adminSessionStore) valid(token string) bool {
	if token == "" {
		return false
	}
	now := time.Now()
	s.mu.RLock()
	expireAt, ok := s.sessions[token]
	s.mu.RUnlock()
	if !ok || expireAt.Before(now) {
		if ok {
			s.mu.Lock()
			delete(s.sessions, token)
			s.mu.Unlock()
		}
		return false
	}
	s.mu.Lock()
	s.sessions[token] = now.Add(adminSessionTTL)
	s.mu.Unlock()
	return true
}

func (s *adminSessionStore) remove(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

func RegisterRoutes(r *gin.Engine) {
	r.GET("/admin/login", serveLoginPage)
	r.POST("/admin/api/login", loginAPI)
	r.POST("/admin/api/logout", adminRequired(true), logoutAPI)

	adminPage := r.Group("/admin")
	adminPage.Use(adminRequired(false))
	{
		adminPage.GET("", serveDashboardPage)
		adminPage.GET("/system", serveSystemPage)
		adminPage.GET("/users", serveUsersPage)
		adminPage.GET("/users/manage", serveUsersManagePage)
		adminPage.GET("/matches", serveMatchesPage)
	}

	adminAPI := r.Group("/admin/api")
	adminAPI.Use(adminRequired(true))
	{
		adminAPI.GET("/overview", overviewAPI)
		adminAPI.GET("/system", systemAPI)
		adminAPI.GET("/users", usersAPI)
		adminAPI.GET("/users/all", usersAllAPI)
		adminAPI.PUT("/users/:id", updateUserAPI)
		adminAPI.DELETE("/users/:id", deleteUserAPI)
		adminAPI.GET("/matches", matchesAPI)
	}
}

func adminRequired(api bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(adminCookieName)
		if err != nil || !sessionStore.valid(token) {
			if api {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

func loginAPI(c *gin.Context) {
	var req struct {
		Password string `json:"password" form:"password"`
	}
	if err := c.ShouldBind(&req); err != nil || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if subtle.ConstantTimeCompare([]byte(req.Password), []byte(config.AdminPassword)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}
	token, err := sessionStore.create()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}
	c.SetCookie(adminCookieName, token, int(adminSessionTTL.Seconds()), "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func logoutAPI(c *gin.Context) {
	token, _ := c.Cookie(adminCookieName)
	sessionStore.remove(token)
	c.SetCookie(adminCookieName, "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func overviewAPI(c *gin.Context) {
	var onlineCount int64
	game.GlobalManager.OnlineClients.Range(func(_, _ interface{}) bool {
		onlineCount++
		return true
	})
	var matchCount int64
	game.GlobalManager.ActiveMatches.Range(func(_, _ interface{}) bool {
		matchCount++
		return true
	})
	c.JSON(http.StatusOK, gin.H{
		"uptime_seconds": int64(time.Since(serverStartedAt).Seconds()),
		"uptime_text":    time.Since(serverStartedAt).Round(time.Second).String(),
		"online_users":   onlineCount,
		"active_matches": matchCount,
		"server_time":    time.Now().UTC().Format(time.RFC3339),
	})
}

func systemAPI(c *gin.Context) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	c.JSON(http.StatusOK, gin.H{
		"go_routines":          runtime.NumGoroutine(),
		"cpu_cores":            runtime.NumCPU(),
		"memory_alloc_mb":      bytesToMB(mem.Alloc),
		"memory_heap_inuse_mb": bytesToMB(mem.HeapInuse),
		"memory_sys_mb":        bytesToMB(mem.Sys),
		"gc_count":             mem.NumGC,
		"uptime_text":          time.Since(serverStartedAt).Round(time.Second).String(),
	})
}

func usersAPI(c *gin.Context) {
	var users []models.User
	if err := database.DB.Select("id", "username", "player_name", "player_tag", "gold", "diamonds", "season_wins", "is_online").Where("is_online = ?", true).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items := make([]adminUserItem, 0, len(users))
	for _, u := range users {
		items = append(items, toAdminUser(u))
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	c.JSON(http.StatusOK, gin.H{"count": len(items), "users": items})
}

func usersAllAPI(c *gin.Context) {
	limit := 500
	if q := c.Query("limit"); q != "" {
		if v, err := strconv.Atoi(q); err == nil && v > 0 && v <= 2000 {
			limit = v
		}
	}
	page := 1
	if q := c.Query("page"); q != "" {
		if v, err := strconv.Atoi(q); err == nil && v > 0 {
			page = v
		}
	}
	var total int64
	if err := database.DB.Model(&models.User{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
		if page > totalPages {
			page = totalPages
		}
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit
	var users []models.User
	if err := database.DB.Select("id", "username", "player_name", "player_tag", "gold", "diamonds", "season_wins", "is_online").Order("id desc").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items := make([]adminUserItem, 0, len(users))
	for _, u := range users {
		items = append(items, toAdminUser(u))
	}
	c.JSON(http.StatusOK, gin.H{
		"count":       len(items),
		"total":       total,
		"page":        page,
		"page_size":   limit,
		"total_pages": totalPages,
		"users":       items,
	})
}

func updateUserAPI(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := uint(id64)
	var req struct {
		PlayerName *string `json:"player_name"`
		PlayerTag  *int    `json:"player_tag"`
		Gold       *int    `json:"gold"`
		Diamonds   *int    `json:"diamonds"`
		SeasonWins *int    `json:"season_wins"`
		IsOnline   *bool   `json:"is_online"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	updates := map[string]interface{}{}
	if req.PlayerName != nil {
		updates["player_name"] = *req.PlayerName
	}
	if req.PlayerTag != nil {
		updates["player_tag"] = *req.PlayerTag
	}
	if req.Gold != nil {
		updates["gold"] = *req.Gold
	}
	if req.Diamonds != nil {
		updates["diamonds"] = *req.Diamonds
	}
	if req.SeasonWins != nil {
		updates["season_wins"] = *req.SeasonWins
	}
	if req.IsOnline != nil {
		updates["is_online"] = *req.IsOnline
	}
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}
	if err := database.DB.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func deleteUserAPI(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := uint(id64)
	if err := database.DB.Where("user_id = ?", id).Delete(&models.Deck{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := database.DB.Where("id = ?", id).Delete(&models.User{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func matchesAPI(c *gin.Context) {
	type matchItem struct {
		MatchID           int64  `json:"match_id"`
		Status            string `json:"status"`
		MatchType         string `json:"match_type"`
		CurrentTurn       int    `json:"current_turn"`
		CurrentActionID   int    `json:"current_action_id"`
		PlayerLeft        uint   `json:"player_left"`
		PlayerRight       uint   `json:"player_right"`
		PlayerStatusLeft  string `json:"player_status_left"`
		PlayerStatusRight string `json:"player_status_right"`
		WinnerID          uint   `json:"winner_id"`
		WinnerSide        string `json:"winner_side"`
		BotEnabled        bool   `json:"bot_enabled"`
	}
	items := make([]matchItem, 0)
	game.GlobalManager.ActiveMatches.Range(func(_, value interface{}) bool {
		m := value.(*game.Match)
		m.RLock()
		items = append(items, matchItem{
			MatchID:           m.MatchID,
			Status:            m.Status,
			MatchType:         m.MatchType,
			CurrentTurn:       m.CurrentTurn,
			CurrentActionID:   m.CurrentActionID,
			PlayerLeft:        m.PlayerLeft,
			PlayerRight:       m.PlayerRight,
			PlayerStatusLeft:  m.PlayerStatusLeft,
			PlayerStatusRight: m.PlayerStatusRight,
			WinnerID:          m.WinnerID,
			WinnerSide:        m.WinnerSide,
			BotEnabled:        m.BotEnabled,
		})
		m.RUnlock()
		return true
	})
	sort.Slice(items, func(i, j int) bool { return items[i].MatchID > items[j].MatchID })
	c.JSON(http.StatusOK, gin.H{"count": len(items), "matches": items})
}

func serveLoginPage(c *gin.Context) {
	body, err := fs.ReadFile(staticHTML, "static/login.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to load login page")
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", body)
}

func serveDashboardPage(c *gin.Context) {
	serveAdminPage(c, "Main Panel", "static/dashboard.html")
}

func serveSystemPage(c *gin.Context) {
	serveAdminPage(c, "System", "static/system.html")
}

func serveUsersPage(c *gin.Context) {
	serveAdminPage(c, "Online Users", "static/users.html")
}

func serveUsersManagePage(c *gin.Context) {
	serveAdminPage(c, "User Manager", "static/users_manage.html")
}

func serveMatchesPage(c *gin.Context) {
	serveAdminPage(c, "Matches", "static/matches.html")
}

func serveAdminPage(c *gin.Context, title, bodyPath string) {
	page, err := adminPageHTML(title, bodyPath)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
}

func bytesToMB(v uint64) float64 {
	return float64(v) / 1024.0 / 1024.0
}

func toAdminUser(u models.User) adminUserItem {
	return adminUserItem{
		ID:         u.ID,
		Username:   u.Username,
		PlayerName: u.PlayerName,
		PlayerTag:  u.PlayerTag,
		Gold:       u.Gold,
		Diamonds:   u.Diamonds,
		SeasonWins: u.SeasonWins,
		IsOnline:   u.IsOnline,
	}
}

func adminPageHTML(title, bodyPath string) (string, error) {
	tpl, err := fs.ReadFile(staticHTML, "static/base.html")
	if err != nil {
		return "", err
	}
	body, err := fs.ReadFile(staticHTML, bodyPath)
	if err != nil {
		return "", err
	}
	page := string(tpl)
	page = strings.ReplaceAll(page, "{{TITLE}}", title)
	page = strings.ReplaceAll(page, "{{BODY}}", string(body))
	return page, nil
}
