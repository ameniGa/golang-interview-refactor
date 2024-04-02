package controllers

type CartItemForm struct {
	Product  string `form:"product"   binding:"required"`
	Quantity string `form:"quantity"  binding:"required"`
}
