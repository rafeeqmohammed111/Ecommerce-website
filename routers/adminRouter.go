package routers

import (
	"project/admin"

	"project/middleware"

	"github.com/gin-gonic/gin"
)

func AdminRouter(r *gin.RouterGroup) {

	r.GET("/", admin.AdminPage)
	r.POST("/login", admin.AdminLogin)

	r.Use(middleware.AdminAuthMiddleware()) // Apply the middleware to all routes in this group

	r.POST("/logout", admin.AdminLogout)

	// User management
	r.GET("/user_management", admin.UserList)
	r.PATCH("/user_management/user_edit/:ID", admin.EditUserDetails)
	r.PATCH("/user_management/user_block/:ID", admin.BlockUser)
	r.DELETE("/user_management/user_delete/:ID", admin.DeleteUser)

	// Product management
	r.GET("/products", admin.ProductList)
	r.POST("/products/add_products", admin.AddProducts)
	r.PATCH("/products/edit_products/:ID", admin.EditProducts)
	r.DELETE("/products/delete_products/:ID", admin.DeleteProducts)
	r.POST("/upload", admin.UploadProductImage)

	// Category management
	r.GET("/categories", admin.CategoryList)
	r.POST("/categories/add_category", admin.AddCategory)
	r.PATCH("/categories/edit_category/:ID", admin.EditCategories)
	r.DELETE("/categories/delete_category/:ID", admin.DeleteCategories)
	r.PATCH("/categories/block_category/:ID", admin.BlockCategory)

	// **********order management***********
	r.GET("/orders", admin.AdminOrderView)
	r.PATCH("/ordercancel/:ID", admin.AdminCancelOrder)
	r.PATCH("/orderstatus/:ID", admin.AdminOrderStatus)

	//===================== Coupon managment ====================
	r.GET("/coupon", admin.CouponView)
	r.POST("/coupon", admin.CouponCreate)
	r.DELETE("/coupon/:ID", admin.CouponDelete)

	// =================== offers management =====================
	r.GET("/offer/:ID", admin.OfferShow)
	r.POST("/offer", admin.OfferAdd)
	r.DELETE("/offer/:ID", admin.OfferDelete)

	// ===================== sales report =========================
	r.GET("/sales/report", admin.GenerateReport)
	r.GET("/sales/report/excel", admin.SalesReportExcel)
	r.GET("/sales/report/pdf", admin.SalesReportPDF)

	r.GET("/bestselling", admin.BestSelling)


	// *************ledger******************
	r.GET("/ledger/monthly-sales-summary", admin.GetMonthlySalesSummary)

}
