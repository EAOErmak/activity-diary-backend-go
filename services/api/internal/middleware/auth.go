package middleware

import (
	"net/http"
	"strings"

	"activitydiary/api/internal/auth"
	"activitydiary/api/internal/response"

	"github.com/gin-gonic/gin"
)

func Auth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.Error(c, http.StatusUnauthorized, "Authorization token is required")
			c.Abort()
			return
		}
		claims, err := jwtManager.Parse(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "Invalid authorization token")
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "ADMIN" {
			response.Error(c, http.StatusForbidden, "Admin access is required")
			c.Abort()
			return
		}
		c.Next()
	}
}
