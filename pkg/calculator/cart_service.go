package calculator

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"interview/pkg/entity"
	"strconv"
)

// Handler encapsulates the behaviour of cart service
type Handler interface {
	AddItemToCart(ctx context.Context, sessionID string, data CartItem) Response
	DeleteCartItem(ctx context.Context, sessionID, cartItemID string) Response
	GetCartData(ctx context.Context, sessionID string) Response
}

type cartService struct {
	repo         Repository
	priceMapping map[string]float64
}

// NewCartService creates a cart service implementation
func NewCartService(itemPriceMapping map[string]float64, repo Repository) (Handler, error) {
	if itemPriceMapping == nil {
		return nil, errors.New("missing prices config")
	}
	if repo == nil {
		panic("missing repository")
	}
	return &cartService{repo: repo, priceMapping: itemPriceMapping}, nil
}

type CartItem struct {
	Product  string
	Quantity string
}

func (c *cartService) AddItemToCart(ctx context.Context, sessionID string, data CartItem) Response {
	if sessionID == "" || data.Quantity == "" || data.Product == "" {
		return Response{
			Code:  302,
			Error: errors.New("invalid arguments"),
		}
	}

	var isCartNew bool
	cartEntity, err := c.repo.GetCart(ctx, sessionID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return Response{Code: 302, Error: err}
		}
		isCartNew = true
		cartEntity = entity.CartEntity{
			SessionID: sessionID,
			Status:    entity.CartOpen,
		}
		err := c.repo.AddCart(ctx, &cartEntity)
		if err != nil {
			return Response{Code: 302, Error: err}
		}
	}

	item, ok := c.priceMapping[data.Product]
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
		cartItemEntity, err := c.repo.GetItem(ctx, cartEntity.ID, data.Product)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return Response{
				Code:  302,
				Error: err,
			}
		}
		if err == nil {
			cartItemEntity.Quantity += int(quantity)
			cartItemEntity.Price += item * float64(quantity)
			err = c.repo.SaveItem(ctx, &cartItemEntity)
			return Response{
				Code:  302,
				Error: err,
			}
		}
	}
	err = c.repo.SaveItem(ctx, &entity.CartItem{
		CartID:      cartEntity.ID,
		ProductName: data.Product,
		Quantity:    int(quantity),
		Price:       item * float64(quantity),
	})
	return Response{
		Code:  302,
		Error: err,
	}
}

func (c *cartService) DeleteCartItem(ctx context.Context, sessionID, cartItemID string) Response {
	cartEntity, err := c.repo.GetCart(ctx, sessionID)
	if err != nil {
		return Response{Code: 302}
	}

	if cartEntity.Status == entity.CartClosed {
		return Response{Code: 302}
	}

	_cartItemID, err := strconv.Atoi(cartItemID)
	if err != nil {
		return Response{Code: 302}
	}

	cartItemEntity, err := c.repo.GetItemByID(ctx, uint(_cartItemID))
	if err != nil {
		return Response{Code: 302}
	}

	err = c.repo.DeleteItem(ctx, cartItemEntity)
	return Response{Code: 302, Error: err}
}

func (c *cartService) GetCartData(ctx context.Context, sessionID string) Response {
	if sessionID == "" {
		return Response{
			Code: 200,
		}
	}

	cartEntity, err := c.repo.GetCart(ctx, sessionID)
	if err != nil {
		// TODO should return error in case of some error != recordNotFound
		return Response{
			Code: 200,
		}
	}

	cartItems, err := c.repo.GetItems(ctx, cartEntity.ID)
	if err != nil {
		// TODO should return error in case of some error != recordNotFound
		return Response{
			Code: 200,
		}
	}

	data := make([]map[string]interface{}, 0)
	for _, cartItem := range cartItems {
		item := map[string]interface{}{
			"ID":       cartItem.ID,
			"Quantity": cartItem.Quantity,
			"Price":    cartItem.Price,
			"Product":  cartItem.ProductName,
		}

		data = append(data, item)
	}
	return Response{
		Code: 200,
		Data: data,
	}
}
