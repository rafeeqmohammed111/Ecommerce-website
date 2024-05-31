package routers

import (
	"project/middleware"
	"project/user"

	"github.com/gin-gonic/gin"
)

var roleuser = "User"

func UserGroup(r *gin.RouterGroup) {
	r.GET("/", user.LoadingPage)
	r.GET("/user/getproducts", user.GetAllProducts)
	r.POST("/user/signup", user.UserSignUp)
	r.POST("/user/login", user.UserLogin)
	r.POST("/user/logout", user.UserLogin)
	r.POST("/user/signup/otp", user.OtpCheck)
// 
r.POST("/user/forgotpass", user.ForgotUserCheck)
	r.POST("/user/forgotpass/otp", user.ForgotOtpCheck)
	r.PATCH("/user/new-password", user.NewPasswordSet)

	// user profile

	r.GET("/user/profile", middleware.AuthMiddleware(roleuser), user.UserProfile)
	r.POST("/user/address", middleware.AuthMiddleware(roleuser), user.AddressStore)
	r.PATCH("/user/address/:ID", middleware.AuthMiddleware(roleuser), user.AddressEdit)
	r.DELETE("/user/address/:ID", middleware.AuthMiddleware(roleuser), user.AddressDelete)
	r.PATCH("/user/edit", middleware.AuthMiddleware(roleuser), user.EditUserProfile)

	//================= User cart ======================
	r.GET("/user/cart", middleware.AuthMiddleware(roleuser), user.CartView)
	r.POST("/user/cart/:ID", middleware.AuthMiddleware(roleuser), user.CartStore)
	r.PATCH("/user/cart/:ID/add", middleware.AuthMiddleware(roleuser), user.CartProductAdd)
	r.PATCH("/user/cart/:ID/remove", middleware.AuthMiddleware(roleuser), user.CartProductRemove)
	r.DELETE("/user/cart/:ID/delete", middleware.AuthMiddleware(roleuser), user.CartProductDelete)

	//============================= filter products ====================
	r.GET("/user/filter", user.SearchProduct)

}
