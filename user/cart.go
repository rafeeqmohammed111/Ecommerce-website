package user

import (
	"fmt"
	"net/http"
	"project/initializer"
	"project/models"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// CartView retrieves and displays the items in the user's cart along with total amount .
// @Summary View cart items
// @Description Retrieves and displays the items in the user's cart along with total amount.

func CartView(c *gin.Context) {
	var cartView []models.Cart
	var cartShow []gin.H
	session := sessions.Default(c)
	userid := session.Get("user_id")
	// userid := c.GetUint("userid")
	var totalAmount float64
	var count float64

	// Fetch cart items with associated products
	if err := initializer.DB.Where("user_id = ?", userid).Joins("Product").Find(&cartView).Error; err != nil {
		c.JSON(400, gin.H{
			"status": "Fail",
			"error":  "failed to fetch data",
			"code":   400,
		})
		return
	}

	// Process each cart item
	for _, v := range cartView {
		price := float64(v.Quantity) * float64(v.Product.Price)
		totalAmount += price
		count += 1
		cartShow = append(cartShow, gin.H{
			"product": gin.H{
				"id":          v.Product.ID,
				"name":        v.Product.Name,
				"price":       v.Product.Price,
				"size":        v.Product.Size,
				"color":       v.Product.Color,
				"description": v.Product.Description,
				"image_path":  v.Product.ImagePath,
			},
			"quantity": v.Quantity,
		})
	}

	// Handle empty cart case
	if count == 0 {
		c.JSON(200, gin.H{
			"status":  "Success",
			"message": "No product in your cart.",
			"data":    nil,
			"total":   0,
		})
		return
	}

	// Return cart data
	c.JSON(200, gin.H{
		"data":          cartShow,
		"totalProducts": count,
		"totalAmount":   totalAmount,
		"status":        "Success",
	})
}

// CartStore adds a product to the user's cart if it's not already added.
// @Summary Add product to cart
// @Description Adds a product to the user's cart if it's not already added.
// @Tags Cart
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param ID path string true "Product ID"
// @Success 200 {json} JSON Response "Item was successfully added."
// @Failure 400 {json} JSON ErrorResponse  "Invalid input data."
// @Router /cart/{ID} [post]
func CartStore(c *gin.Context) {
	session := sessions.Default(c)
	userid := session.Get("user_id")
	fmt.Println("=================", userid)

	userId, ok := userid.(uint)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Fail",
			"error":  "invalid user ID",
			"code":   400,
		})
		return
	}
	id := c.Param("ID")

	// Query the database to check if the product is already in the cart
	var cart models.Cart
	err := initializer.DB.Where("user_id=? AND product_id=?", userId, id).First(&cart).Error
	if err != nil {
		// If the product is not already in the cart, attempt to add it
		cart.UserId = userId
		cart.ProductId, _ = strconv.Atoi(id)
		cart.Quantity = 1
		if err := initializer.DB.Create(&cart).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Fail",
				"error":  "failed to add to cart",
				"code":   400,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status":  "Success",
				"message": "product added to cart",
			})
		}
	} else {
		// If the product is already in the cart, respond accordingly
		c.JSON(http.StatusConflict, gin.H{
			"status": "Exist",
			"error":  "product already added",
			"code":   409,
		})
	}
}

// CartProductAdd increases the quantity of a product in the user's cart if it's available and within the quantity limit.
// @Summary Increase quantity of product in cart
// @Description Increases the quantity of a product in the user's cart if it's available and within the quantity limit.
// @Tags Cart
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param ID path string true "Product ID"
// @Success 200 {string} string "one more quantity added""
// @Failure 201 {string} string "can't add more quantity"
// @Failure 202 {string} string "product out of stock"
// @Failure 400 {string} string "failed to add to one more"
// @Failure 404 {string} string "failed to fetch product stock details/can't find product"
// @Router /cart/{ID}/add [patch]
func CartProductAdd(c *gin.Context) {
	var cartStore models.Cart
	var productStock models.Products
	session := sessions.Default(c)
	userid := session.Get("user_id")
	fmt.Println("=================", userid)
	// userId := c.GetUint("userid")
	id := c.Param("ID")
	if err := initializer.DB.First(&productStock, id).Error; err != nil {
		c.JSON(404, gin.H{
			"status": "Fail",
			"error":  "failed to fetch product stock deatails",
			"code":   404,
		})
	}

	err := initializer.DB.Where("user_id=? AND product_id=?", userid, id).First(&cartStore).Error
	if err != nil {
		c.JSON(404, gin.H{
			"status": "Fail",
			"error":  "can't find product",
			"code":   404,
		})
	} else {
		cartStore.Quantity += 1
		if uint((productStock.Quantity)) >= cartStore.Quantity {
			if cartStore.Quantity <= 5 {
				err := initializer.DB.Where("user_id=? AND product_id=?", userid, cartStore.ProductId).Save(&cartStore)
				if err.Error != nil {
					c.JSON(400, gin.H{
						"status": "Fail",
						"error":  "failed to add to one more",
						"code":   400,
					})
				} else {
					c.JSON(200, gin.H{
						"status":   "Success",
						"message":  "one more quantity added",
						"quantity": cartStore.Quantity,
					})
				}
			} else {
				c.JSON(201, gin.H{
					"status":   "Fail",
					"error":    "can't add more quantity ",
					"maxLimit": "You can only carry up to 5 items at a time.",
					"code":     201,
				})
			}
		} else {
			c.JSON(202, gin.H{
				"status": "Fail",
				"error":  "product out of stock",
				"code":   202,
			})
		}
	}
}

// CartProductRemove decreases the quantity of a product in the user's cart if it's available.
// @Summary Decrease quantity of product in cart
// @Description Decreases the quantity of a product in the user's cart if it's available.
// @Tags Cart
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param ID path string true "Product ID"
// @Success 200 {string} string "one more quantity removed"
// @Failure 202 {string} string "can't remove one more"
// @Failure 400 {string} string "failed to update"
// @Failure 404 {string} string "can't find product"
// @Router /cart/{ID}/remove [patch]
func CartProductRemove(c *gin.Context) {
	var cartStore models.Cart
	session := sessions.Default(c)
	userid := session.Get("user_id")
	fmt.Println("=================", userid)
	// userId := c.GetUint("userid")
	id := c.Param("ID")
	err := initializer.DB.Where("user_id=? AND product_id=?", userid, id).First(&cartStore).Error
	if err != nil {
		c.JSON(404, gin.H{
			"status": "Fail",
			"error":  "can't find product",
			"code":   404,
		})
	} else {
		cartStore.Quantity -= 1
		if cartStore.Quantity >= 1 {
			err := initializer.DB.Where("user_id=? AND product_id=?", userid, cartStore.ProductId).Save(&cartStore)
			if err.Error != nil {
				c.JSON(400, gin.H{
					"status": "Fail",
					"error":  "failed to update",
					"code":   400,
				})
			} else {
				c.JSON(200, gin.H{
					"status":   "Success",
					"message":  "one more quantity removed",
					"quantity": cartStore.Quantity,
				})
			}
		} else {
			c.JSON(202, gin.H{
				"status": "Success",
				"error":  "can't remove one more",
			})
		}
	}
}

// CartProductDelete removes a product from the user's cart.
// @Summary Remove product from cart
// @Description Removes a product from the user's cart.
// @Tags Cart
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param ID path string true "Product ID"
// @Success 200 {string}  string "Item has been deleted."
// @Failure 400 {string}  string "Failed to delete item."
// @Failure 404 {string}  string "Can't find this item in your cart."
// @Router /cart/{ID}/delete [delete]
func CartProductDelete(c *gin.Context) {
	var ProductRemove models.Cart
	session := sessions.Default(c)
	userid := session.Get("user_id")
	fmt.Println("=================", userid)
	// userId := c.GetUint("userid")
	id := c.Param("ID")
	if err := initializer.DB.Where("product_id=? AND user_id=?", id, userid).First(&ProductRemove).Error; err != nil {
		c.JSON(404, gin.H{
			"status": "Fail",
			"error":  "Product not added to cart",
			"code":   404,
		})
	} else {
		if err := initializer.DB.Delete(&ProductRemove).Error; err != nil {
			c.JSON(400, gin.H{
				"status": "Fail",
				"error":  "Failed to delete item",
				"code":   400,
			})
			return
		}
		c.JSON(200, gin.H{
			"status":  "Success",
			"message": "product remove successfully",
		})
	}
}
