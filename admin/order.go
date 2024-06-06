package admin

import (
	"log"
	"net/http"
	"project/initializer"
	"project/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AdminOrderView(c *gin.Context) {
	var orderItems []models.OrderItems
	var orderShow []gin.H

	if err := initializer.DB.Preload("Order").Find(&orderItems).Error; err != nil {
		c.JSON(404, gin.H{
			"status": "Fail",
			"error":  "can't find the orders",
			"code":   "404",
		})
		return
	}

	for _, v := range orderItems {
		orderShow = append(orderShow, gin.H{
			"id":          v.Id,
			"orderId":     v.OrderId,
			"productName": v.ProductId,
			"quantity":    v.Quantity,
			"price":       v.SubTotal,
			"status":      v.OrderStatus,
		})
	}

	c.JSON(200, gin.H{
		"status": "success",
		"code":   "200",
		"orders": orderShow,
	})
}

// AdminCancelOrder function to cancel an order by ID
func AdminCancelOrder(c *gin.Context) {
	id := c.Param("ID")
	orderId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Fail",
			"error":  "Invalid order ID",
			"code":   http.StatusBadRequest,
		})
		return
	}

	var orderItem models.OrderItems
	if err := initializer.DB.Where("id = ?", orderId).Preload("Order").First(&orderItem).Error; err != nil {
		log.Println("Error fetching order item:", err)
		c.JSON(http.StatusNotFound, gin.H{
			"status":     "Fail",
			"error":      "Can't find the order",
			"error_code": http.StatusNotFound,
		})
		return
	}

	// Check if the order status is already cancelled
	if orderItem.OrderStatus == "cancelled" {
		c.JSON(http.StatusAccepted, gin.H{
			"status":  "Warning",
			"message": "This order has already been cancelled, please check your order",
			"code":    http.StatusAccepted,
		})
		return
	}

	// Update order status to cancelled
	orderItem.OrderStatus = "cancelled"
	if err := initializer.DB.Model(&orderItem).Updates(orderItem).Error; err != nil {
		log.Println("Error saving order item:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Fail",
			"error":  "Failed to save changes to the database",
			"code":   http.StatusInternalServerError,
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"status":  "Success",
		"message": "Order is cancelled",
		"data":    orderItem.OrderStatus,
	})
}


func AdminOrderStatus(c *gin.Context) {
	id := c.Param("ID")
	var orderStatus models.OrderItems
	orderStatusChenge := c.Request.FormValue("status")
	if orderStatusChenge == "" {
		c.JSON(406, gin.H{
			"status": "Fail",
			"error":  "Enter the Status",
			"code":   406,
		})
		return
	}
	if err := initializer.DB.First(&orderStatus, id).Error; err != nil {
		c.JSON(404, gin.H{
			"status": "Fail",
			"error":  "can't find order",
			"code":   404,
		})
		return
	}
	orderStatus.OrderStatus = orderStatusChenge
	initializer.DB.Save(&orderStatus)
	c.JSON(200, gin.H{
		"status":  "success",
		"message": "order status changed to  " + orderStatusChenge,
		"data":    orderStatus.OrderStatus,
	})

}