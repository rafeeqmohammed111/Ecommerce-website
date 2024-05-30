package user

import (
	"fmt"
	"project/initializer"
	"project/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// UserProfile returns details of the authenticated user.
func UserProfile(c *gin.Context) {
	session := sessions.Default(c)

	fmt.Printf("Session ID: %v\n", session.Get("user_id"))
	fmt.Printf("Session role: %v\n", session.Get("role"))
	userId := session.Get("user_id")
	var user models.Users
	if err := initializer.DB.Preload("Addresses").Find(&user, userId).Error; err != nil {
		c.JSON(500, gin.H{
			"status": "fail",
			"error":  "failed to find user",
		})
		return
	}

	userData := gin.H{
		"name":      user.Name,
		"email":     user.Email,
		"gender":    user.Gender,
		"phone":     user.Phone,
		"addresses": user.Addresses,
	}

	c.JSON(200, gin.H{
		"status": "success",
		"data":   userData,
	})
}

type addressUpdate struct {
	Address string `json:"address"`
	City    string `json:"city"`
	State   string `json:"state"`
	Pincode int    `json:"pincode"`
	Country string `json:"country"`
	Phone   int    `json:"phone"`
}

// AddressStore adds a new address for the authenticated user.
func AddressStore(c *gin.Context) {
	session := sessions.Default(c)
	userId := session.Get("user_id")
	fmt.Println("====================================", userId)
	var addressBind addressUpdate
	if err := c.BindJSON(&addressBind); err != nil {
		c.JSON(400, gin.H{
			"status": "fail",
			"error":  "failed to bind data",
		})
		return
	}

	address := models.Address{
		Address: addressBind.Address,
		City:    addressBind.City,
		State:   addressBind.State,
		Country: addressBind.Country,
		Pincode: addressBind.Pincode,
		Phone:   addressBind.Phone,
		UserID:  userId.(uint),
	}

	if err := initializer.DB.Create(&address).Error; err != nil {
		c.JSON(400, gin.H{
			"status": "fail",
			"error":  "failed to create address",
		})
		return
	}

	c.JSON(201, gin.H{
		"status":  "success",
		"message": "New address added successfully",
	})
}

// AddressEdit updates an existing address for the authenticated user.
func AddressEdit(c *gin.Context) {
	var address models.Address
	id := c.Param("id")
	if err := initializer.DB.First(&address, id).Error; err != nil {
		c.JSON(404, gin.H{
			"status": "fail",
			"error":  "failed to find address",
		})
		return
	}

	var addressBind addressUpdate
	if err := c.BindJSON(&addressBind); err != nil {
		c.JSON(400, gin.H{
			"status": "fail",
			"error":  "failed to bind data",
		})
		return
	}

	address.Address = addressBind.Address
	address.City = addressBind.City
	address.State = addressBind.State
	address.Country = addressBind.Country
	address.Pincode = addressBind.Pincode
	address.Phone = addressBind.Phone

	if err := initializer.DB.Save(&address).Error; err != nil {
		c.JSON(400, gin.H{
			"status": "fail",
			"error":  "failed to update address",
		})
		return
	}

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "address updated successfully",
	})
}

// AddressDelete deletes an existing address for the authenticated user.
func AddressDelete(c *gin.Context) {
	var address models.Address
	id := c.Param("id")
	if err := initializer.DB.First(&address, id).Error; err != nil {
		c.JSON(404, gin.H{
			"status": "fail",
			"error":  "failed to find address",
		})
		return
	}

	if err := initializer.DB.Delete(&address).Error; err != nil {
		c.JSON(400, gin.H{
			"status": "fail",
			"error":  "failed to delete address",
		})
		return
	}

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Address deleted successfully",
	})
}

type userDetailUpdate struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Phone  int    `json:"phone"`
	Gender string `json:"gender"`
}

// EditUserProfile updates the profile of the authenticated user.
func EditUserProfile(c *gin.Context) {
	session := sessions.Default(c)
	userId := session.Get("user_id")
	var user models.Users
	if err := initializer.DB.First(&user, userId).Error; err != nil {
		c.JSON(404, gin.H{
			"status": "fail",
			"error":  "user not found",
		})
		return
	}

	var userUpdate userDetailUpdate
	if err := c.BindJSON(&userUpdate); err != nil {
		c.JSON(400, gin.H{
			"status": "fail",
			"error":  "failed to bind data",
		})
		return
	}

	user.Name = userUpdate.Name
	user.Email = userUpdate.Email
	user.Phone = userUpdate.Phone
	user.Gender = userUpdate.Gender

	if err := initializer.DB.Save(&user).Error; err != nil {
		c.JSON(400, gin.H{
			"status": "fail",
			"error":  "failed to update data",
		})
		return
	}

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "user profile updated successfully",
	})
}
