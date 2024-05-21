package main

import (
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

	user := router.Group("/")
	routers.UserGroup(user)

	admin := router.Group("/admin")

	routers.AdminRouter(admin)

	router.Run(":8080")

}
