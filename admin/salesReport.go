package admin

import (
	"fmt"
	"project/initializer"
	"project/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/tealeg/xlsx"
)

func filterOrdersByTimeRange(timeRange string) ([]models.Order, error) {
	var orders []models.Order
	now := time.Now()
	var startTime time.Time

	switch timeRange {
	case "daily":
		startTime = now.AddDate(0, 0, -1)
	case "weekly":
		startTime = now.AddDate(0, 0, -7)
	case "yearly":
		startTime = now.AddDate(-1, 0, 0)
	default:
		startTime = now
	}

	if err := initializer.DB.Where("created_at >= ?", startTime).Preload("Coupon").Preload("OrderItems.Product").Find(&orders).Error; err != nil {
		return nil, err
	}

	return orders, nil
}

type ReportData struct {
	TimeRange        string  `json:"time_range"`
	TotalOrders      int     `json:"total_orders"`
	TotalSalesAmount float64 `json:"total_sales_amount"`
	TotalDiscount    float64 `json:"total_discount"`
}

// GenerateReport generates a sales report based on the specified time range(day ,week,year).
func GenerateReport(c *gin.Context) {
	timeRange := c.Query("time_range")
	if timeRange == "" {
		timeRange = "daily"
	}

	// Fetch orders based on time range
	orders, err := filterOrdersByTimeRange(timeRange)
	if err != nil {
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to fetch orders",
			"code":   500,
		})
		return
	}

	var totalOrders int
	var totalSalesAmount float64
	var totalDiscount float64

	for _, order := range orders {
		totalOrders++
		totalSalesAmount += order.OrderAmount
		if order.CouponID != nil {
			totalDiscount += order.Coupon.Discount
		}
	}

	report := ReportData{
		TimeRange:        timeRange,
		TotalOrders:      totalOrders,
		TotalSalesAmount: totalSalesAmount,
		TotalDiscount:    totalDiscount,
	}

	// Respond with JSON data
	c.JSON(200, report)
}

func SalesReportExcel(c *gin.Context) {
	timeRange := c.Query("time_range")
	if timeRange == "" {
		timeRange = "daily"
	}

	orders, err := filterOrdersByTimeRange(timeRange)
	if err != nil {
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to fetch sales data",
			"code":   500,
		})
		return
	}

	// Create new Excel file
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Sales Report")
	if err != nil {
		c.JSON(400, gin.H{
			"status": "Fail",
			"error":  "Failed to create Excel sheet",
			"code":   400,
		})
		return
	}

	headers := []string{"Order ID", "Product Name", "Order Date", "Total Amount", "Coupon Discount"}
	row := sheet.AddRow()
	for _, header := range headers {
		cell := row.AddCell()
		cell.Value = header
	}

	// Add sales data
	var totalAmount float64
	var totalDiscount float64
	for _, order := range orders {
		for _, item := range order.OrderItems {
			row := sheet.AddRow()
			row.AddCell().Value = order.Id // Keep order.Id as string
			row.AddCell().Value = item.Product.Name
			row.AddCell().Value = order.OrderDate.Format("2006-01-02")
			row.AddCell().Value = fmt.Sprintf("%.2f", order.OrderAmount)
			discount := 0.0
			if order.CouponID != nil {
				discount = order.Coupon.Discount
			}
			row.AddCell().Value = fmt.Sprintf("%.2f", discount)
			totalAmount += order.OrderAmount
			totalDiscount += discount
		}
	}

	totalRow := sheet.AddRow()
	totalRow.AddCell()
	totalRow.AddCell()
	totalRow.AddCell().Value = "Total Amount:"
	totalRow.AddCell().Value = fmt.Sprintf("%.2f", totalAmount)
	totalRow.AddCell().Value = fmt.Sprintf("%.2f", totalDiscount)

	// Save Excel file
	excelPath := "/home/abuaibak/Desktop/sales_report.xlsx"
	if err := file.Save(excelPath); err != nil {
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to save Excel file",
			"code":   500,
		})
		return
	}

	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", "sales_report.xlsx"))
	c.Writer.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.File(excelPath)
}

func SalesReportPDF(c *gin.Context) {
	timeRange := c.Query("time_range")
	if timeRange == "" {
		timeRange = "daily"
	}

	orders, err := filterOrdersByTimeRange(timeRange)
	if err != nil {
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to fetch sales data",
			"code":   500,
		})
		return
	}

	// Create new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)

	headers := []string{"Order ID", "Product", "Order Date", "Total Amount", "Coupon Discount"}
	for _, header := range headers {
		pdf.Cell(50, 10, header)
	}
	pdf.Ln(-1)

	// Add sales data
	var totalAmount float64
	var totalDiscount float64
	for _, order := range orders {
		for _, item := range order.OrderItems {
			pdf.Cell(50, 10, order.Id) // Keep order.Id as string
			pdf.Cell(50, 10, item.Product.Name)
			pdf.Cell(50, 10, order.OrderDate.Format("2006-01-02"))
			pdf.Cell(50, 10, fmt.Sprintf("%.2f", order.OrderAmount))
			discount := 0.0
			if order.CouponID != nil {
				discount = order.Coupon.Discount
			}
			pdf.Cell(50, 10, fmt.Sprintf("%.2f", discount))
			pdf.Ln(-1)
			totalAmount += order.OrderAmount
			totalDiscount += discount
		}
	}

	// Save
	pdfPath := "/home/abuaibak/Desktop/sales_report.pdf"
	if err := pdf.OutputFileAndClose(pdfPath); err != nil {
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to generate PDF file",
			"code":   500,
		})
		return
	}

	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", "sales_report.pdf"))
	c.Writer.Header().Set("Content-Type", "application/pdf")
	c.File(pdfPath)
}
