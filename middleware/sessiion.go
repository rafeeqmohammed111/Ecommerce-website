package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func SessionCreate(email string, role string, c *gin.Context) {
	session := sessions.Default(c)
	session.Set("user_email", email)
	session.Set("role", role)
	err := session.Save()
	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to create session",
		})
	}
}

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		role := session.Get("role")
		if role == nil || role != "admin" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}
