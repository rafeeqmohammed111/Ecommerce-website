// package user

// import (
// 	"project/initializer"
// 	"project/models"

// 	"github.com/gin-contrib/sessions"
// 	"github.com/gin-gonic/gin"
// 	"gorm.io/gorm"
// )

// func FetchCanceledOrdersAndUpdateWallet(c *gin.Context) {
// 	session := sessions.Default(c)
// 	userID, ok := session.Get("user_id").(uint)
// 	if !ok {
// 		c.JSON(401, gin.H{"message": "Unauthorized"})
// 		return
// 	}

// 	var canceledOrders []models.Order
// 	if err := initializer.DB.Where("user_id = ? AND order_status = ?", userID, "canceled").Find(&canceledOrders).Error; err != nil {
// 		c.JSON(400, gin.H{
// 			"status": "Fail",
// 			"error":  "Failed to fetch canceled orders",
// 			"code":   400,
// 		})
// 		return
// 	}

// 	var totalRefundAmount float64
// 	for _, order := range canceledOrders {
// 		totalRefundAmount += order.TotalAmount
// 	}

// 	var wallet models.Wallet
// 	if err := initializer.DB.First(&wallet, "user_id = ?", userID).Error; err != nil {
// 		// If the wallet is not found, creating new one
// 		if err == gorm.ErrRecordNotFound {
// 			wallet = models.Wallet{
// 				UserId:  int(userID),
// 				Balance: 0,
// 			}
// 			if err := initializer.DB.Create(&wallet).Error; err != nil {
// 				c.JSON(500, gin.H{
// 					"status": "Fail",
// 					"error":  "Failed to create wallet",
// 					"code":   500,
// 				})
// 				return
// 			}
// 		} else {
// 			c.JSON(500, gin.H{
// 				"status": "Fail",
// 				"error":  "Failed to fetch wallet",
// 				"code":   500,
// 			})
// 			return
// 		}
// 	}

// 	wallet.Balance += totalRefundAmount
// 	if err := initializer.DB.Save(&wallet).Error; err != nil {
// 		c.JSON(500, gin.H{
// 			"status": "Fail",
// 			"error":  "Failed to update wallet balance",
// 			"code":   500,
// 		})
// 		return
// 	}

//		c.JSON(200, gin.H{
//			"status":            "Success",
//			"message":           " updated wallet balance",
//			"totalRefundAmount": totalRefundAmount,
//			"walletBalance":     wallet.Balance,
//			"canceledOrders":    canceledOrders,
//		})
//	}
package user

import (
	"net/http"
	"project/initializer"
	"project/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func FetchCanceledOrdersAndUpdateWallet(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("user_id").(uint)
	if !ok {
		c.JSON(401, gin.H{"message": "Unauthorized"})
		return
	}
	var wallet models.Wallet
	if err := initializer.DB.Where("user_id=?", userID).Find(&wallet).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "faild to find the user"})
	}
	c.JSON(http.StatusOK, gin.H{
		"yuour wallet Balance": wallet.Balance,
		"credited amount":      wallet.CreditedAmount,
	})
}
