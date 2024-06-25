package user

import (
    "bytes"
    "fmt"
    "net/http"
    "project/initializer"
    "project/models"
    "strconv"

    "github.com/gin-contrib/sessions"
    "github.com/gin-gonic/gin"
    "github.com/jung-kurt/gofpdf"
)

func CreateInvoice(c *gin.Context) {
    session := sessions.Default(c)
    userID, ok := session.Get("user_id").(uint)
    if !ok {
        c.JSON(401, gin.H{"message": "Unauthorized"})
        return
    }
    orderId := c.Param("id")
    var user models.Users
    if err := initializer.DB.First(&user, userID).Error; err != nil {
        c.JSON(404, gin.H{
            "status": "Fail",
            "error":  "User not found",
            "code":   404,
        })
        return
    }

    var orderItem []models.OrderItems
    if err := initializer.DB.Preload("Product").Where("order_id = ?", orderId).Find(&orderItem).Error; err != nil {
        c.JSON(503, gin.H{
            "status": "Fail",
            "error":  "Failed to fetch orders",
            "code":   503,
        })
        return
    }

    var orders models.Order
    if err := initializer.DB.Preload("Address").Where("id = ?", orderId).Find(&orders).Error; err != nil {
        c.JSON(503, gin.H{
            "status": "Fail",
            "error":  "Failed to fetch orders",
            "code":   503,
        })
        return
    }

    var order models.Order
    var Discount float64
    initializer.DB.Find(&order, orderId)

    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()
    pdf.SetFont("Arial", "B", 20)
    pdf.Ln(5)
    pdf.CellFormat(0, 0, "INVOICE", "", 0, "C", false, 0, "")
    pdf.SetFont("Arial", "", 12)
    pdf.Ln(30)
    pdf.Cell(10, -32, "Invoice No: "+orderId)
    pdf.Ln(5)
    pdf.Cell(10, -32, "Invoice Date: "+order.OrderDate.Format("2006-01-02"))
    pdf.Ln(15)
    pdf.SetFont("Arial", "", 12)
    pdf.Cell(10, -32, "Bill To: ")
    pdf.Ln(5)
    pdf.Cell(10, -32, "Customer Name: "+user.Name)
    pdf.SetFont("Arial", "", 12)
    pdf.Ln(5)
    for _, val := range orderItem {
        pdf.Cell(10, -32, "Address: "+val.Order.Address.City+", "+val.Order.Address.State)
        pdf.Ln(5)
        pdf.Cell(10, -32, strconv.Itoa(val.Order.Address.Pincode))
        pdf.Ln(5)
        pdf.Cell(10, -32, "Phone no : "+strconv.Itoa(user.Phone))
        pdf.Ln(5)
        pdf.SetFont("Arial", "", 12)
        pdf.Ln(10)
        break
    }

    pdf.SetXY(10, 20)
    pdf.CellFormat(170, 30, "Hilofy", "", 0, "R", false, 0, "")
    pdf.SetFont("Arial", "", 12)
    pdf.CellFormat(12, 40, "dilka , rashka del", "", 0, "R", false, 0, "")
    pdf.CellFormat(12, 50, "15th floor ,Ph: +324 36545", "", 0, "R", false, 0, "")
    pdf.Ln(60)

    pdf.SetFillColor(220, 220, 220)
    pdf.CellFormat(20, 10, "No.", "1", 0, "C", true, 0, "")
    pdf.CellFormat(70, 10, "Item Name", "1", 0, "C", true, 0, "")
    pdf.CellFormat(30, 10, "Quantity", "1", 0, "C", true, 0, "")
    pdf.CellFormat(30, 10, "Product Price", "1", 0, "C", true, 0, "")
    pdf.CellFormat(40, 10, "Total Price", "1", 0, "C", true, 0, "")
    pdf.Ln(10)

    totalAmount := 0.0
    for i, order := range orderItem {
        pdf.CellFormat(20, 10, fmt.Sprintf("%d", i+1), "1", 0, "C", false, 0, "")
        pdf.CellFormat(70, 10, order.Product.Name, "1", 0, "", false, 0, "")
        pdf.CellFormat(30, 10, fmt.Sprintf("%d", order.Quantity), "1", 0, "C", false, 0, "")
        pdf.CellFormat(30, 10, fmt.Sprintf("%.2f", float64(order.Product.Price)), "1", 0, "R", false, 0, "")
        pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", order.SubTotal), "1", 0, "R", false, 0, "")
        pdf.Ln(10)
        totalAmount += float64(order.SubTotal)
    }

    if order.ShippingCharge > 0 {
        order.OrderAmount -= float64(order.ShippingCharge)
    }

   
    if order.Coupon != nil {
        Discount = order.Coupon.Discount
    }

    totalAmount -= Discount
    if Discount > 0 {
        pdf.CellFormat(150, 10, "Discount:", "1", 0, "R", true, 0, "")
        pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", Discount), "1", 0, "R", true, 0, "")
        pdf.Ln(10)
    }
    if order.ShippingCharge > 0 {
        totalAmount += float64(order.ShippingCharge)
        pdf.CellFormat(150, 10, "Shipping charge:", "1", 0, "R", true, 0, "")
        pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", order.ShippingCharge), "1", 0, "R", true, 0, "")
        pdf.Ln(10)
    }
    Discount = 0
    pdf.CellFormat(150, 10, "Total Amount: ", "1", 0, "R", true, 0, "")
    pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", totalAmount), "1", 0, "R", true, 0, "")

  
    var buf bytes.Buffer
    if err := pdf.Output(&buf); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate pdf", "details": err.Error()})
        return
    }

    c.Header("Content-Type", "application/pdf")
    c.Header("Content-Disposition", "attachment; filename=invoice.pdf")
    c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}
