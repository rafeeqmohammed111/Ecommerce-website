package routers

import (
	controller "project/controller/admin"

	"project/middleware"

	"github.com/gin-gonic/gin"
)

func AdminRouter(r *gin.RouterGroup) {
	// Apply the middleware to all routes in this group

	r.GET("/", controller.AdminPage)
	r.POST("/login", controller.AdminLogin)

	r.Use(middleware.AdminAuthMiddleware())

	r.POST("/logout", controller.AdminLogout)

	// User management
	r.GET("/user_management", controller.UserList)
	r.PATCH("/user_management/user_edit/:ID", controller.EditUserDetails)
	r.PATCH("/user_management/user_block/:ID", controller.BlockUser)
	r.DELETE("/user_management/user_delete/:ID", controller.DeleteUser)

	// Product management
	r.GET("/products", controller.ProductList)
	r.POST("/products/add_products", controller.AddProducts)
	r.PATCH("/products/edit_products/:ID", controller.EditProducts)
	r.DELETE("/products/delete_products/:ID", controller.DeleteProducts)
	r.POST("/upload", controller.UploadProductImage)

	// Category management
	r.GET("/categories", controller.CategoryList)
	r.POST("/categories/add_category", controller.AddCategory)
	r.PATCH("/categories/edit_category/:ID", controller.EditCategories)
	r.DELETE("/categories/delete_category/:ID", controller.DeleteCategories)
	r.PATCH("/categories/block_category/:ID", controller.BlockCategory)
}
