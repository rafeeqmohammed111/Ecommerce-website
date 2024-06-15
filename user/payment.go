package user

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"project/initializer"
	"project/models"

	"github.com/gin-gonic/gin"
	"github.com/razorpay/razorpay-go"
)

// PaymentHandler initiates payment with Razorpay
func PaymentHandler(orderID string, amount float64) (string, error) {
	client := razorpay.NewClient(os.Getenv("RAZORPAY_KEY"), os.Getenv("RAZORPAY_SECRET"))
	orderParams := map[string]interface{}{
		"amount":   int(amount * 100), // Razorpay expects amount in paise
		"currency": "INR",
		"receipt":  orderID,
	}
	order, err := client.Order.Create(orderParams, nil)
	if err != nil {
		return "", errors.New("PAYMENT NOT INITIATED: " + err.Error())
	}

	razorID, ok := order["id"].(string)
	if !ok {
		return "", errors.New("PAYMENT NOT INITIATED: invalid Razorpay order ID")
	}

	return razorID, nil
}

// PaymentConfirmation handles Razorpay payment confirmation
func PaymentConfirmation(c *gin.Context) {
	var paymentStore models.PaymentDetails
	var paymentDetails = make(map[string]string)
	if err := c.BindJSON(&paymentDetails); err != nil {
		c.JSON(400, gin.H{
			"status": "fail",
			"error":  "Invalid request body",
			"code":   400,
		})
		return
	}

	// Verify Razorpay payment signature
	err := RazorPaymentVerification(paymentDetails["signature"], paymentDetails["order_id"], paymentDetails["payment_id"])
	if err != nil {
		c.JSON(400, gin.H{
			"status": "fail",
			"error":  "Payment verification failed",
			"code":   400,
		})
		return
	}

	// Fetch order details from the database based on order_id
	if err := initializer.DB.First(&paymentStore, "order_id=?", paymentDetails["order_id"]).Error; err != nil {
		c.JSON(404, gin.H{
			"status": "fail",
			"error":  "Order details not found",
			"code":   404,
		})
		return
	}

	// Update payment details in the database
	paymentStore.PaymentId = paymentDetails["payment_id"]
	paymentStore.PaymentStatus = "success"
	if err := initializer.DB.Save(&paymentStore).Error; err != nil {
		c.JSON(500, gin.H{
			"status": "fail",
			"error":  "Failed to update payment details",
			"code":   500,
		})
		return
	}

	// Update product quantities or any other related operations
	// Fetch order items and update quantities, etc.
	var orderItems []models.OrderItems
	if err := initializer.DB.Where("order_id = ?", paymentDetails["order_id"]).Find(&orderItems).Error; err != nil {
		c.JSON(500, gin.H{
			"status": "fail",
			"error":  "Failed to fetch order items",
			"code":   500,
		})
		return
	}

	// Example: Update product quantities
	for _, item := range orderItems {
		var product models.Products
		if err := initializer.DB.First(&product, item.ProductId).Error; err != nil {
			fmt.Printf("Failed to find product with ID %d\n", item.ProductId)
			continue
		}

		// Adjust product quantities or other operations
		product.Quantity -= int(item.Quantity)
		if err := initializer.DB.Save(&product).Error; err != nil {
			fmt.Printf("Failed to update product quantity for product ID %d\n", item.ProductId)
			continue
		}
	}

	// Success response
	c.JSON(200, gin.H{
		"status":     "success",
		"message":    "Payment confirmed successfully",
		"order_id":   paymentDetails["order_id"],
		"payment_id": paymentDetails["payment_id"],
	})
}

// RazorPaymentVerification verifies Razorpay payment signature
func RazorPaymentVerification(sign, orderId, paymentId string) error {
	signature := sign
	secret := os.Getenv("RAZORPAY_SECRET")
	data := orderId + "|" + paymentId
	h := hmac.New(sha256.New, []byte(secret))
	_, err := h.Write([]byte(data))
	if err != nil {
		panic(err)
	}
	sha := hex.EncodeToString(h.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(sha), []byte(signature)) != 1 {
		return errors.New("PAYMENT FAILED")
	} else {
		return nil
	}
}
