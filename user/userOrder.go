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

// Order cancelation using order id and update other details and status
// @Summary Order cancel
// @Description Order cancel and update the status , other details
// @Tags Order
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "orderItems order ID"
// @Param reason formData string true "Cancelation reason?"
// @Success 200 {json} SuccessResponse
// @Failure 400 {json} ErrorResponse
// @Router /ordercancel/{id} [patch]
func CancelOrder(c *gin.Context) {
	orderId, err := strconv.Atoi(c.Param("ID"))
	if err != nil {
		c.JSON(400, gin.H{
			"status": "Fail",
			"error":  "Invalid order ID",
			"code":   400,
		})
		return
	}
	reason := c.PostForm("reason")
	session := sessions.Default(c)
	userID := session.Get("user_id").(uint) // Assuming user_id is stored as uint in the session

	tx := initializer.DB.Begin()
	if reason == "" {
		c.JSON(402, gin.H{
			"status":  "Fail",
			"message": "Please provide a valid cancellation reason.",
			"code":    402,
		})
		return
	}

	// Fetch the order using the order ID
	var order models.Order
	if err := tx.First(&order, orderId).Error; err != nil {
		log.Println("Error fetching order:", err)
		c.JSON(404, gin.H{
			"status": "Fail",
			"error":  "can't find order",
			"code":   404,
		})
		tx.Rollback()
		return
	}

	// Check if the order belongs to the logged-in user
	if order.UserId != int(userID) {
		c.JSON(403, gin.H{
			"status": "Fail",
			"error":  "You are not authorized to cancel this order.",
			"code":   403,
		})
		tx.Rollback()
		return
	}

	// Fetch all order items associated with this order
	var orderItems []models.OrderItems
	if err := tx.Where("order_id = ?", orderId).Find(&orderItems).Error; err != nil {
		log.Println("Error fetching order items:", err)
		c.JSON(404, gin.H{
			"status": "Fail",
			"error":  "can't find order items",
			"code":   404,
		})
		tx.Rollback()
		return
	}

	// Update the status of each order item to "cancelled"
	for i := range orderItems {
		if orderItems[i].OrderStatus == "cancelled" {
			c.JSON(202, gin.H{
				"status":  "Fail",
				"message": "product already cancelled",
				"code":    202,
			})
			return
		}
		orderItems[i].OrderStatus = "cancelled"
		orderItems[i].OrderCancelReason = reason
		if err := tx.Save(&orderItems[i]).Error; err != nil {
			log.Println("Error saving order item:", err)
			c.JSON(500, gin.H{
				"status": "Fail",
				"error":  "Failed to save changes to database.",
				"code":   500,
			})
			tx.Rollback()
			return
		}
	}

	//========== check coupon condition ============
	// var couponRemove models.Coupon
	// if order.CouponCode != "" {
	// 	if err := initializer.DB.First(&couponRemove, "code=?", order.CouponCode).Error; err != nil {
	// 		c.JSON(404, gin.H{
	// 			"status": "Fail",
	// 			"error":  "can't find coupon code",
	// 			"code":   404,
	// 		})
	// 		tx.Rollback()
	// 	}
	// }
	// if couponRemove.CouponCondition > int(order.OrderAmount) {
	// 	order.OrderAmount += couponRemove.Discount
	// 	order.OrderAmount -= float64(orderItems[i].SubTotal)
	// 	order.CouponCode = ""
	// }

	if err := tx.Save(&order).Error; err != nil {
		log.Println("Error saving order:", err)
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "failed to update order details",
			"code":   500,
		})
		tx.Rollback()
		return
	}

	// var walletUpdate models.Wallet
	// if err := tx.First(&walletUpdate, "user_id = ?", order.UserId).Error; err != nil {
	// 	log.Println("Error fetching wallet:", err)
	// 	c.JSON(501, gin.H{
	// 		"status": "Fail",
	// 		"error":  "failed to fetch wallet details",
	// 		"code":   501,
	// 	})
	// 	tx.Rollback()
	// 	return
	// } else {
	// 	walletUpdate.Balance += order.OrderAmount
	// 	if err := tx.Save(&walletUpdate).Error; err != nil {
	// 		log.Println("Error saving wallet:", err)
	// 		c.JSON(500, gin.H{
	// 			"status": "Fail",
	// 			"error":  "failed to update wallet balance",
	// 			"code":   500,
	// 		})
	// 		tx.Rollback()
	// 		return
	// 	}
	// }

	if err := tx.Commit().Error; err != nil {
		c.JSON(201, gin.H{
			"status":  "Fail",
			"message": "failed to commit transaction",
			"code":    201,
		})
		tx.Rollback()
	} else {
		c.JSON(201, gin.H{
			"status":  "Success",
			"message": "Order Cancelled",
			"data":    "cancelled",
		})
	}
}
