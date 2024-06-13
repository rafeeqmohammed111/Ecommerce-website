package user

import (
	"net/http"
	"project/initializer"
	"project/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Function to search category by ID it will show the detaild result
func SearchCategoryByID(c *gin.Context) {
	var category models.Category
	var products []models.Products

	categoryID := c.Param("id")

	
	if err := initializer.DB.First(&category, categoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"status": "Fail", "message": "Category not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Fail", "message": "Error fetching category"})
		return
	}

	// Find products associated with the category
	if err := initializer.DB.Where("category_id = ?", categoryID).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Fail", "message": "Error fetching products"})
		return
	}


	simplifiedProducts := make([]models.SimplifiedProduct, len(products))
	for i, product := range products {
		simplifiedProducts[i] = models.SimplifiedProduct{
			ID:       product.ID,
			Name:     product.Name,
			Price:    product.Price,
			Quantity: product.Quantity,
		}
	}

	// Return the category details along with simplified products
	c.JSON(http.StatusOK, gin.H{
		"status":              "Success",
		"category_details":    category,
		"associated_products": simplifiedProducts,
	})
}
