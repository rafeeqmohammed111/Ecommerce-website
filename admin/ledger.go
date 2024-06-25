package admin

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"project/initializer"
	"project/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

func printSection(pdf *gofpdf.Fpdf, label string, value string) {
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(47.5, 10, label, "1", 0, "L", false, 0, "")
	pdf.CellFormat(47.5, 10, value, "1", 0, "R", false, 0, "")
	pdf.Ln(-1)
}

func printMapSection(pdf *gofpdf.Fpdf, label string, data map[string]interface{}) {
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(47.5, 10, label, "1", 1, "L", false, 0, "")
	for key, value := range data {
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(47.5, 10, key, "1", 0, "L", false, 0, "")
		pdf.CellFormat(47.5, 10, fmt.Sprintf("%v", value), "1", 1, "R", false, 0, "")
	}
	pdf.Ln(-1)
}

func printSingleLineSection(pdf *gofpdf.Fpdf, label string, value string) {
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(95, 10, label, "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 10, value, "1", 1, "R", false, 0, "")
	pdf.Ln(-1)
}

func GetMonthlySalesSummary(c *gin.Context) {
	var onlineSales, codSales, canceledAmount, couponTotal, totalSales, profit float64
	var totalProductsSold uint
	productSales := make(map[string]uint)
	productBalance := make(map[string]float64)
	categorySales := make(map[string]uint)
	productStock := make(map[string]uint)
	couponDetails := make(map[string]float64)

	startOfMonth := time.Now().UTC().AddDate(0, 0, -time.Now().Day()+1).Truncate(24 * time.Hour)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	// orders created within this month
	var orders []models.Order
	if err := initializer.DB.Where("created_at BETWEEN ? AND ?", startOfMonth, endOfMonth).Find(&orders).Error; err != nil {
		log.Println("Error fetching orders:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Fail",
			"error":  "Failed to fetch orders",
			"code":   http.StatusInternalServerError,
		})
		return
	}

	// Process each order
	for _, order := range orders {
		totalSales += order.OrderAmount

		//  sales by payment method
		switch order.OrderPaymentMethod {
		case "ONLINE":
			onlineSales += order.OrderAmount
		case "COD":
			codSales += order.OrderAmount
		}

		//  order items for each order
		var orderItems []models.OrderItems
		if err := initializer.DB.Where("order_id = ?", order.Id).Find(&orderItems).Error; err != nil {
			log.Println("Error fetching order items:", err)
			continue
		}

		// taking each order item
		for _, item := range orderItems {
			totalProductsSold += item.Quantity

			// counting canceled amount
			if item.OrderStatus == "cancelled" {
				canceledAmount += item.SubTotal
			}

			//  sale and blnc
			productSales[item.ProductName] += item.Quantity
			productBalance[item.ProductName] += item.SubTotal

			//  ctgry sale
			categorySales[item.Category] += item.Quantity

			// prdct stk
			var product models.Products
			if err := initializer.DB.First(&product, "name = ?", item.ProductName).Error; err == nil {
				productStock[item.ProductName] = product.Stock
			}

			//  coupon details
			if order.CouponCode != "" {
				var coupon models.Coupon
				if err := initializer.DB.First(&coupon, "code = ?", order.CouponCode).Error; err == nil {
					couponDetails[coupon.Code] += float64(coupon.Discount)
					couponTotal += float64(coupon.Discount)
				}
			}
		}
	}

	//  profit (10% of total sales amount)
	profit = totalSales * 0.10

	// Generate PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(190, 10, "LEDGER REPORT", "", 1, "C", false, 0, "")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(95, 10, "Sales Summary", "1", 0, "C", false, 0, "")
	pdf.CellFormat(95, 10, "Product Details", "1", 1, "C", false, 0, "")

	// Print Sales Summary
	printSection(pdf, "Online Sales", fmt.Sprintf("%.2f", onlineSales))
	printSection(pdf, "COD Sales", fmt.Sprintf("%.2f", codSales))
	printSection(pdf, "Canceled Amount", fmt.Sprintf("%.2f", canceledAmount))
	printSection(pdf, "Coupon Total", fmt.Sprintf("%.2f", couponTotal))
	printSection(pdf, "Total Sales Amount", fmt.Sprintf("%.2f", totalSales))
	printSection(pdf, "Profit (10% of Sales)", fmt.Sprintf("%.2f", profit))

	// Print Product Details
	printMapSection(pdf, "Product Sales", convertIntMapToStringMap(productSales))
	printMapSection(pdf, "Product Balance", convertFloat64MapToStringMap(productBalance))
	printMapSection(pdf, "Category Sales", convertIntMapToStringMap(categorySales))
	printMapSection(pdf, "Product Stock", convertIntMapToStringMap(productStock))
	printMapSection(pdf, "Coupon Details", convertFloat64MapToStringMap(couponDetails))
	printSingleLineSection(pdf, "Current Stock of Products", getCurrentStock(productStock))

	fileName := "LEDGER REPORT.pdf"
	err := pdf.OutputFileAndClose(fileName)
	if err != nil {
		log.Println("Error generating PDF:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Fail",
			"error":  "Failed to generate PDF",
			"code":   http.StatusInternalServerError,
		})
		return
	}

	// Send the PDF file
	c.FileAttachment(fileName, fileName)

	// Delete the file after making
	defer os.Remove(fileName)
}

// convert uint to interface{}
func convertIntMapToStringMap(input map[string]uint) map[string]interface{} {
	output := make(map[string]interface{})
	for key, value := range input {
		output[key] = value
	}
	return output
}

// convert float to interface{}
func convertFloat64MapToStringMap(input map[string]float64) map[string]interface{} {
	output := make(map[string]interface{})
	for key, value := range input {
		output[key] = value
	}
	return output
}

// current stock of products
func getCurrentStock(productStock map[string]uint) string {
	stock := ""
	for product, quantity := range productStock {
		stock += fmt.Sprintf("%s: %d\n", product, quantity)
	}
	return stock
}
