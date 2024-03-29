package calculator

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"interview/pkg/entity"
)

type Repository interface {
	GetCart(ctx context.Context, sessionID string) (entity.CartEntity, error)
	AddCart(ctx context.Context, cart *entity.CartEntity) error
	SaveItem(ctx context.Context, item *entity.CartItem) error
	GetItemByID(ctx context.Context, id uint) (entity.CartItem, error)
	GetItem(ctx context.Context, cartID uint, productName string) (entity.CartItem, error)
	GetItems(ctx context.Context, cartID uint) ([]entity.CartItem, error)
	DeleteItem(ctx context.Context, item entity.CartItem) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetCart(ctx context.Context, sessionID string) (entity.CartEntity, error) {
	var cartEntity entity.CartEntity
	result := r.db.WithContext(ctx).Where(fmt.Sprintf("status = '%s' AND session_id = '%s'", entity.CartOpen, sessionID)).First(&cartEntity)
	return cartEntity, result.Error
}

func (r *repository) AddCart(ctx context.Context, cart *entity.CartEntity) error {
	res := r.db.WithContext(ctx).Create(cart)
	return res.Error
}

func (r *repository) SaveItem(ctx context.Context, item *entity.CartItem) error {
	res := r.db.WithContext(ctx).Save(item)
	return res.Error
}

func (r *repository) GetItemByID(ctx context.Context, id uint) (entity.CartItem, error) {
	var cartItemEntity entity.CartItem
	res := r.db.WithContext(ctx).Where(" ID  = ?", id).First(&cartItemEntity)
	return cartItemEntity, res.Error
}

func (r *repository) GetItem(ctx context.Context, cartID uint, productName string) (entity.CartItem, error) {
	var cartItemEntity entity.CartItem
	res := r.db.WithContext(ctx).Where(" cart_id = ? and product_name  = ?", cartID, productName).First(&cartItemEntity)
	return cartItemEntity, res.Error
}

func (r *repository) GetItems(ctx context.Context, cartID uint) ([]entity.CartItem, error) {
	var cartItems []entity.CartItem
	res := r.db.WithContext(ctx).Where(fmt.Sprintf("cart_id = %d", cartID)).Find(&cartItems)
	return cartItems, res.Error
}

func (r *repository) DeleteItem(ctx context.Context, item entity.CartItem) error {
	res := r.db.WithContext(ctx).Delete(&item)
	return res.Error
}
