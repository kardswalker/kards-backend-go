package game

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"kards-backend-go/internal/config"
	"kards-backend-go/internal/database"
	"kards-backend-go/internal/models"

	"github.com/gorilla/websocket"
)

// 允许所有来源的跨域请求，并指明支持 "ws" 子协议
var wsUpgrader = websocket.Upgrader{
	CheckOrigin:  func(r *http.Request) bool { return true },
	Subprotocols: []string{"ws"},
}

// StartWSServer 在独立端口 (默认 5232) 启动 WebSocket 服务
func (gm *GameManager) StartWSServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", gm.handleWSConnection)

	addr := fmt.Sprintf(":%d", config.WSPort)
	log.Printf("📡 WebSocket 独立服务器已启动，监听端口: %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("❌ WebSocket 服务器启动失败: %v", err)
	}
}

// handleWSConnection 处理每个客户端的连接
func (gm *GameManager) handleWSConnection(w http.ResponseWriter, r *http.Request) {
	// 1. 获取 Authorization 头
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Printf("⚠️ WS 认证失败: 缺少 Authorization 头")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 智能处理 Header 格式
	tokenStr := authHeader
	if strings.HasPrefix(authHeader, "JWT ") {
		tokenStr = authHeader[4:]
	} else if strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr = authHeader[7:]
	}

	// 2. 查找用户
	var user models.User
	if err := database.DB.Where("player_jwt = ?", tokenStr).First(&user).Error; err != nil {
		log.Printf("⚠️ WS 认证失败: 数据库中未找到该 Token")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 3. 升级 HTTP 连接为 WebSocket
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WS 升级失败: %v", err)
		return
	}

	// 4. 登记客户端
	gm.OnlineClients.Store(user.ID, conn)
	gm.SetPlayerOnlineStatus(user.ID, true)
	log.Printf("🔌 玩家 %d (%s) 已连接 WebSocket，使用的协议: %s", user.ID, user.PlayerName, conn.Subprotocol())

	// 保证断开时清理资源
	defer func() {
		gm.OnlineClients.Delete(user.ID)
		gm.SetPlayerOnlineStatus(user.ID, false)
		gm.EndMatchBySurrender(user.ID, "surrender")
		conn.Close()
		log.Printf("👋 玩家 %d (%s) WebSocket 已断开", user.ID, user.PlayerName)
	}()

	// 6. 进入消息处理循环 (处理客户端后续发来的真实 Ping 和对战数据)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break // 发生错误或客户端主动断开
		}

		// 解析 JSON 消息
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		// 提取必需字段
		channel, _ := msg["channel"].(string)
		msgContent, _ := msg["message"].(string)
		context, _ := msg["context"].(string)

		// 安全提取 receiver ID
		var receiverID uint
		if recVal, ok := msg["receiver"]; ok {
			switch v := recVal.(type) {
			case float64:
				receiverID = uint(v)
			case string:
				var parsed uint
				fmt.Sscanf(v, "%d", &parsed)
				receiverID = parsed
			}
		}

		// --- 核心路由分支 ---
		switch channel {

		case "ping":
			// 返回正常的 Pong
			resp := map[string]interface{}{
				"message":   "pong",
				"channel":   "ping",
				"context":   "",
				"timestamp": time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
				"sender":    user.ID,
				"receiver":  "",
			}
			conn.WriteJSON(resp)

		case "touchcard":
			// 触摸卡牌动画：直接转发给对手
			if receiverConn, ok := gm.OnlineClients.Load(receiverID); ok {
				resp := map[string]interface{}{
					"message":   msgContent,
					"channel":   "touchcard",
					"context":   context,
					"timestamp": time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
					"sender":    user.ID,
					"receiver":  receiverID,
				}
				receiverConn.(*websocket.Conn).WriteJSON(resp)
			}

		case "notification":
			// 对战通知与状态同步
			if msgContent == "websocketcheck" || msgContent == "matchaction" || msgContent == "im_here" {
				if receiverConn, ok := gm.OnlineClients.Load(receiverID); ok {

					respContext := context
					if msgContent == "im_here" {
						respContext = ""
					}

					resp := map[string]interface{}{
						"message":   msgContent,
						"channel":   "notification",
						"context":   respContext,
						"timestamp": time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
						"sender":    user.ID,
						"receiver":  receiverID,
					}
					receiverConn.(*websocket.Conn).WriteJSON(resp)
				}
			}
		}
	}
}
