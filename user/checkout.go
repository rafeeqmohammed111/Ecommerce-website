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
	// "github.com/google/uuid"
)

// CheckOut handles the checkout process for placing an order
func CheckOut(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id").(uint)
	var cartItems []models.Cart
	initializer.DB.Preload("Product").Where("user_id=?", userID).Find(&cartItems)
	if len(cartItems) == 0 {
		c.JSON(404, gin.H{
			"status":  "Fail",
			"message": "Please add some items to your cart first.",
			"code":    404,
		})
		return
	}

	// Check if given payment method and address are provided
	paymentMethod := c.Request.PostFormValue("payment")
	addressID, err := strconv.Atoi(c.Request.PostFormValue("address"))
	if err != nil || addressID == 0 {
		c.JSON(400, gin.H{
			"status": "Fail",
			"error":  "Invalid Address ID or Payment Method",
			"code":   400,
		})
		return
	}

	// Stock check and amount calculation
	var totalAmount float64
	for _, val := range cartItems {
		amount := float64(val.Product.Price) * float64(val.Quantity)
		if val.Quantity > uint(val.Product.Quantity) {
			c.JSON(400, gin.H{
				"status": "Fail",
				"error":  fmt.Sprintf("Insufficient stock for product %s", val.Product.Name),
				"code":   400,
			})
			return
		}
		totalAmount += amount
	}

	// Coupon validation
	var discountAmount float64
	couponCode := c.Request.FormValue("coupon")
	if couponCode != "" {
		var couponCheck models.Coupon
		if err := initializer.DB.Where("code=? AND valid_from < ? AND valid_to > ? AND coupon_condition <= ?", couponCode, time.Now(), time.Now(), totalAmount).First(&couponCheck).Error; err != nil {
			c.JSON(200, gin.H{
				"error": "Coupon not valid",
			})
			return
		}
		discountAmount = couponCheck.Discount
		totalAmount -= discountAmount
	}

	// Delivery charges
	var shippingCharge float64
	if totalAmount < 1000 {
		shippingCharge = 40
		totalAmount += shippingCharge
	}

	// COD checking
	if paymentMethod == "COD" {
		if totalAmount > 1000 {
			c.JSON(202, gin.H{
				"status":      "Fail",
				"message":     "Orders greater than 1000 rupees are not eligible for COD",
				"totalAmount": totalAmount,
				"code":        202,
			})
			return
		}
	}

	// Generate a unique order ID using UUID and random numeric string
	const charset = "123456789"
	randomBytes := make([]byte, 8)
	_, err = rand.Read(randomBytes)
	if err != nil {
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to generate order ID",
			"code":   500,
		})
		return
	}
	for i, b := range randomBytes {
		randomBytes[i] = charset[b%byte(len(charset))]
	}
	numericPart := string(randomBytes)
	orderID := numericPart

	// Start the transaction
	tx := initializer.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Handle online payment
	var rzpOrderId string
	if paymentMethod == "ONLINE" {
		fmt.Println("order id : ", orderID)
		fmt.Println("total amount : ", totalAmount)
		rzpOrderId, err = PaymentHandler(orderID, totalAmount)
		if err != nil {
			c.JSON(500, gin.H{
				"status": "Fail",
				"error":  fmt.Sprintf("Failed to create orderId: %v", err),
				"code":   500,
			})
			tx.Rollback()
			return
		}
		fmt.Println("razor pay id : ", rzpOrderId)
	}

	// Insert order details into the database
	order := models.Order{
		Id:                 orderID,
		UserId:             int(userID),
		OrderPaymentMethod: paymentMethod,
		AddressId:          addressID,
		OrderAmount:        totalAmount,
		ShippingCharge:     float32(shippingCharge),
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

	// Insert order items into the database
	for _, val := range cartItems {
		orderItem := models.OrderItems{
			OrderId:     orderID,
			ProductId:   val.ProductId,
			Quantity:    val.Quantity,
			SubTotal:    float64(val.Product.Price) * float64(val.Quantity),
			OrderStatus: "pending",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{
				"status": "Fail",
				"error":  "Failed to store item details",
				"code":   500,
			})
			return
		}

		// Manage the stock for COD
		var product models.Products
		tx.First(&product, val.ProductId)
		product.Quantity -= int(val.Quantity)
		if err := tx.Save(&product).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{
				"status": "Fail",
				"error":  "Failed to update product stock",
				"code":   500,
			})
			return
		}
	}

	// ****Delete all items from user cart****
	if err := tx.Where("user_id =?", userID).Delete(&models.Cart{}).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to delete data in cart",
			"code":   500,
		})
		return
	}

	// ****Commit transaction if no error****
	if err := tx.Commit().Error; err != nil {
		// tx.Roll
		tx.Rollback()
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to commit transaction",
			"code":   500,
		})
		return
	}

	// Success Response
	if paymentMethod == "COD" {
		c.JSON(200, gin.H{
			"status":      "Success",
			"message":     "Order placed successfully. Order will arrive within 4 days.",
			"payment":     "COD",
			"totalAmount": totalAmount,
			"discount":    discountAmount,
		})
	} else if paymentMethod == "ONLINE" {
		c.JSON(200, gin.H{
			"status":      "Success",
			"message":     "Please complete the payment",
			"totalAmount": totalAmount,
			"orderId":     rzpOrderId,
			"discount":    discountAmount,
		})
	}
}
