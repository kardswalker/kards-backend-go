package handlers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kards-backend-go/internal/config"
	"kards-backend-go/internal/database"
	"kards-backend-go/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func GetRoot(c *gin.Context) {
	baseURL := "http://" + config.Host + ":" + config.Port

	// 1. 初始化基础 Endpoints
	endpoints := map[string]interface{}{
		"draft":        baseURL + "/draft/",
		"email":        baseURL + "/email/set",
		"lobbyplayers": baseURL + "/lobbyplayers",
		"matches":      baseURL + "/matches",
		"matches2":     baseURL + "/matches/v2/",
		"my_draft":     nil,
		"my_items":     nil,
		"my_player":    nil,
		"players":      baseURL + "/players",
		"purchase":     baseURL + "/store/v2/txn",
		"root":         baseURL,
		"session":      baseURL + "/session",
		"store":        baseURL + "/store/",
		"tourneys":     baseURL + "/tourney/",
		"transactions": baseURL + "/store/txn",
		"view_offers":  baseURL + "/store/v2/",
	}

	currentUser := map[string]interface{}{}

	// 2. 核心手工解析逻辑 (不依赖中间件)
	authHeader := c.GetHeader("Authorization")

	// 调试日志：如果发现还是没数据，看一眼控制台有没有输出这个
	if authHeader != "" {
		tokenStr := strings.TrimPrefix(authHeader, "JWT ")

		// 解析客户端 Token
		tokenClient, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return config.JWTKey, nil
		})

		if tokenClient != nil && tokenClient.Valid {
			if claimsClient, ok := tokenClient.Claims.(jwt.MapClaims); ok {

				// 关键点：JWT 里的数字需要先转为 float64 再转 uint
				var uID uint
				if idVal, exists := claimsClient["user_id"]; exists {
					uID = uint(idVal.(float64))
				}

				// 查找数据库
				var user models.User
				if err := database.DB.First(&user, uID).Error; err == nil {

					// 解析数据库存的 Token (payload_server)
					tokenServer, _ := jwt.Parse(user.PlayerJWT, func(token *jwt.Token) (interface{}, error) {
						return config.JWTKey, nil
					})

					if tokenServer != nil {
						claimsServer := tokenServer.Claims.(jwt.MapClaims)

						expC := claimsClient["exp"].(float64)
						expS := claimsServer["exp"].(float64)

						// 校验过期时间差 (abs < 24h)
						if math.Abs(expC-expS) < 86400 {
							uIDStr := strconv.Itoa(int(user.ID))

							// 填充动态 URL
							endpoints["my_draft"] = fmt.Sprintf("%s/draft/%s", baseURL, uIDStr)
							endpoints["my_items"] = fmt.Sprintf("%s/items/%s", baseURL, uIDStr)
							endpoints["my_player"] = fmt.Sprintf("%s/players/%s", baseURL, uIDStr)

							// 填充 current_user (精准复刻 Python 字段)
							currentUser = map[string]interface{}{
								"client_id":   uIDStr,
								"exp":         uIDStr, // 特殊逻辑：Kards 要求此处填 ID
								"external_id": user.Username,
								"iat":         int64(expS), // 使用 server 的 exp 作为 iat
								"identity_id": uIDStr,
								"iss":         "cometkards",
								"jti":         "",
								"language":    "zh-Hans",
								"payment":     "notavailable",
								"player_id":   uIDStr,
								"provider":    "device",
								"roles":       []string{},
								"tier":        "LIVE",
								"user_id":     uIDStr,
								"user_name":   user.Username,
							}
						}
					}
				}
			}
		}
	}

	// 3. 返回 JSON (会自动经过 FixResponseHeaders 中间件移除 charset)
	c.JSON(http.StatusOK, gin.H{
		"build_info": gin.H{
			"build_timestamp": "2025-10-13T17:31:40Z",
			"commit_hash":     "dfaf581c",
			"version":         config.Version,
		},
		"current_user": currentUser,
		"endpoints":    endpoints,
		"host_info": gin.H{
			"container_name": "kards-backend-LIVE",
			"docker_image":   "618005890699.dkr.ecr.eu-west-1.amazonaws.com/kards-backend:live",
			"host_address":   config.Host,
			"host_name":      "cometkards",
			"instance_id":    "i-03598bff8bd68fdee",
		},
		"server_time":  time.Now().UTC().Format("2006-01-02T15:04:05.000") + "Z",
		"service_name": "kards-backend",
		"tenant_name":  "1939-kardslive",
		"tier_name":    "LIVE",
	})
}
