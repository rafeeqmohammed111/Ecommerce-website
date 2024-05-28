package admin

import (
	"net/http"
	"os"
	"path/filepath"
	"project/initializer"
	"project/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AdminPage(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Welcome to the admin page"})
}

func AdminLogin(c *gin.Context) {
	var LogJs struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	err := c.ShouldBindJSON(&LogJs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error binding data"})
		return
	}

	if LogJs.Username == "rafeeqmohammed111" && LogJs.Password == "rafeeq@123" {
		SessionCreation(LogJs.Username, "admin", c)
		c.JSON(http.StatusAccepted, gin.H{"message": "Successfully logged in"})
		return
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
	}
}

func AdminLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear() // Clear all session data
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log out"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

func SessionCreation(email string, role string, c *gin.Context) {
	session := sessions.Default(c)
	session.Set("admin", email)
	session.Set("role", role)
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create the session",
		})
		return
	}
}

func UserList(c *gin.Context) {
	var user_management []models.Users
	initializer.DB.Order("ID").Find(&user_management)

	var userList []gin.H
	for _, val := range user_management {
		user := gin.H{
			"ID":       val.ID,
			"name":     val.Name,
			"username": val.Username,
			"Email":    val.Email,
			"gender":   val.Gender,
			"status":   val.Blocking,
		}
		userList = append(userList, user)
	}
	c.JSON(200, userList)
}

func EditUserDetails(c *gin.Context) {
	var userEdit models.Users
	id := c.Param("ID")
	err := initializer.DB.First(&userEdit, id)
	if err.Error != nil {
		c.JSON(500, gin.H{"error": "Can't find user"})
	} else {
		err := c.ShouldBindJSON(&userEdit)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to bind data"})
		} else {
			if err := initializer.DB.Save(&userEdit).Error; err != nil {
				c.JSON(500, gin.H{"error": "Failed to update details"})
			} else {
				c.JSON(200, gin.H{"message": "User updated successfully"})
			}
		}
	}
}

func BlockUser(c *gin.Context) {
	var blockUser models.Users
	id := c.Param("ID")
	err := initializer.DB.First(&blockUser, id)
	if err.Error != nil {
		c.JSON(500, gin.H{"error": "Can't find user"})
	} else {
		if blockUser.Blocking {
			blockUser.Blocking = false
			c.JSON(200, gin.H{"message": "User unblocked"})
		} else {
			blockUser.Blocking = true
			c.JSON(200, gin.H{"message": "User blocked"})
		}
		if err := initializer.DB.Save(&blockUser).Error; err != nil {
			c.JSON(500, gin.H{"error": "Failed to block/unblock user"})
		}
	}
}

func DeleteUser(c *gin.Context) {
	var deleteUser models.Users
	id := c.Param("ID")
	err := initializer.DB.First(&deleteUser, id)
	if err.Error != nil {
		c.JSON(500, gin.H{"error": "Can't find user"})
	} else {
		err := initializer.DB.Delete(&deleteUser).Error
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to delete user"})
		} else {
			c.JSON(200, gin.H{"message": "User deleted successfully"})
		}
	}
}

//-------------------product management-----------------

func ProductList(c *gin.Context) {
	var productList []models.Products
	err := initializer.DB.Joins("Category").Find(&productList).Error
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch details"})
	} else {
		for _, val := range productList {
			if !val.Category.Blocking {
				continue
			} else {
				c.JSON(200, gin.H{
					"Product Id":       val.ID,
					"Product Name":     val.Name,
					"Product Price":    val.Price,
					"Product Size":     val.Size,
					"Product Color":    val.Color,
					"Product Quantity": val.Quantity,
					"Category name":    val.Category.Category_name,
					"Product Image":    val.ImagePath,
					"Product Status":   val.Status,
					"category id":      val.CategoryId,
				})
			}
		}
	}
}

func UploadProductImage(c *gin.Context) {
	file, err := c.FormFile("p_imagepath")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file"})
		return
	}
	imageDir := "./images/"
	if err := os.MkdirAll(imageDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}
	imagePath := filepath.Join(imageDir, file.Filename)
	if err := c.SaveUploadedFile(file, imagePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload photo"})
		return
	}
	productID := c.PostForm("product_id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}
	if err := initializer.DB.Model(&models.Products{}).Where("id = ?", productID).Update("image_path", imagePath).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "file_path": imagePath})
}

func AddProducts(c *gin.Context) {
	var addProduct models.Products
	err := c.ShouldBindJSON(&addProduct)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON binding error"})
		return
	}

	var checkCategory models.Category
	if err := initializer.DB.First(&checkCategory, addProduct.CategoryId).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No category found"})
		return
	}

	addProduct.Status = true
	if result := initializer.DB.Create(&addProduct); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product created successfully. Please upload the product image.", "productId": addProduct.ID})
}

func EditProducts(c *gin.Context) {
	var editProducts models.Products
	id := c.Param("ID")
	err := initializer.DB.First(&editProducts, id)
	if err.Error != nil {
		c.JSON(500, gin.H{"error": "Can't find product"})
	} else {
		err := c.ShouldBindJSON(&editProducts)
		if err != nil {
			c.JSON(500, "Failed to bind details")
		} else {
			if err := initializer.DB.Save(&editProducts).Error; err != nil {
				c.JSON(500, "Failed to edit details")
			} else {
				c.JSON(200, "Successfully edited product")
			}
		}
	}
}

func DeleteProducts(c *gin.Context) {
	var deleteProducts models.Products
	id := c.Param("ID")
	err := initializer.DB.First(&deleteProducts, id)
	if err.Error != nil {
		c.JSON(500, gin.H{"error": "Can't find product"})
	} else {
		err := initializer.DB.Delete(&deleteProducts).Error
		if err != nil {
			c.JSON(500, "Failed to delete product")
		} else {
			c.JSON(200, "Product deleted successfully")
		}
	}
}
