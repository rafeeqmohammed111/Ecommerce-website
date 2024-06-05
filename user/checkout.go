package user

import (
	"crypto/rand"
	"fmt"
	"project/initializer"
	"project/models"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Place the order based on cart items with given payment system and shipping method
// @Summery  place an order
// @Description place an order by given cart items, calculate total price of all products in the shopping cart, generate a unique OrderID, place the order with given payment method
// @Tags  Order
// @Accept multipart/form-data
// @Produce  json
// @Secure ApiKeyAuth
// @Param payment formData string true "Payment Method"
// @Param address formData string true  "Address ID"
// @Param coupon formData string false "Coupon Code"
// @Success 200 {json} SuccessResponse
// @Failure 400 {json} ErrorResponse
// @Router /checkout [post]
func CheckOut(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id").(uint) // Assuming user_id is stored as uint in the session

	couponCode := ""
	var cartItems []models.Cart
	initializer.DB.Preload("Product").Where("user_id=?", userID).Find(&cartItems)
	if len(cartItems) == 0 {
		c.JSON(404, gin.H{
			"status":  "Fail",
			"message": "please add some items to your cart firstly.",
			"code":    404,
		})
		return
	}
	// ============= check if given payment method and address =============
	paymentMethod := c.Request.PostFormValue("payment")
	Address, _ := strconv.ParseUint(c.Request.PostFormValue("address"), 10, 64)
	if paymentMethod == "" || Address == 0 {
		c.JSON(400, gin.H{
			"status": "Fail",
			"error":  "Payment Method and Address are required",
			"code":   400,
		})
		return
	}
	if paymentMethod != "COD" {
		c.JSON(400, gin.H{
			"status": "Fail",
			"error":  "Only COD payment method is allowed",
			"code":   400,
		})
		return
	}

	// ============= stock check and amount calc ===================
	var totalAmount float64
	for _, val := range cartItems {
		Amount := float64(val.Product.Price) * float64(val.Quantity)
		if val.Quantity > uint(val.Product.Quantity) {
			c.JSON(400, gin.H{
				"status": "Fail",
				"error":  "Insufficient stock for product " + val.Product.Name,
				"code":   400,
			})
			return
		}
		totalAmount += Amount
	}

	// ================== coupon validation ===============
	couponCode = c.Request.FormValue("coupon")
	var couponCheck models.Coupon
	var userLimitCheck models.Order
	if couponCode != "" {
		if err := initializer.DB.First(&userLimitCheck, "coupon_code", couponCode).Error; err == nil {
			c.JSON(409, gin.H{
				"status": "Fail",
				"error":  "Coupon already used",
				"code":   409,
			})
			return
		}
		if err := initializer.DB.Where("code=? AND valid_from < ? AND valid_to > ? AND coupon_condition <= ?", couponCode, time.Now(), time.Now(), totalAmount).First(&couponCheck).Error; err != nil {
			c.JSON(200, gin.H{
				"error": "Coupon Not valid",
			})
			return
		} else {
			totalAmount -= couponCheck.Discount
		}
	}

	// ================== order id creation =======================
	const charset = "123456789"
	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		fmt.Println(err)
	}
	var orderIdString string
	for _, b := range randomBytes {
		orderIdString += string(charset[b%byte(len(charset))])
	}
	orderID, _ := strconv.Atoi(orderIdString)

	//================ Start the transaction ===================
	tx := initializer.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// ============== Delivery charges ==============
	var ShippingCharge float64
	if totalAmount < 1000 {
		ShippingCharge = 40
		totalAmount += ShippingCharge
	}
	// ============== COD checking =====================
	if totalAmount > 1000 {
		c.JSON(202, gin.H{
			"status":      "Fail",
			"message":     "Greater than 1000 rupees should not accept COD",
			"totalAmount": totalAmount,
			"code":        202,
		})
		return
	}

	// ================= insert order details into database ===================
	order := models.Order{
		Id:                 uint(orderID),
		UserId:             int(userID),
		OrderPaymentMethod: paymentMethod,
		AddressId:          int(Address),
		OrderAmount:        totalAmount,
		ShippingCharge:     float32(ShippingCharge),
		OrderDate:          time.Now(),
		CouponCode:         couponCode,
	}
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to place order",
			"code":   500,
		})
		return
	}

	// ============ insert order items into database ==================
	for _, val := range cartItems {
		OrderItems := models.OrderItems{
			OrderId:     uint(orderID),
			ProductId:   val.ProductId,
			Quantity:    val.Quantity,
			SubTotal:    float64(val.Product.Price) * float64(val.Quantity),
			OrderStatus: "pending",
		}
		if err := tx.Create(&OrderItems).Error; err != nil {
			tx.Rollback()
			c.JSON(501, gin.H{
				"status": "Fail",
				"error":  "Failed to store items details",
				"code":   501,
			})
			return
		}
		// ============= manage the stock for COD ============
		var productQuantity models.Products
		tx.First(&productQuantity, val.ProductId)
		if err := tx.Save(val.Product).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{
				"status": "Fail",
				"error":  "Failed to Update Product Stock",
				"code":   500,
			})
			return
		}
	}

	// =============== delete all items from user cart ==============
	if err := tx.Where("user_id =?", userID).Delete(&models.Cart{}).Error; err != nil {
		tx.Rollback()
		c.JSON(400, gin.H{
			"status": "Fail",
			"error":  "Failed to delete data in cart",
			"code":   400,
		})
		return
	}

	//================= commit transaction whether no error ==================
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to commit transaction",
			"code":   500,
		})
		return
	}

	c.JSON(501, gin.H{
		"status":      "Success",
		"Order":       "Order Placed successfully",
		"payment":     "COD",
		"totalAmount": totalAmount,
		"message":     "Order will arrive within 4 days",
	})
}
//  offer discount and online payment methord logics has to impliment later 