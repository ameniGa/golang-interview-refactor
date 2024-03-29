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
	if sessionID == "" || data.Quantity == "" || data.Product == "" {
		return Response{
			Code:  302,
			Error: errors.New("invalid arguments"),
		}
	}

	var isCartNew bool
	cartEntity, err := cal.getCart(ctx, sessionID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return Response{Code: 302, Error: err}
		}
		isCartNew = true
		cartEntity = entity.CartEntity{
			SessionID: sessionID,
			Status:    entity.CartOpen,
		}
		err := cal.addCart(ctx, &cartEntity)
		if err != nil {
			return Response{Code: 302, Error: err}
		}
	}

	item, ok := cal.priceMapping[data.Product]
	if !ok {
		return Response{
			Code:  302,
			Error: errors.New("invalid item name"),
		}
	}
	quantity, err := strconv.ParseInt(data.Quantity, 10, 0)
	if err != nil {
		return Response{
			Code:  302,
			Error: errors.New("invalid quantity"),
		}
	}

	if !isCartNew {
		cartItemEntity, err := cal.getItem(ctx, cartEntity.ID, data.Product)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return Response{
				Code: 302,
			}
		}
		if err == nil {
			cartItemEntity.Quantity += int(quantity)
			cartItemEntity.Price += item * float64(quantity)
			cal.saveItem(ctx, &cartItemEntity)
			return Response{
				Code: 302,
			}
		}
	}
	err = cal.saveItem(ctx, &entity.CartItem{
		CartID:      cartEntity.ID,
		ProductName: data.Product,
		Quantity:    int(quantity),
		Price:       item * float64(quantity),
	})
	return Response{
		Code:        302,
		RedirectURL: "/",
		Error:       err,
	}
}

func (cal *calculator) getCart(ctx context.Context, sessionID string) (entity.CartEntity, error) {
	var cartEntity entity.CartEntity
	result := cal.db.WithContext(ctx).Where(fmt.Sprintf("status = '%s' AND session_id = '%s'", entity.CartOpen, sessionID)).First(&cartEntity)
	return cartEntity, result.Error
}

func (cal *calculator) addCart(ctx context.Context, cart *entity.CartEntity) error {
	res := cal.db.WithContext(ctx).Create(cart)
	return res.Error
}

func (cal *calculator) saveItem(ctx context.Context, item *entity.CartItem) error {
	res := cal.db.WithContext(ctx).Save(item)
	return res.Error
}

func (cal *calculator) getItem(ctx context.Context, id uint, productName string) (entity.CartItem, error) {
	var cartItemEntity entity.CartItem
	res := cal.db.Where(" cart_id = ? and product_name  = ?", id, productName).First(&cartItemEntity)
	return cartItemEntity, res.Error
}

func (cal *calculator) getItemByID(ctx context.Context, id uint) (entity.CartItem, error) {
	var cartItemEntity entity.CartItem
	res := cal.db.WithContext(ctx).Where(" ID  = ?", id).First(&cartItemEntity)
	return cartItemEntity, res.Error
}

func (cal *calculator) deleteItem(ctx context.Context, item entity.CartItem) error {
	res := cal.db.WithContext(ctx).Delete(&item)
	return res.Error
}

func (cal *calculator) DeleteCartItem(ctx context.Context, sessionID, cartItemID string) Response {
	cartEntity, err := cal.getCart(ctx, sessionID)
	if err != nil {
		return Response{Code: 302, RedirectURL: "/"}
	}

	if cartEntity.Status == entity.CartClosed {
		return Response{Code: 302, RedirectURL: "/"}
	}

	_cartItemID, err := strconv.Atoi(cartItemID)
	if err != nil {
		return Response{Code: 302, RedirectURL: "/"}
	}

	cartItemEntity, err := cal.getItemByID(ctx, uint(_cartItemID))
	if err != nil {
		return Response{Code: 302, RedirectURL: "/"}
	}

	err = cal.deleteItem(ctx, cartItemEntity)
	return Response{Code: 302, RedirectURL: "/", Error: err}
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
		// TODO should return error in case of some error != recordNotFound
		return
	}

	var cartItems []entity.CartItem
	result = db.WithContext(ctx).Where(fmt.Sprintf("cart_id = %d", cartEntity.ID)).Find(&cartItems)
	if result.Error != nil {
		// TODO should return error in case of some error != recordNotFound
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
