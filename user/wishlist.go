package user

import (
	"project/initializer"
	"project/models"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// authentication need to add**************************
// wishlist items show
// @Summary Wishlist show
// @Description Added wishlist product list shown
// @Tags Wishlist
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {json} SuccessResponse
// @Failure 401 {json} ErrorResponse
// @Router /wishlist [get]
func WishlistProducts(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("user_id").(uint)
	if !ok {
		c.JSON(401, gin.H{"message": "Unauthorized"})
		return
	}
	var wishlist []models.Wishlist
	var wishlistShow []gin.H

	if err := initializer.DB.Where("user_id=?", userID).Preload("Product").Find(&wishlist).Error; err != nil {
		c.JSON(400, gin.H{
			"message": "Fail",
			"error ":  "failed to fetch wishlist items",
		})
		return
	}
	if len(wishlist) == 0 {
		c.JSON(400, gin.H{
			"message":  "Fail",
			"message ": "No item found in wishlist",
		})
		return
	}
	for _, v := range wishlist {
		wishlistShow = append(wishlistShow, gin.H{
			"ProductName":  v.Product.Name,
			"productPrice": v.Product.Price,
			"productSize":  v.Product.Size,
		})
	}
	c.JSON(200, gin.H{
		"status": "success",
		"data":   wishlistShow,
	})
}

// Add items to wishlist for future references
// @Summary Wishlist add product
// @Description Add product that likes to wishlist
// @Tags Wishlist
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "product id"
// @Success 200 {json} SuccessResponse
// @Failure 401 {json} ErrorResponse
// @Router /wishlist [post]
func WishlistAdd(c *gin.Context) {
	var wishAdd models.Wishlist
	session := sessions.Default(c)
    userID, ok := session.Get("user_id").(uint)
    if !ok {
        c.JSON(401, gin.H{"message": "Unauthorized"})
        return
    }
	id := c.Param("ID")
	err := initializer.DB.Where("user_id=? AND product_id=?", userID, id).First(&wishAdd)
	if err.Error != nil {
		wishAdd.UserId = int(userID)
		wishAdd.ProductId, _ = strconv.Atoi(id)
		if err := initializer.DB.Create(&wishAdd).Error; err != nil {
			c.JSON(400, gin.H{
				"status": "fail",
				"error":  "Failed to add to wishlist",
				"code":   400,
			})
			return
		}
		c.JSON(200, gin.H{
			"status":  "Success",
			"message": "Item added to wishlist",
		})
	} else {
		c.JSON(409, gin.H{
			"status": "Fail",
			"error":  "This item already added",
			"code":   409,
		})
	}
}

// Delete existing product from wishlist
// @Summary Wishlist remove product
// @Description remove product that from the wishlist
// @Tags Wishlist
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "product id"
// @Success 200 {json} SuccessResponse
// @Failure 401 {json} ErrorResponse
// @Router /wishlist [delete]
func WishlistDelete(c *gin.Context) {
	var wishlistDelete models.Wishlist
	session := sessions.Default(c)
    userID, ok := session.Get("user_id").(uint)
    if !ok {
        c.JSON(401, gin.H{"message": "Unauthorized"})
        return
    }
	id := c.Param("ID")
	if err := initializer.DB.Where("product_id=? AND user_id=?", id, userID).Delete(&wishlistDelete).Error; err != nil {
		c.JSON(501, gin.H{
			"status": "Fail",
			"error":  "failed to remove Item",
			"code":   501,
		})
		return
	}
	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Item remove successfully",
	})
}
