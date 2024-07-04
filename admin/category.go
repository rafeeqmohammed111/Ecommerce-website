package admin

import (
	"net/http"

	"project/initializer"
	"project/models"

	"github.com/gin-gonic/gin"
)

func CategoryList(c *gin.Context) {
	var categorylist []models.Category
	initializer.DB.Find(&categorylist)
	for _, v := range categorylist {
		c.JSON(200, gin.H{
			"Id":                   v.ID,
			"Category_name":        v.Category_name,
			"Category_description": v.Category_description,
			"Category_status":      v.Blocking,
			
		})
	}
}
func AddCategory(c *gin.Context) {
	var addcategory models.Category
	if err := c.ShouldBind(&addcategory); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bind data"})
		return
	}
	addcategory.Blocking = true
	if result := initializer.DB.Create(&addcategory); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert product"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "category created successfully"})
}
func EditCategories(c *gin.Context) {
	var editcategory models.Category
	id := c.Param("ID")
	err := initializer.DB.First(&editcategory, id)
	if err.Error != nil {
		c.JSON(500, gin.H{"error": "can't find category"})
	} else {
		err := c.ShouldBindJSON(&editcategory)
		if err != nil {
			c.JSON(500, "failed to bild details")
		} else {
			if err := initializer.DB.Save(&editcategory).Error; err != nil {
				c.JSON(500, "failed to edit details")
			}
			c.JSON(200, "successfully edited category")
		}
	}
}
func DeleteCategories(c *gin.Context) {
	var deletecategory models.Category
	id := c.Param("ID")
	err := initializer.DB.First(&deletecategory, id)
	if err.Error != nil {
		c.JSON(500, gin.H{"error": "can't find category"})
	} else {
		err := initializer.DB.Delete(&deletecategory).Error
		if err != nil {
			c.JSON(500, "failed to delete category")
		} else {
			c.JSON(200, "category deleted successfully")
		}
	}
}
func BlockCategory(c *gin.Context) {
	var blockCategory models.Category
	id := c.Param("ID")
	err := initializer.DB.First(&blockCategory, id)
	if err.Error != nil {
		c.JSON(500, gin.H{"error": "can't find Category"})
	} else {
		if blockCategory.Blocking {
			blockCategory.Blocking = false
			c.JSON(200, "Category blocked")
		} else {
			blockCategory.Blocking = true
			c.JSON(200, "Category unblocked")
		}
		if err := initializer.DB.Save(&blockCategory).Error; err != nil {
			c.JSON(500, "failed to block/unblock Category")
		}
	}
}

// everything don
