package middleware

import (
	"math"
	"strings"

	"kards-backend-go/internal/config"
	"kards-backend-go/internal/database"
	"kards-backend-go/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 1. 排除路径 (白名单)
		if path == "/" || path == "/session" || path == "/.com/config" || strings.HasPrefix(path, "/items/") {
			c.Next()
			return
		}

		// 2. 检查 Authorization 头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"title":       "401 Unauthorized",
				"description": "Warning",
			})
			return
		}

		// 3. 提取 token（支持 "JWT "、"Bearer " 前缀，或直接裸 token）
		tokenStr := authHeader
		if strings.HasPrefix(authHeader, "JWT ") {
			tokenStr = strings.TrimPrefix(authHeader, "JWT ")
		} else if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}
		// 否则直接使用原值

		// 4. 解析 Token (payload_client)
		tokenClient, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return config.JWTKey, nil
		})
		if err != nil || tokenClient == nil || !tokenClient.Valid {
			c.AbortWithStatusJSON(401, gin.H{
				"title":       "401 Unauthorized",
				"description": "Warning",
			})
			return
		}

		claimsClient, ok := tokenClient.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{
				"title":       "401 Unauthorized",
				"description": "Warning",
			})
			return
		}

		// 5. 获取用户ID
		userIDFloat, ok := claimsClient["user_id"].(float64)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{
				"title":       "401 Unauthorized",
				"description": "Warning",
			})
			return
		}
		userID := uint(userIDFloat)

		// 6. 数据库二次校验
		var user models.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			c.AbortWithStatusJSON(401, gin.H{
				"title":       "401 Unauthorized",
				"description": "Warning",
			})
			return
		}

		// 7. 过期时间差校验
		tokenServer, _ := jwt.Parse(user.PlayerJWT, func(token *jwt.Token) (interface{}, error) {
			return config.JWTKey, nil
		})
		if tokenServer == nil {
			c.AbortWithStatusJSON(401, gin.H{
				"title":       "401 Unauthorized",
				"description": "Warning",
			})
			return
		}

		claimsServer, ok := tokenServer.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{
				"title":       "401 Unauthorized",
				"description": "Warning",
			})
			return
		}

		expC, okC := claimsClient["exp"].(float64)
		expS, okS := claimsServer["exp"].(float64)
		if !okC || !okS {
			c.AbortWithStatusJSON(401, gin.H{
				"title":       "401 Unauthorized",
				"description": "Warning",
			})
			return
		}

		if math.Abs(expC-expS) < 86400 {
			c.Set("user", &user)
			c.Next()
		} else {
			c.AbortWithStatusJSON(401, gin.H{
				"title":       "401 Unauthorized",
				"description": "Warning",
			})
		}
	}
}
