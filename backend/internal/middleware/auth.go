package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"negative-ion-respirator/backend/internal/service"
)

func AuthRequired(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 40101, "message": "missing token"})
			c.Abort()
			return
		}
		claims, err := authSvc.ValidateToken(strings.TrimPrefix(auth, "Bearer "))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 40102, "message": "invalid token"})
			c.Abort()
			return
		}
		c.Set("user_id", claims["user_id"])
		c.Set("role", claims["role"])
		c.Next()
	}
}
