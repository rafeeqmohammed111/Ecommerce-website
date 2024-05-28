package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func SessionCreate(email string, role string, c *gin.Context) {
	session := sessions.Default(c)
	session.Set("user_email", email)
	session.Set("role", role)
	err := session.Save()
	if err!= nil {
        fmt.Println("Failed to create session:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
    } else {
        c.JSON(http.StatusOK, gin.H{"message": "session created successfully"})
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
func AuthMiddleware(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        session := sessions.Default(c)
        userRole := session.Get("role")
        fmt.Printf("Checking role: %v\n", userRole) // Debugging statement
        if userRole == nil || userRole.(string)!= role {
            c.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": "You don't have access to this resource"})
            c.Abort()
            return
        }
        c.Next()
    }
}

