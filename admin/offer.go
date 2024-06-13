package admin

import (
	"net/http"
	"project/initializer"
	"project/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// OfferList godoc
// @Summary Get a list of offers
// @Description Retrieve a list of all available offers
// @Tags Admin/Offer
// @ID getOfferList
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {json} JSON "OK"
// @Failure 400 {string} string error message
// @Router /admin/offer [get]
func OfferShow(c *gin.Context) {
	var offer models.Offer
	offerID := c.Param("ID")

	if err := initializer.DB.First(&offer, offerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"status": "Fail",
				"error":  "offer not found",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Fail",
			"error":  "failed to retrieve offer",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "Success",
		"offer":  offer,
	})
}

// OfferAdd godoc
// @Summary Add a new offer
// @Description Add a new offer to the system
// @Tags Admin/Offer
// @ID addOffer
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param offer body models.Offer true "Offer details"
// @Success 200 {json}  JSON "New Offer Created"
// @Failure 400 {json}  ErrorResponse "Failed to create offer"
// @Router /admin/offer [post]
func OfferAdd(c *gin.Context) {
	var addOffer models.Offer

	// Bind JSON request body to addOffer struct
	if err := c.BindJSON(&addOffer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Fail",
			"error":  "failed to bind data",
		})
		return
	}

	// Check if ProductID and CategoryID are provided
	if addOffer.ProductID == 0 || addOffer.CategoryID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Fail",
			"error":  "product_id and category_id are required fields",
		})
		return
	}

	// Validate if the associated product exists
	var product models.Products
	if err := initializer.DB.First(&product, addOffer.ProductID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Fail",
				"error":  "product_id does not exist",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Fail",
			"error":  "error checking product existence",
		})
		return
	}

	// Validate if the associated category exists
	var category models.Category
	if err := initializer.DB.First(&category, addOffer.CategoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Fail",
				"error":  "category_id does not exist",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Fail",
			"error":  "error checking category existence",
		})
		return
	}

	// Create the offer in the database
	if err := initializer.DB.Create(&addOffer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Fail",
			"error":  "failed to create offer",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "New offer created",
	})
}

// OfferDelete godoc
// @Summary Delete an offer by ID
// @Description Delete an offer from the system by its unique identifier
// @Tags Admin/Offer
// @ID deleteOffer
// @Produce json
// @Security ApiKeyAuth
// @Param ID path int true "Offer ID"
// @Success 200 {json}  string  "Deleted Successfully"
// @Failure 400 {json}     ErrorResponse "Failed to delete offer"
// @Router /admin/offer/{ID} [delete]
func OfferDelete(c *gin.Context) {
	var offer models.Offer
	offerID := c.Param("ID")

	if err := initializer.DB.Where("id = ?", offerID).Delete(&offer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"status": "Fail",
				"error":  "offer not found",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Fail",
			"error":  "failed to delete offer",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Offer deleted successfully",
	})
}
