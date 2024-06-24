package user

import (
	"log"
	"project/initializer"
	"project/models"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// User orders list show fetching from order table
// @Summary User Orders list
// @Description Fetch order table details and show the order list
// @Tags Order
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {json} SuccessResponse
// @Failure 400 {json} ErrorResponse
// @Router /orders [get]
// ============== list the orders to user ===============
func OrderView(c *gin.Context) {
	var orders []models.Order
	var orderShow []gin.H
	session := sessions.Default(c)
	userID := session.Get("user_id").(uint) // Assuming user_id is stored as uint in the session
	initializer.DB.Where("user_id=?", userID).Find(&orders)
	for _, v := range orders {
		orderShow = append(orderShow, gin.H{
			"orderId":       v.Id,
			"userId":        v.UserId,
			"addressId":     v.AddressId,
			"paymentMethod": v.OrderPaymentMethod,
			"orderAmount":   v.OrderAmount,
			"orderDate":     v.OrderDate,
		})
	}
	c.JSON(200, gin.H{
		"status": "success",
		"orders": orderShow,
	})
}

// Order placement response with details
// @Summary Place Order
// @Description Place an order and return order details
// @Tags Order
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param order body models.Order true "Order"
// @Success 200 {json} SuccessResponse
// @Failure 400 {json} ErrorResponse
// @Router /placeorder [post]
func PlaceOrder(c *gin.Context) {
	var order models.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(400, gin.H{
			"status": "Fail",
			"error":  "Invalid request payload",
			"code":   400,
		})
		return
	}

	tx := initializer.DB.Begin()

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to place order",
			"code":   500,
		})
		return
	}

	var couponDiscount float64
	if order.CouponCode != "" {
		var coupon models.Coupon
		if err := tx.First(&coupon, "code=?", order.CouponCode).Error; err == nil {
			couponDiscount = float64(coupon.Discount)
			if coupon.CouponCondition <= int(order.OrderAmount) {
				order.OrderAmount -= couponDiscount
				tx.Save(&order)
			}
		}
	}

	var orderItems []models.OrderItems
	if err := tx.Where("order_id = ?", order.Id).Find(&orderItems).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to fetch order items",
			"code":   500,
		})
		return
	}

	var productIds []uint
	for _, item := range orderItems {
		productIds = append(productIds, item.Product.ID)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to commit transaction",
			"code":   500,
		})
		return
	}

	c.JSON(200, gin.H{
		"status":                 "Success",
		"orderId":                order.Id,
		"productIds":             productIds,
		"couponAmount":           couponDiscount,
		"totalAmountAfterCoupon": order.OrderAmount,
	})
}

func CancelOrder(c *gin.Context) {
	orderItemId, err := strconv.Atoi(c.Param("orderItemId"))
	if err != nil {
		c.JSON(400, gin.H{
			"status": "Fail",
			"error":  "Invalid order item ID",
			"code":   400,
		})
		return
	}
	reason := c.PostForm("reason")
	// session := sessions.Default(c)
	// userID := session.Get("user_id").(uint)

	tx := initializer.DB.Begin()
	if reason == "" {
		c.JSON(402, gin.H{
			"status":  "Fail",
			"message": "Please provide a valid cancellation reason.",
			"code":    402,
		})
		return
	}

	// Fetch the order item using the order item ID
	var orderItem models.OrderItems
	if err := tx.First(&orderItem, orderItemId).Error; err != nil {
		log.Println("Error fetching order item:", err)
		c.JSON(404, gin.H{
			"status": "Fail",
			"error":  "Can't find order item",
			"code":   404,
		})
		tx.Rollback()
		return
	}

	// Fetch the corresponding order
	var order models.Order
	if err := tx.First(&order, orderItem.OrderId).Error; err != nil {
		log.Println("Error fetching order:", err)
		c.JSON(404, gin.H{
			"status": "Fail",
			"error":  "Can't find order",
			"code":   404,
		})
		tx.Rollback()
		return
	}

	// Update the status of the order item to "cancelled"
	if orderItem.OrderStatus == "cancelled" {
		c.JSON(202, gin.H{
			"status":  "Fail",
			"message": "Product already cancelled",
			"code":    202,
		})
		return
	}
	orderItem.OrderStatus = "cancelled"
	orderItem.OrderCancelReason = reason
	if err := tx.Save(&orderItem).Error; err != nil {
		log.Println("Error saving order item:", err)
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to save changes to database.",
			"code":   500,
		})
		tx.Rollback()
		return
	}

	// Fetch the product price
	var product models.Products
	if err := tx.First(&product, orderItem.ProductId).Error; err != nil {
		log.Println("Error fetching product:", err)
		c.JSON(404, gin.H{
			"status": "Fail",
			"error":  "Can't find product",
			"code":   404,
		})
		tx.Rollback()
		return
	}

	// Calculate the cancel amount and update the order's total amount
	cancelAmount := orderItem.SubTotal
	if float64(product.Price) > cancelAmount {
		cancelAmount = float64(product.Price)
	}

	order.OrderAmount -= cancelAmount

	// Check coupon condition
	var couponDiscount float64
	if order.CouponCode != "" {
		var coupon models.Coupon
		if err := initializer.DB.First(&coupon, "code=?", order.CouponCode).Error; err == nil {
			couponDiscount = float64(coupon.Discount)
			if coupon.CouponCondition <= int(order.OrderAmount) {
				order.OrderAmount -= couponDiscount
				tx.Save(&order)
			}
		}
	}

	if err := tx.Save(&order).Error; err != nil {
		log.Println("Error saving order:", err)
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to update order details",
			"code":   500,
		})
		tx.Rollback()
		return
	}

	// Update wallet balance if necessary
	if order.OrderPaymentMethod == "online" {
		var walletUpdate models.Wallet
		if err := tx.First(&walletUpdate, "user_id=?", order.UserId).Error; err != nil {
			c.JSON(501, gin.H{
				"status": "Fail",
				"error":  "Failed to fetch wallet details",
				"code":   501,
			})
			tx.Rollback()
			return
		}

		walletUpdate.Balance += cancelAmount
		if err := tx.Save(&walletUpdate).Error; err != nil {
			log.Println("Error updating wallet:", err)
			c.JSON(500, gin.H{
				"status": "Fail",
				"error":  "Failed to update wallet balance.",
				"code":   500,
			})
			tx.Rollback()
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(201, gin.H{
			"status":  "Fail",
			"message": "Failed to commit transaction",
			"code":    201,
		})
		tx.Rollback()
	} else {
		c.JSON(201, gin.H{
			"status":  "Success",
			"message": "Order Item Cancelled",
			"data":    "cancelled",
		})
	}
}

func UserOrderStatus(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id").(uint) 

	var orders []models.Order
	if err := initializer.DB.Where("user_id = ?", userID).Find(&orders).Error; err != nil {
		c.JSON(400, gin.H{
			"status": "Fail",
			"error":  "Failed to fetch orders",
			"code":   400,
		})
		return
	}

	var orderStatuses []gin.H
	for _, order := range orders {
		var orderItems []models.OrderItems
		if err := initializer.DB.Where("order_id = ?", order.Id).Find(&orderItems).Error; err != nil {
			c.JSON(400, gin.H{
				"status": "Fail",
				"error":  "Failed to fetch order items",
				"code":   400,
			})
			return
		}

		for _, item := range orderItems {
			orderStatuses = append(orderStatuses, gin.H{
				"orderId":      order.Id,
				"itemId":       item.Id,
				"productId":    item.ProductId,
				"quantity":     item.Quantity,
				"subTotal":     item.SubTotal,
				"orderStatus":  item.OrderStatus,
				"cancelReason": item.OrderCancelReason,
			})
		}

		couponDiscount := 0
		if order.CouponCode != "" {
			var coupon models.Coupon
			if err := initializer.DB.First(&coupon, "code = ?", order.CouponCode).Error; err == nil {
				couponDiscount = int(coupon.Discount)
			}
		}

		orderStatuses = append(orderStatuses, gin.H{
			"orderId":                order.Id,
			"userId":                 order.UserId,
			"addressId":              order.AddressId,
			"paymentMethod":          order.OrderPaymentMethod,
			"orderAmount":            order.OrderAmount,
			"couponCode":             order.CouponCode,
			"couponDiscount":         couponDiscount,
			"totalAmountAfterCoupon": order.OrderAmount - float64(couponDiscount),
			"orderDate":              order.OrderDate,
		})
	}

	c.JSON(200, gin.H{
		"status": "success",
		"orders": orderStatuses,
	})
}
