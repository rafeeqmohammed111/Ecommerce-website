package user

import (
	"project/initializer"
	"project/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func FetchCanceledOrdersAndUpdateWallet(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("user_id").(uint)
	if !ok {
		c.JSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	var canceledOrders []models.Order
	if err := initializer.DB.Where("user_id = ? AND order_status = ?", userID, "canceled").Find(&canceledOrders).Error; err != nil {
		c.JSON(400, gin.H{
			"status": "Fail",
			"error":  "Failed to fetch canceled orders",
			"code":   400,
		})
		return
	}

	var totalRefundAmount float64
	for _, order := range canceledOrders {
		totalRefundAmount += order.TotalAmount
	}

	var wallet models.Wallet
	if err := initializer.DB.First(&wallet, "user_id = ?", userID).Error; err != nil {
		// If the wallet is not found, create a new one
		if err == gorm.ErrRecordNotFound {
			wallet = models.Wallet{
				User_id: int(userID), // Correctly assign the userID here
				Balance: 0,
			}
			if err := initializer.DB.Create(&wallet).Error; err != nil {
				c.JSON(500, gin.H{
					"status": "Fail",
					"error":  "Failed to create wallet",
					"code":   500,
				})
				return
			}
		} else {
			c.JSON(500, gin.H{
				"status": "Fail",
				"error":  "Failed to fetch wallet",
				"code":   500,
			})
			return
		}
	}

	wallet.Balance += totalRefundAmount
	if err := initializer.DB.Save(&wallet).Error; err != nil {
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to update wallet balance",
			"code":   500,
		})
		return
	}

	c.JSON(200, gin.H{
		"status":            "Success",
		"message":           "Fetched canceled orders and updated wallet balance",
		"totalRefundAmount": totalRefundAmount,
		"walletBalance":     wallet.Balance,
		"canceledOrders":    canceledOrders,
	})
}
