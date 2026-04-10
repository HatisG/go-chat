package middleware

import (
	"go-chat/internal/config"
	"go-chat/pkg/errcode"
	"go-chat/pkg/jwt"
	"go-chat/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, errcode.Unauthorized)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, errcode.Unauthorized)
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwt.ParseToken(tokenString, config.AppConfig.JWT.Secret)
		if err != nil {
			response.Error(c, errcode.Unauthorized)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()

	}

}
