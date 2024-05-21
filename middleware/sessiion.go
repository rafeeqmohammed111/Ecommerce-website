package middleware

import (
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
func AuthMiddleware(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		sessionRole := session.Get("role")

		if sessionRole != role {
			c.JSON(401, gin.H{
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
