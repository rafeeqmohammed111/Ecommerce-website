package models

import (
	"time"

	"gorm.io/gorm"
)

type Users struct {
	ID        uint      `json:"userid" gorm:"primaryKey"`
	Name      string    `gorm:"not null" json:"name"`
	Username  string    `gorm:"not null" json:"username"`
	Email     string    `gorm:"not null" json:"email"`
	Password  string    `gorm:"not null" json:"password"`
	Gender    string    `json:"gender"`
	Phone     int       `gorm:"not null" json:"phone"`
	Blocking  bool      `json:"blocking"`
	Addresses []Address `gorm:"foreignkey:UserID"`
}

type OtpMail struct {
	Id        uint
	Email     string `gorm:"unique" json:"email"`
	Otp       string `gorm:"not null" json:"otp"`
	CreatedAt time.Time
	ExpireAt  time.Time `gorm:"type:timestamp;not null"`
}
type Products struct {
	gorm.Model
	Name        string `gorm:"unique" json:"p_name"`
	Price       uint   `json:"p_price"`
	Size        string `json:"p_size"`
	Color       string `json:"p_color"`
	Quantity    int    `json:"p_quantity"`
	Description string `json:"p_description"`
	ImagePath   string `json:"p_imagepath"`
	Status      bool   `json:"p_blocking"`
	CategoryId  int    `json:"category_id"`
	Category    Category
}
type Category struct {
	gorm.Model
	ID                   uint   `json:"id" gorm:"primaryKey"`
	Category_name        string `gorm:"not null" json:"category_name"`
	Category_description string `gorm:"not null" json:"category_description"`
	Blocking             bool   `gorm:"not null" json:"category_blocking"`
}
type Address struct {
	ID      uint   `gorm:"primary_key"`
	Address string `gorm:"size:255"`
	City    string `gorm:"size:255"`
	State   string `gorm:"size:255"`
	Country string `gorm:"size:255"`
	Pincode int
	Phone   int
	UserID  uint // Foreign key to link to the User model
}
type Cart struct {
	Id        uint
	UserId    uint `json:"userid"`
	User      Users
	ProductId int `json:"product_id"`
	Product   Products
	Quantity  uint
}
type Order struct {
	Id                 uint
	UserId             int `json:"orderId"`
	User               Users
	AddressId          int `json:"orderAddress"`
	Address            Address
	CouponCode         string `json:"orderCoupon"`
	OrderPaymentMethod string `json:"orderPayment"`
	OrderAmount        float64
	ShippingCharge     float32
	OrderDate          time.Time
	OrderUpdate        time.Time
}
type OrderItems struct {
	Id                uint `gorm:"primary key"`
	OrderId           uint
	Order             Order
	ProductId         int
	Product           Products
	Quantity          uint
	SubTotal          float64
	OrderStatus       string
	OrderCancelReason string
}
type Wallet struct {
	gorm.Model
	User_id int
	User    Users
	Balance float64
}
type Coupon struct {
	ID              uint      `gorm:"primarykey"`
	Code            string    `gorm:"unique" json:"code"`
	Discount        float64   `json:"discount"`
	CouponCondition int       `json:"condition"`
	ValidFrom       time.Time `json:"valid_from"`
	ValidTo         time.Time `json:"valid_to"`
}
type Offer struct {
	Id           uint
	ProductId    int       `json:"productid"`
	SpecialOffer string    `json:"offer"`
	Discount     float64   `json:"discount"`
	ValidFrom    time.Time `json:"valid_from"`
	ValidTo      time.Time `json:"valid_to"`
}
type PaymentDetails struct {
	gorm.Model
	PaymentId     string
	Order_Id      string
	Receipt       uint
	PaymentStatus string
	PaymentAmount float64
}
type Wishlist struct {
	Id        uint
	UserId    int
	User      Users
	ProductId int
	Product   Products
}