package calculator

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"html/template"
	db2 "interview/pkg/db"
	"interview/pkg/entity"
	"log"
	"strconv"
	"strings"
)

type Handler interface {
	AddItemToCart(ctx context.Context, sessionID string, data CartItem) Response
	DeleteCartItem(c *gin.Context)
	GetCartData(c *gin.Context)
}

type calculator struct {
	db           *gorm.DB
	priceMapping map[string]float64
}

func NewCalculator(itemPriceMapping map[string]float64, db *gorm.DB) Handler {
	return &calculator{db: db, priceMapping: itemPriceMapping}
}

type CartItem struct {
	Product  string
	Quantity string
}

func (cal *calculator) AddItemToCart(ctx context.Context, sessionID string, data CartItem) Response {
	db := cal.db

	var isCartNew bool
	var cartEntity entity.CartEntity
	result := db.WithContext(ctx).Where(fmt.Sprintf("status = '%s' AND session_id = '%s'", entity.CartOpen, sessionID)).First(&cartEntity)

	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return Response{
				Code:        302,
				RedirectURL: "/",
			}
		}
		isCartNew = true
		cartEntity = entity.CartEntity{
			SessionID: sessionID,
			Status:    entity.CartOpen,
		}
		db.Create(&cartEntity)
	}

	item, ok := cal.priceMapping[data.Product]
	if !ok {
		return Response{
			Code:        302,
			RedirectURL: "/?error=invalid item name",
		}
	}

	quantity, err := strconv.ParseInt(data.Quantity, 10, 0)
	if err != nil {
		return Response{
			Code:        302,
			RedirectURL: "/?error=invalid quantity",
		}
	}

	var cartItemEntity entity.CartItem
	if isCartNew {
		cartItemEntity = entity.CartItem{
			CartID:      cartEntity.ID,
			ProductName: data.Product,
			Quantity:    int(quantity),
			Price:       item * float64(quantity),
		}
		db.Create(&cartItemEntity)
	} else {
		result = db.Where(" cart_id = ? and product_name  = ?", cartEntity.ID, data.Product).First(&cartItemEntity)

		if result.Error != nil {
			if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return Response{
					Code:        302,
					RedirectURL: "/",
				}
			}
			cartItemEntity = entity.CartItem{
				CartID:      cartEntity.ID,
				ProductName: data.Product,
				Quantity:    int(quantity),
				Price:       item * float64(quantity),
			}
			db.Create(&cartItemEntity)

		} else {
			cartItemEntity.Quantity += int(quantity)
			cartItemEntity.Price += item * float64(quantity)
			db.Save(&cartItemEntity)
		}
	}
	return Response{
		Code:        302,
		RedirectURL: "/",
	}
}

func (cal *calculator) DeleteCartItem(c *gin.Context) {
	cartItemIDString := c.Query("cart_item_id")
	if cartItemIDString == "" {
		c.Redirect(302, "/")
		return
	}

	cookie, _ := c.Request.Cookie("ice_session_id")

	db := cal.db

	var cartEntity entity.CartEntity
	result := db.Where(fmt.Sprintf("status = '%s' AND session_id = '%s'", entity.CartOpen, cookie.Value)).First(&cartEntity)
	if result.Error != nil {
		c.Redirect(302, "/")
		return
	}

	if cartEntity.Status == entity.CartClosed {
		c.Redirect(302, "/")
		return
	}

	cartItemID, err := strconv.Atoi(cartItemIDString)
	if err != nil {
		c.Redirect(302, "/")
		return
	}

	var cartItemEntity entity.CartItem

	result = db.Where(" ID  = ?", cartItemID).First(&cartItemEntity)
	if result.Error != nil {
		c.Redirect(302, "/")
		return
	}

	db.Delete(&cartItemEntity)
	c.Redirect(302, "/")
}

func (cal *calculator) GetCartData(c *gin.Context) {
	data := map[string]interface{}{
		"Error": c.Query("error"),
		//"cartItems": cartItems,
	}

	cookie, err := c.Request.Cookie("ice_session_id")
	if err == nil {
		data["CartItems"] = getCartItemData(cookie.Value)
	}

	html, err := renderTemplate(data)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(500)
		return
	}

	c.Header("Content-Type", "text/html")
	c.String(200, html)
}

func getCartItemData(sessionID string) (items []map[string]interface{}) {
	db := db2.GetDatabase()
	var cartEntity entity.CartEntity
	result := db.Where(fmt.Sprintf("status = '%s' AND session_id = '%s'", entity.CartOpen, sessionID)).First(&cartEntity)

	if result.Error != nil {
		return
	}

	var cartItems []entity.CartItem
	result = db.Where(fmt.Sprintf("cart_id = %d", cartEntity.ID)).Find(&cartItems)
	if result.Error != nil {
		return
	}

	for _, cartItem := range cartItems {
		item := map[string]interface{}{
			"ID":       cartItem.ID,
			"Quantity": cartItem.Quantity,
			"Price":    cartItem.Price,
			"Product":  cartItem.ProductName,
		}

		items = append(items, item)
	}
	return items
}

func renderTemplate(pageData interface{}) (string, error) {
	// Read and parse the HTML template file
	tmpl, err := template.ParseFiles("static/add_item_form.html")
	if err != nil {
		return "", fmt.Errorf("Error parsing template: %v ", err)
	}

	// Create a strings.Builder to store the rendered template
	var renderedTemplate strings.Builder

	err = tmpl.Execute(&renderedTemplate, pageData)
	if err != nil {
		return "", fmt.Errorf("Error parsing template: %v ", err)
	}

	// Convert the rendered template to a string
	resultString := renderedTemplate.String()

	return resultString, nil
}
