package middlewares

import (
	"goozinshe/config"
	"goozinshe/models"
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

	tokenString := strings.Split(authHeader, "Bearer ")[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Config.JwtSecretKey), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, models.NewApiError("invalid token"))
		c.Abort()
		return
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewApiError("error while getting subject"))
		c.Abort()
		return
	}

	userId, _ := strconv.Atoi(subject)
	c.Set("userId", userId)
	c.Next()
}
