package user

import (
	"project/initializer"
	"project/models"

	"github.com/gin-gonic/gin"
)

// var products []models.Products

// @Summary		Landing page
// @Description	Get a list of products from the database
// @Tags			LandingPage
// @Accept			json
// @Produce		json
// @Success		200	{string}	OK
// @Router			/ [get]
func ProductsPage(c *gin.Context) {
	var products []models.Products
	err := initializer.DB.Order("products.name").Find(&products).Error
	if err != nil {
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to find products",
			"code":   500,
		})
		return
	}

	var productList []gin.H
	for _, product := range products {
		productList = append(productList, gin.H{
			"Id":    product.ID,
			"Name":  product.Name,
			"Price": product.Price,
		})
	}

	c.JSON(200, gin.H{
		"status": "Success",
		"data":   productList,
	})
}
func ProductDetails(c *gin.Context) {
	var product models.Products
	id := c.Param("id")

	// Fetch product details
	if err := initializer.DB.First(&product, id).Error; err != nil {
		c.JSON(404, gin.H{
			"status": "Fail",
			"error":  "Can't find product",
			"code":   404,
		})
		return
	}

	var quantityStatus string
	if product.Quantity == 0 {
		quantityStatus = "Out of stock"
	} else {
		quantityStatus = "Product available"
	}

	// Fetch related products
	var relatedProducts []models.Products
	if err := initializer.DB.Where("category_id = ? AND id != ?", product.CategoryId, id).Find(&relatedProducts).Error; err != nil {
		c.JSON(500, gin.H{
			"status": "Fail",
			"error":  "Failed to find related products",
			"code":   500,
		})
		return
	}

	var relatedProductsList []gin.H
	for _, relatedProduct := range relatedProducts {
		relatedProductsList = append(relatedProductsList, gin.H{
			"ProductName":  relatedProduct.Name,
			"ProductPrice": relatedProduct.Price,
			"ProductSize":  relatedProduct.Size,
		})
	}

	productDetails := gin.H{
		"name":            product.Name,
		"price":           product.Price,
		"description":     product.Description,
		"size":            product.Size,
		"color":           product.Color,
		"imageURL":        product.ImagePath,
		"categoryId":      product.CategoryId,
		"stockStatus":     quantityStatus,
		"relatedProducts": relatedProductsList,
	}

	c.JSON(200, gin.H{
		"status": "Success",
		"data":   productDetails,
	})
}
