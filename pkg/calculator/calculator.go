package calculator

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"interview/pkg/entity"
	"strconv"
)

type Handler interface {
	AddItemToCart(ctx context.Context, sessionID string, data CartItem) Response
	DeleteCartItem(ctx context.Context, sessionID, cartItemID string) Response
	GetCartData(ctx context.Context, sessionID string) Response
}

type calculator struct {
	db           *gorm.DB
	priceMapping map[string]float64
}

func NewCalculator(itemPriceMapping map[string]float64, db *gorm.DB) (Handler, error) {
	if itemPriceMapping == nil {
		return nil, errors.New("missing prices config")
	}
	if db == nil {
		panic("invalid db connection")
	}
	return &calculator{db: db, priceMapping: itemPriceMapping}, nil
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
			return Response{Code: 302, RedirectURL: "/"}
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

func (cal *calculator) DeleteCartItem(ctx context.Context, sessionID, cartItemID string) Response {
	db := cal.db

	var cartEntity entity.CartEntity
	result := db.WithContext(ctx).Where(fmt.Sprintf("status = '%s' AND session_id = '%s'", entity.CartOpen, sessionID)).First(&cartEntity)
	if result.Error != nil {
		return Response{Code: 302, RedirectURL: "/"}
	}

	if cartEntity.Status == entity.CartClosed {
		return Response{Code: 302, RedirectURL: "/"}
	}

	_cartItemID, err := strconv.Atoi(cartItemID)
	if err != nil {
		return Response{Code: 302, RedirectURL: "/"}
	}

	var cartItemEntity entity.CartItem

	result = db.WithContext(ctx).Where(" ID  = ?", _cartItemID).First(&cartItemEntity)
	if result.Error != nil {
		return Response{Code: 302, RedirectURL: "/"}
	}

	db.Delete(&cartItemEntity)
	return Response{Code: 302, RedirectURL: "/"}
}

func (cal *calculator) GetCartData(ctx context.Context, sessionID string) Response {
	if sessionID == "" {
		return Response{
			Code: 200,
		}
	}
	data := cal.getCartItemData(ctx, sessionID)
	return Response{
		Code: 200,
		Data: data,
	}
}

func (cal *calculator) getCartItemData(ctx context.Context, sessionID string) (items []map[string]interface{}) {
	db := cal.db

	var cartEntity entity.CartEntity
	result := db.WithContext(ctx).Where(fmt.Sprintf("status = '%s' AND session_id = '%s'", entity.CartOpen, sessionID)).First(&cartEntity)
	if result.Error != nil {
		return
	}

	var cartItems []entity.CartItem
	result = db.WithContext(ctx).Where(fmt.Sprintf("cart_id = %d", cartEntity.ID)).Find(&cartItems)
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
