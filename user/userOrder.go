package user

import (
	"project/initializer"
	"project/models"

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
			"userName":      v.UserId,
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
	var orderItem models.OrderItems
	orderItemId := c.Param("ID")
	reason := c.Request.FormValue("reason")
	session := sessions.Default(c)
	userID := session.Get("user_id").(uint) // Assuming user_id is stored as uint in the session
	tx := initializer.DB.Begin()
	if reason == "" {
		c.JSON(402, gin.H{
			"status":  "Fail",
			"message": "Please provide a valid cancellation reason.",
			"code":    402,
		})
	} else {
		if err := tx.First(&orderItem, orderItemId).Error; err != nil {
			c.JSON(404, gin.H{
				"status": "Fail",
				"error":  "can't find order",
				"code":   404,
			})
			tx.Rollback()
			return
		}
		// Check if the order belongs to the logged-in user
		if orderItem.Order.UserId != int(userID) {
			c.JSON(403, gin.H{
				"status": "Fail",
				"error":  "You are not authorized to cancel this order.",
				"code":   403,
			})
			tx.Rollback()
			return
		}
		if orderItem.OrderStatus == "cancelled" {
			c.JSON(202, gin.H{
				"status":  "Fail",
				"message": "product already cancelled",
				"code":    202,
			})
			return
		}
		// ======= update status as cancelled ======
		orderItem.OrderStatus = "cancelled"
		orderItem.OrderCancelReason = reason
		if err := tx.Save(&orderItem).Error; err != nil {
			c.JSON(500, gin.H{
				"status": "Fail",
				"error":  "Failed to  save changes to database.",
				"code":   500,
			})
			tx.Rollback()
			return
		}

		var orderAmount models.Order
		if err := tx.First(&orderAmount, orderItem.OrderId).Error; err != nil {
			c.JSON(404, gin.H{
				"status": "Fail",
				"error":  "failed to find order details",
				"code":   404,
			})
			tx.Rollback()
			return
		}
		//========== check coupon condition ============
		var couponRemove models.Coupon
		if orderAmount.CouponCode != "" {
			if err := initializer.DB.First(&couponRemove, "code=?", orderAmount.CouponCode).Error; err != nil {
				c.JSON(404, gin.H{
					"status": "Fail",
					"error":  "can't find coupon code",
					"code":   404,
				})
				tx.Rollback()
			}
		}
		if couponRemove.CouponCondition > int(orderAmount.OrderAmount) {
			orderAmount.OrderAmount += couponRemove.Discount
			orderAmount.OrderAmount -= float64(orderItem.SubTotal)
			orderAmount.CouponCode = ""
		}
		if err := tx.Save(&orderAmount).Error; err != nil {
			c.JSON(500, gin.H{
				"status": "Fail",
				"error":  "failed to update order details",
				"code":   500,
			})
			tx.Rollback()
			return
		}
		var walletUpdate models.Wallet
		if err := tx.First(&walletUpdate, "user_id=?", orderAmount.UserId).Error; err != nil {
			c.JSON(501, gin.H{
				"status": "Fail",
				"error":  "failed to fetch wallet details",
				"code":   501,
			})
			tx.Rollback()
			return
		} else {
			walletUpdate.Balance += orderAmount.OrderAmount
			tx.Save(&walletUpdate)
		}
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
				"data":    orderItem.OrderStatus,
			})
		}
	}
}
