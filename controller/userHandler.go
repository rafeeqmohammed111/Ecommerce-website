package controller

import (
	"fmt"
	"net/http"
	"project/handler"
	"project/initializer"
	"project/models"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	// "golang.org/x/crypto/bcrypt"
)

var LogJs models.Users
var otp string
var newUser models.Users

func LoadingPage(c *gin.Context) {
	c.JSON(200, gin.H{"name": "Welcome to loading page"})
}
func GetAllProducts(c *gin.Context) {
	var products []models.Products
	if err := initializer.DB.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}


func UserSignUp(c *gin.Context) {
	// var newUser models.Users
	var otpStore models.OtpMail
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(501, gin.H{"error": "json binding error"})
		return
	}

	err := initializer.DB.First(&models.Users{}, "email=?", newUser.Email).Error
	if err == nil {
		c.JSON(501, gin.H{"error": "Email address already exist"})
		return
	}

	// Generate and send OTP
	otp = handler.GenerateOtp()
	fmt.Println("----------------", otp, "-----------------")

	err = handler.SendOtp(newUser.Email, otp)
	if err != nil {
		c.JSON(500, "failed to send otp")
		return
	}

	otpStore = models.OtpMail{
		Otp:       otp,
		Email:     newUser.Email,
		CreatedAt: time.Now(),
		ExpireAt:  time.Now().Add(180 * time.Second),
	}
	if err := initializer.DB.Create(&otpStore).Error; err != nil {
		c.JSON(500, gin.H{"error": "failed to save otp details"})
		return
	}

	c.JSON(200, "OTP sent to email. Please verify to complete registration.")
}

func OtpCheck(c *gin.Context) {
	var otpcheck models.OtpMail

	if err := c.ShouldBindJSON(&otpcheck); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to bind otp details"})
		return
	}

	var otpExistTable models.OtpMail
	if err := initializer.DB.Find(&otpExistTable, "email = ?", otpcheck.Email).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "OTP not found for this email"})
		return
	}

	if err := initializer.DB.Where("otp = ? AND expire_at > ?", otpcheck.Otp, time.Now()).First(&otpExistTable).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP"})
		return
	}

	fmt.Println("correct otp")

	// Retrieve the temporarily stored user details
	// var newUser models.Users
	if err := initializer.DB.Where("email = ?", newUser.Email).Find(&otpcheck).Error; err != nil {
		c.JSON(500, "failed to retrieve user details")
		return
	}

	// Save the user details after OTP verification
	if err := initializer.DB.Create(&newUser).Error; err != nil {
		c.JSON(500, "failed to save user details")
		return
	}

	// Delete the OTP entry after successful verification
	if err := initializer.DB.Delete(&otpExistTable).Error; err != nil {
		c.JSON(500, "failed to delete otp data")
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "User created successfully"})
}

func ResendOtp(c *gin.Context) {
	var otpStore models.OtpMail
	otp = handler.GenerateOtp()

	err := handler.SendOtp(newUser.Email, otp)
	if err != nil {
		c.JSON(500, "failed to send otp")
		return
	}
	c.JSON(200, "otp send to mail  "+otp)

	// saving/upd deatails in the database

	result := initializer.DB.First(&otpStore, "email=?", newUser.Email)
	if result.Error != nil {
		otpStore = models.OtpMail{
			Otp:       otp,
			Email:     newUser.Email,
			CreatedAt: time.Now(),
			ExpireAt:  time.Now().Add(60 * time.Second),
		}

		err := initializer.DB.Create(&otpStore)
		if err.Error != nil {
			c.JSON(500, gin.H{"error": "failed to save otp details"})
			return
		}

	} else {
		err := initializer.DB.Model(&otpStore).Where("email=?", newUser.Email).Updates(models.OtpMail{
			Otp:      otp,
			ExpireAt: time.Now().Add(15 * time.Second),
		})
		if err.Error != nil {
			c.JSON(500, "failed too update data")
			return
		}
	}

}



func UserLogin(c *gin.Context) {

	var loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {

		fmt.Println("Error binding JSON:", err)

		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to bind login details"})
		return
	}

	fmt.Println("Login Request Email:", loginRequest.Email)

	var userFromDB models.Users
	if err := initializer.DB.Where("email=?", loginRequest.Email).First(&userFromDB).Error; err != nil {

		fmt.Println("fetching data faild from DB:", err)

		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})

		return
	}

	if userFromDB.Password != loginRequest.Password {
		fmt.Println("Password mismatch")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid  password"})
		return
	}

	if userFromDB.Blocking {
		c.JSON(http.StatusForbidden, gin.H{"error": "user blocked"})
		return
	}
	session := sessions.Default(c)
	session.Set("user_email", userFromDB.Email)
	session.Set("user_id", userFromDB.ID)
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "login successful"})
}

func UserLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear() // Clear all session data
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "successfully logged out"})
}
