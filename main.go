package main

import (
	"net/http"
	"project/initializer"
	"project/routers"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"github.com/gin-gonic/gin"
)

func init() {
	initializer.EnvLoad()
	initializer.LoadDatabase()
}

func main() {
	router := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	userGroup := router.Group("/")
	routers.UserGroup(userGroup)

	admin := router.Group("/admin")
	routers.AdminRouter(admin)

	router.Run(":8080")

}
func UserGroup(group *gin.RouterGroup) {
	group.GET("/user/profile", func(c *gin.Context) {
		// Your user profile logic here
		c.JSON(http.StatusOK, gin.H{"message": "User Profile"})
	})
}

// Example of a simple handler for demonstration
func AdminRouter(group *gin.RouterGroup) {
	group.GET("/admin/dashboard", func(c *gin.Context) {
		// Your admin dashboard logic here
		c.JSON(http.StatusOK, gin.H{"message": "Admin Dashboard"})
	})
}

// Define your sessions package import here
var _ = sessions.Default
