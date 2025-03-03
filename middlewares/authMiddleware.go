package middlewares

import (
	"goozinshe/config"
	"goozinshe/models"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, models.NewApiError("authorization header required"))
		c.Abort()
		return
	}

	tokenParts := strings.Split(authHeader, "Bearer ")
	if len(tokenParts) < 2 {
		c.JSON(http.StatusUnauthorized, models.NewApiError("invalid authorization header format"))
		c.Abort()
		return
	}

	tokenString := tokenParts[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Config.JwtSecretKey), nil
	})
	if err != nil || !token.Valid {
		log.Println("ERROR: Ошибка парсинга токена:", err)
		c.JSON(http.StatusUnauthorized, models.NewApiError("invalid token"))
		c.Abort()
		return
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		log.Println("ERROR: Ошибка получения subject:", err)
		c.JSON(http.StatusUnauthorized, models.NewApiError("error while getting subject"))
		c.Abort()
		return
	}

	userId, _ := strconv.Atoi(subject)
	c.Set("userId", userId)
	c.Next()
}
