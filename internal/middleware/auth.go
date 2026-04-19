package middleware

import (
	"strings"

	"kards-backend-go/internal/config"
	"kards-backend-go/internal/database"
	"kards-backend-go/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type authClaims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func unauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(401, gin.H{
		"title":       "401 Unauthorized",
		"description": "Warning",
	})
}

func JWTAuth() gin.HandlerFunc {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{config.JWTAlgorithm}))

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/" || path == "/session" || path == "/.com/config" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			unauthorized(c)
			return
		}

		tokenStr := authHeader
		if strings.HasPrefix(authHeader, "JWT ") {
			tokenStr = strings.TrimPrefix(authHeader, "JWT ")
		} else if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}

		claims := &authClaims{}
		token, err := parser.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return config.JWTKey, nil
		})
		if err != nil || token == nil || !token.Valid || claims.UserID == 0 {
			unauthorized(c)
			return
		}

		var user models.User
		if err := database.DB.First(&user, claims.UserID).Error; err != nil {
			unauthorized(c)
			return
		}

		if user.PlayerJWT == "" || user.PlayerJWT != tokenStr {
			unauthorized(c)
			return
		}

		c.Set("user", &user)
		c.Next()
	}
}
