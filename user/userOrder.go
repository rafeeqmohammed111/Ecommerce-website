package user

import (
	"fmt"
	"log"
	"net/http"
	"project/initializer"
	"project/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	// "gorm.io/gorm"
	// "gorm.io/gorm"
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
	userID := session.Get("user_id").(uint)
	initializer.DB.Where("user_id=?", userID).Find(&orders)
	for _, v := range orders {
		orderShow = append(orderShow, gin.H{
			"orderId":       v.Id,
			"userId":        v.UserId,
			"addressId":     v.AddressId,
			"paymentMethod": v.OrderPaymentMethod,
			"orderAmount":   v.OrderAmount,
			"orderDate":     v.OrderDate,
			// "paymentStatus": v.PaymentStatus,
		})
	}
	c.JSON(200, gin.H{
		"status": "success",
		"orders": orderShow,
	})
}

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

	order.PaymentStatus = "pending"

	if err := initializer.DB.Create(&order).Error; err != nil {
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
		if err := initializer.DB.First(&coupon, "code=?", order.CouponCode).Error; err == nil {
			couponDiscount = float64(coupon.Discount)
			if coupon.CouponCondition <= int(order.OrderAmount) {
				order.OrderAmount -= couponDiscount
				initializer.DB.Save(&order)
			}
		}
	}

	var orderItems []models.OrderItems
	if err := initializer.DB.Where("order_id = ?", order.Id).Find(&orderItems).Error; err != nil {
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

	c.JSON(200, gin.H{
		"status":                 "Success",
		"orderId":                order.Id,
		"productIds":             productIds,
		"couponAmount":           couponDiscount,
		"totalAmountAfterCoupon": order.OrderAmount,
	})
}

func CancelOrder(c *gin.Context) {
	orderID := c.PostForm("orderId")
	itemId := c.PostForm("itemId")
	reason := c.PostForm("reason")

	if orderID == "" && itemId == "" {
		c.JSON(400, gin.H{
			"status": "Fail",
			"error":  "Either orderId or itemId must be provided",
			"code":   400,
		})
		return
	}

	var orderItem models.OrderItems
	if orderID != "" {
		var order models.Order
		if err := initializer.DB.First(&order, "id = ?", orderID).Error; err != nil {
			log.Println("Error fetching order:", err)
			c.JSON(404, gin.H{
				"status": "Fail",
				"error":  "Order not found",
				"code":   404,
			})
			return
		}
		if err := initializer.DB.First(&orderItem, "order_id = ?", orderID).Error; err != nil {
			log.Println("Error fetching order item:", err)
			c.JSON(404, gin.H{
				"status": "Fail",
				"error":  "Order item not found",
				"code":   404,
			})
			return
		}
	} else if itemId != "" {
		if err := initializer.DB.First(&orderItem, "id = ?", itemId).Error; err != nil {
			log.Println("Error fetching order item:", err)
			c.JSON(404, gin.H{
				"status": "Fail",
				"error":  "Order item not found",
				"code":   404,
			})
			return
		}
	}

	if orderItem.OrderStatus == "cancelled" {
		c.JSON(202, gin.H{
			"status":  "Fail",
			"message": "Order item already cancelled",
			"code":    202,
		})
		return
	}

	orderItem.OrderStatus = "cancelled"
	orderItem.OrderCancelReason = reason

	if err := initializer.DB.Save(&orderItem).Error; err != nil {
		log.Println("Error saving order item:", err)
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to save changes to order item",
			"code":   500,
		})
		return
	}

	var order models.Order
	if err := initializer.DB.First(&order, orderItem.OrderId).Error; err != nil {
		log.Println("Error fetching order:", err)
		c.JSON(404, gin.H{
			"status": "Fail",
			"error":  "Order not found",
			"code":   404,
		})
		return
	}

	cancelAmount := orderItem.SubTotal
	fmt.Println("========================================", cancelAmount)

	if order.OrderAmount > cancelAmount {
		order.OrderAmount -= cancelAmount

		var couponDiscount float64
		if order.CouponCode != "" {
			var coupon models.Coupon
			if err := initializer.DB.First(&coupon, "code=?", order.CouponCode).Error; err == nil {
				couponDiscount = float64(coupon.Discount)
				if coupon.CouponCondition <= int(order.OrderAmount) {
					order.OrderAmount -= couponDiscount
				}
			}
		}

		if err := initializer.DB.Save(&order).Error; err != nil {
			log.Println("Error saving order:", err)
			c.JSON(500, gin.H{
				"status": "Fail",
				"error":  "Failed to update order details",
				"code":   500,
			})
			return
		}
	}

	if order.OrderPaymentMethod == "ONLINE" {
		var wallet models.Wallet
		userID := order.UserId

		if err := initializer.DB.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			wallet = models.Wallet{
				UserId:  userID,
				Balance: 0,
			}
			if err := initializer.DB.Create(&wallet).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create the wallet"})
				return
			}
		}

		fmt.Println("Initial wallet balance:", wallet.Balance)
		fmt.Println("Cancel amount to be added:", cancelAmount)

		wallet.Balance += cancelAmount

		if err := initializer.DB.Save(&wallet).Error; err != nil {
			log.Println("Error updating wallet:", err)
			c.JSON(500, gin.H{
				"status": "Fail",
				"error":  "Failed to update wallet balance",
				"code":   500,
			})
			return
		}

		fmt.Println("Updated wallet balance:", wallet.Balance)
	}

	c.JSON(200, gin.H{
		"status":           "Success",
		"message":          "Order item cancelled successfully",
		"cancelled amount": cancelAmount,
	})
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
			"staus":                  "pending",
		})
	}

	c.JSON(200, gin.H{
		"status": "success",
		"orders": orderStatuses,
	})
}

// func OrderDetails(c *gin.Context) {
// 	orderID := c.Param("ID")

// 	var order models.Order
// 	var orderItems []models.OrderItems
// 	var paymentDetails models.PaymentDetails

// 	// Fetch the order details
// 	if err := initializer.DB.Where("id = ?", orderID).First(&order).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{
// 			"status": "Fail",
// 			"error":  "Order not found",
// 			"code":   http.StatusNotFound,
// 		})
// 		return
// 	}

// 	// Fetch the order items
// 	if err := initializer.DB.Where("order_id = ?", orderID).Find(&orderItems).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"status": "Fail",
// 			"error":  "Failed to fetch order items",
// 			"code":   http.StatusInternalServerError,
// 		})
// 		return
// 	}

// 	// Fetch the payment details
// 	paymentStatus := "pending"
// 	if err := initializer.DB.Where("order_id = ?", orderID).First(&paymentDetails).Error; err != nil {
// 		if err != gorm.ErrRecordNotFound {
// 			c.JSON(http.StatusInternalServerError, gin.H{
// 				"status": "Fail",
// 				"error":  "Failed to fetch payment details",
// 				"code":   http.StatusInternalServerError,
// 			})
// 			return
// 		}
// 	} else {
// 		paymentStatus = paymentDetails.PaymentStatus
// 	}

// 	// Calculate total amount before discount
// 	var totalAmountBeforeDiscount float64
// 	for _, item := range orderItems {
// 		totalAmountBeforeDiscount += item.SubTotal
// 	}

// 	totalDiscount := totalAmountBeforeDiscount - order.OrderAmount

// 	// Prepare the order details response
// 	orderDetails := gin.H{
// 		"userId":                   order.UserId,
// 		"orderAmount":              totalAmountBeforeDiscount,
// 		"couponCode":               order.CouponCode,
// 		"totalAmountAfterDiscount": order.OrderAmount,
// 		"orderDate":                order.OrderDate,
// 		"totalDiscount":            totalDiscount,
// 		"paymentStatus":            paymentStatus,
// 		"items":                    []gin.H{},
// 	}

// 	// Add order items to the response
// 	for _, item := range orderItems {
// 		product := models.Products{}
// 		if err := initializer.DB.Where("id = ?", item.ProductId).First(&product).Error; err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{
// 				"status": "Fail",
// 				"error":  "Failed to fetch product details",
// 				"code":   http.StatusInternalServerError,
// 			})
// 			return
// 		}
// 		orderDetails["items"] = append(orderDetails["items"].([]gin.H), gin.H{
// 			"productId":   item.ProductId,
// 			"productName": product.Name,
// 			"itemId":      item.Id,
// 			"quantity":    item.Quantity,
// 			"status":      item.OrderStatus,
// 			"amount":      item.SubTotal,
// 		})
// 	}

// 	// Return the order details response
// 	c.JSON(http.StatusOK, gin.H{
// 		"status": "success",
// 		"order":  orderDetails,
// 	})
// }

// //order details

// // //new today

// // func RetryPayment(c *gin.Context) {
// // 	orderID := c.Query("order_id")

// // 	if orderID == "" {
// // 		c.JSON(400, gin.H{
// // 			"status": "Fail",
// // 			"error":  "Order ID is required",
// // 			"code":   400,
// // 		})
// // 		return
// // 	}

// // 	var order models.Order
// // 	if err := initializer.DB.Where("id = ?", orderID).Find(&order).Error; err != nil {
// // 		c.JSON(404, gin.H{
// // 			"status": "Fail",
// // 			"error":  "Order not found",
// // 			"code":   404,
// // 		})
// // 		return
// // 	}

// // 	if order.PaymentStatus != "pending" {
// // 		c.JSON(400, gin.H{
// // 			"status": "Fail",
// // 			"error":  "Payment is not pending for this order",
// // 			"code":   400,
// // 		})
// // 		return
// // 	}

// // 	// Call the payment handler to retry the payment
// // 	rzpOrderId, err := PaymentHandler(order.Id, order.OrderAmount)
// // 	if err != nil {
// // 		// Log the error and return a failure response
// // 		log.Println("Error processing payment:", err)
// // 		c.JSON(500, gin.H{
// // 			"status": "Fail",
// // 			"error":  "Failed to process payment",
// // 			"code":   500,
// // 		})
// // 		return
// // 	}

// // 	// Assuming the PaymentHandler returns without error upon successful payment
// // 	order.PaymentStatus = "payment success"
// // 	if err := initializer.DB.Save(&order).Error; err != nil {
// // 		c.JSON(500, gin.H{
// // 			"status": "Fail",
// // 			"error":  "Failed to update order status",
// // 			"code":   500,
// // 		})
// // 		return
// // 	}

// // 	c.JSON(200, gin.H{
// // 		"status":  "success",
// // 		"orderId": rzpOrderId,
// // 	})
// // }
