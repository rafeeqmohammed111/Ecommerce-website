package user

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"project/initializer"
	"project/models"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	// "github.com/google/uuid"
)

var razorpayPaymentID string

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
	if paymentMethod == "COD" && totalAmount > 1000 {
		c.JSON(202, gin.H{
			"status":         "Fail",
			"message":        "Orders greater than 1000 rupees are not eligible for COD",
			"totalAmount":    totalAmount,
			"shippingCharge": shippingCharge,
			"code":           202,
		})
		return
	}

	// Generate unique order ID
	const charset = "123456789"
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
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
	orderID := string(randomBytes)

	tx := initializer.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create  order
	order := models.Order{
		Id:                 orderID,
		UserId:             userID,
		OrderPaymentMethod: paymentMethod,
		AddressId:          addressID,
		OrderAmount:        totalAmount,
		ShippingCharge:     float32(shippingCharge),
		OrderDate:          time.Now(),
		CouponCode:         couponCode,
		PaymentStatus:      "pending",
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

	// Create order items ,updte prdct stock
	for _, val := range cartItems {
		orderItem := models.OrderItems{
			OrderId:     orderID,
			ProductId:   val.ProductId,
			Quantity:    val.Quantity,
			SubTotal:    float64(val.Product.Price) * float64(val.Quantity),
			OrderStatus: "pending",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			ProductName: val.Product.Name,
			Category:    val.Product.Category.Category_name,
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

		// Manage stock COD
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

	// Delete  cart itms
	if err := tx.Where("user_id =?", userID).Delete(&models.Cart{}).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to delete data in cart",
			"code":   500,
		})
		return
	}

	// Commit
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to commit transaction",
			"code":   500,
		})
		return
	}

	// payment
	if paymentMethod == "ONLINE" {
		paymentDetails := models.PaymentDetails{
			PaymentAmount: int(totalAmount),
			PaymentStatus: "pending",
		}
		initializer.DB.Create(&paymentDetails)

		razorId, err := PaymentHandler(orderID, totalAmount)

		fmt.Println("*****************", razorId)

		razorpayPaymentID = razorId

		if err != nil {
			initializer.DB.Model(&paymentDetails).Update("PaymentStatus", "faild")
			c.JSON(400, gin.H{
				"status": "Fail",
				"error":  err,
			})
			return
		}

		receiptID := generateReceiptID()
		paymentUpdate := models.PaymentDetails{
			OrderID:       razorId,
			PaymentAmount: int(totalAmount),
			Receipt:       uint(receiptID),
			PaymentStatus: "Pending",
		}
		initializer.DB.Create(&paymentUpdate)

		c.JSON(200, gin.H{
			"status":         "Success",
			"message":        "Please complete the payment",
			"orderId":        razorId,
			"totalAmount":    totalAmount,
			"shippingCharge": shippingCharge,
		})
		return
	}

	// COD***
	c.JSON(200, gin.H{
		"status":         "Success",
		"message":        "Order placed successfully. Order will arrive within 4 days.",
		"payment":        "COD",
		"totalAmount":    totalAmount,
		"shippingCharge": shippingCharge,
		"discount":       discountAmount,
	})
}

// receipt ID***
func generateReceiptID() uint {
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return 0
	}
	return uint(time.Now().UnixNano())
}

func OrderDetails(c *gin.Context) {
	orderID := c.Param("ID")

	var order models.Order
	var orderItems []models.OrderItems
	var paymentDetails models.PaymentDetails

	// Fetch order details***
	if err := initializer.DB.Where("id = ?", orderID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "Fail",
			"error":  "Order not found",
			"code":   http.StatusNotFound,
		})
		return
	}

	// Fetch order items***
	if err := initializer.DB.Where("order_id = ?", orderID).Preload("Product").Preload("Order").Find(&orderItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Fail",
			"error":  "Failed to fetch order items",
			"code":   http.StatusInternalServerError,
		})
		return
	}

	paymentStatus := "success" // faild to success

	fmt.Println("***********1", paymentStatus)

	if err := initializer.DB.Where("order_id = ?", razorpayPaymentID).First(&paymentDetails).Error; err == nil {
		
		if paymentDetails.PaymentId != "" {

			paymentStatus = "Success"
			fmt.Println("***********2", paymentStatus)
		}
	}

	// Calculate total amount before discount
	var totalAmountBeforeDiscount float64
	for _, item := range orderItems {
		totalAmountBeforeDiscount += item.SubTotal
	}

	totalDiscount := totalAmountBeforeDiscount - order.OrderAmount

	// Prepare order details
	orderDetails := gin.H{
		"userId":                   order.UserId,
		"orderAmount":              totalAmountBeforeDiscount,
		"couponCode":               order.CouponCode,
		"totalAmountAfterDiscount": order.OrderAmount,
		"orderDate":                order.OrderDate,
		"totalDiscount":            totalDiscount,
		"paymentStatus":            paymentStatus,
		"items":                    []gin.H{},
	}

	// Add order items to response
	for _, item := range orderItems {
		orderDetails["items"] = append(orderDetails["items"].([]gin.H), gin.H{
			"orderItemId": item.Id,
			"productId":   item.ProductId,
			"productName": item.Product.Name,
			"orderDate":   item.Order.OrderDate,
			"amount":      item.SubTotal,
			"quantity":    item.Quantity,
			"orderStatus": item.OrderStatus,
			"addressId":   item.Order.AddressId,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       orderDetails,
		"paymentStatus": paymentStatus,
	})
}
