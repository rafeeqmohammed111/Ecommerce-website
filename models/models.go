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
	ID          uint `gorm:"primary_key"`
	Stock      uint
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
	Id                 string       `json:"id"`
	UserId             uint          `json:"user_id"`
	User               Users        `json:"user" gorm:"foreignKey:UserId"`
	AddressId          int          `json:"address_id"`
	Address            Address      `json:"address" gorm:"foreignKey:AddressId"`
	CouponCode         string       `json:"coupon_code"`
	OrderPaymentMethod string       `json:"order_payment_method"`
	OrderAmount        float64      `json:"order_amount"`
	TotalAmount        float64      `json:"total_amount"`
	ShippingCharge     float32      `json:"shipping_charge"`
	OrderStatus        string       `json:"order_status"`
	OrderDate          time.Time    `json:"order_date"`
	OrderUpdate        time.Time    `json:"order_update"`
	CouponID           *uint        `json:"coupon_id"`
	Coupon             *Coupon      `json:"coupon" gorm:"foreignKey:CouponID"`
	OrderItems         []OrderItems `json:"order_items" gorm:"foreignKey:OrderId"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at"`
}

type OrderItems struct {
	Id                uint      `gorm:"primaryKey" json:"id"`
	OrderId           string    `json:"order_id"`
	Order             Order     `json:"order" gorm:"foreignKey:OrderId"`
	ProductId         int       `json:"product_id"`
	Product           Products  `json:"product" gorm:"foreignKey:ProductId"`
	Quantity          uint      `json:"quantity"`
	SubTotal          float64   `json:"sub_total"`
	OrderStatus       string    `json:"order_status"`
	OrderCancelReason string    `json:"order_cancel_reason"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	ProductName   string
	Category        string
}

type Wallet struct {
	gorm.Model
	UserId  uint    `json:"user_id"`
	User    Users   `json:"user" gorm:"foreignKey:UserId"`
	Balance float64 `json:"balance"`
	// Code     string  `json:"code"`
	// Discount float64 `json:"discount"`
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
	ID           uint      `json:"id"`
	ProductID    uint      `json:"product_id"`  // Associated Product ID
	CategoryID   uint      `json:"category_id"` // Associated Category ID
	SpecialOffer string    `json:"special_offer"`
	Discount     float64   `json:"discount"`
	ValidFrom    time.Time `json:"valid_from"`
	ValidTo      time.Time `json:"valid_to"`
}

type PaymentDetails struct {
	gorm.Model
	OrderID       string    `gorm:"index;not null" json:"order_id"`
	PaymentId     string    `gorm:"not null" json:"payment_id"`
	PaymentStatus string    `gorm:"not null" json:"payment_status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Wishlist struct {
	Id        uint
	UserId    int
	User      Users
	ProductId int
	Product   Products
}
type SimplifiedProduct struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Price    uint   `json:"price"`
	Quantity int    `json:"quantity"`
}
