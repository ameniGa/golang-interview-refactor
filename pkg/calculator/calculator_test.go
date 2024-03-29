package calculator_test

import (
	"context"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	. "interview/pkg/calculator"
	"interview/pkg/entity"
	"log"
	"strconv"
	"testing"
)

var itemPriceMapping = map[string]float64{
	"shoe":  100,
	"purse": 200,
	"bag":   300,
	"watch": 300,
}

func TestNewCalculator(t *testing.T) {
	dbMock, _ := NewMockDB()
	t.Run("nil prices map", func(t *testing.T) {
		_, err := NewCalculator(nil, dbMock)
		assert.Error(t, err)
	})
	t.Run("nil db", func(t *testing.T) {
		assert.Panics(t, func() {
			_, _ = NewCalculator(map[string]float64{"bag": 10}, nil)
		})
	})
	t.Run("valid ", func(t *testing.T) {
		h, _ := NewCalculator(map[string]float64{"bag": 10}, dbMock)
		assert.NotNil(t, h)
	})
}

func TestCalculator_AddItemToCart(t *testing.T) {
	gormDB, mock := NewMockDB()
	calc, _ := NewCalculator(itemPriceMapping, gormDB)
	sessionID := "123456"

	t.Run("missing sessionID", func(t *testing.T) {
		res := calc.AddItemToCart(context.TODO(), "", CartItem{
			Product:  "shoe",
			Quantity: "1",
		})
		assert.Error(t, res.Error)
	})

	t.Run("invalid product name", func(t *testing.T) {
		cart := sqlmock.NewRows([]string{"id", "session_id", "status"}).
			AddRow(1, sessionID, entity.CartOpen)
		mock.ExpectQuery("SELECT (.+)").WillReturnRows(cart)

		res := calc.AddItemToCart(context.TODO(), sessionID, CartItem{
			Product:  "non_existing_product",
			Quantity: "1",
		})
		assert.Nil(t, mock.ExpectationsWereMet())
		assert.Error(t, res.Error)
	})

	t.Run("invalid quantity", func(t *testing.T) {
		cart := sqlmock.NewRows([]string{"id", "session_id", "status"}).
			AddRow(1, sessionID, entity.CartOpen)
		mock.ExpectQuery("SELECT (.+)").WillReturnRows(cart)

		res := calc.AddItemToCart(context.TODO(), sessionID, CartItem{
			Product:  "purse",
			Quantity: "ss",
		})
		assert.Nil(t, mock.ExpectationsWereMet())
		assert.Error(t, res.Error)
	})

	t.Run("add new item to empty cart successfully", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+)").WillReturnError(gorm.ErrRecordNotFound)

		// insert cart
		cartID := int64(1)
		mock.ExpectBegin()
		mock.ExpectExec("INSERT (.+)").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sessionID, entity.CartOpen).WillReturnResult(sqlmock.NewResult(cartID, 1))
		mock.ExpectCommit()

		// insert item to cart
		product := "purse"
		quantity := 2

		mock.ExpectBegin()
		mock.ExpectExec("INSERT (.+)").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), cartID, product, quantity, float64(quantity)*itemPriceMapping[product]).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		res := calc.AddItemToCart(context.TODO(), sessionID, CartItem{
			Product:  product,
			Quantity: fmt.Sprint(quantity),
		})

		assert.Nil(t, mock.ExpectationsWereMet())
		assert.NoError(t, res.Error)
	})

	t.Run("add new item to existing cart successfully", func(t *testing.T) {
		cartID := 1
		cart := sqlmock.NewRows([]string{"id", "session_id", "status"}).
			AddRow(cartID, sessionID, entity.CartOpen)
		mock.ExpectQuery("SELECT (.+)").WillReturnRows(cart)

		// check item in cart
		mock.ExpectQuery("SELECT (.+)").WillReturnError(gorm.ErrRecordNotFound)
		// insert item to cart
		product := "purse"
		quantity := 2

		mock.ExpectBegin()
		mock.ExpectExec("INSERT (.+)").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), cartID, product, quantity, float64(quantity)*itemPriceMapping[product]).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		res := calc.AddItemToCart(context.TODO(), sessionID, CartItem{
			Product:  product,
			Quantity: fmt.Sprint(quantity),
		})

		assert.Nil(t, mock.ExpectationsWereMet())
		assert.NoError(t, res.Error)
	})

	t.Run("update successfully an existing item", func(t *testing.T) {
		cartID := 1
		cart := sqlmock.NewRows([]string{"id", "session_id", "status"}).
			AddRow(cartID, sessionID, entity.CartOpen)
		mock.ExpectQuery("SELECT (.+)").WillReturnRows(cart)

		// check item in cart
		product := "purse"
		quantity := 1
		newQuantity := 1

		item := sqlmock.NewRows([]string{"cart_id", "product_name", "quantity", "price"}).
			AddRow(cartID, product, quantity, itemPriceMapping[product])
		mock.ExpectQuery("SELECT (.+)").WillReturnRows(item)

		// update item
		mock.ExpectBegin()
		mock.ExpectExec("INSERT (.+)").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), cartID, product, quantity+newQuantity, float64(quantity+newQuantity)*itemPriceMapping[product]).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		res := calc.AddItemToCart(context.TODO(), sessionID, CartItem{
			Product:  product,
			Quantity: fmt.Sprint(newQuantity),
		})

		assert.Nil(t, mock.ExpectationsWereMet())
		assert.NoError(t, res.Error)
	})
}

func TestCalculator_DeleteCartItem(t *testing.T) {
	gormDB, mock := NewMockDB()
	calc, _ := NewCalculator(itemPriceMapping, gormDB)
	sessionID := "123456"

	t.Run("delete successfully", func(t *testing.T) {
		// get cart
		cartID := 1
		cart := sqlmock.NewRows([]string{"id", "session_id", "status"}).
			AddRow(cartID, sessionID, entity.CartOpen)
		mock.ExpectQuery("SELECT (.+)").WillReturnRows(cart)

		// get item
		itemID := 1
		item := sqlmock.NewRows([]string{"id", "cart_id"}).
			AddRow(itemID, cartID)
		mock.ExpectQuery("SELECT (.+)").WithArgs(itemID).WillReturnRows(item)

		mock.ExpectBegin()
		mock.ExpectExec("UPDATE (.+)").WillReturnResult(sqlmock.NewResult(int64(itemID), 1))
		mock.ExpectCommit()

		res := calc.DeleteCartItem(context.Background(), sessionID, strconv.Itoa(itemID))
		assert.Nil(t, mock.ExpectationsWereMet())
		assert.NoError(t, res.Error)
	})

	t.Run("cart already closed", func(t *testing.T) {
		// get cart
		cartID := 1
		cart := sqlmock.NewRows([]string{"id", "session_id", "status"}).
			AddRow(cartID, sessionID, entity.CartClosed)
		mock.ExpectQuery("SELECT (.+)").WillReturnRows(cart)

		res := calc.DeleteCartItem(context.Background(), sessionID, "1")
		assert.Nil(t, mock.ExpectationsWereMet())
		assert.NoError(t, res.Error)
	})

	t.Run("cart not found", func(t *testing.T) {
		// get cart
		mock.ExpectQuery("SELECT (.+)").WillReturnError(gorm.ErrRecordNotFound)

		res := calc.DeleteCartItem(context.Background(), sessionID, "1")
		assert.Nil(t, mock.ExpectationsWereMet())
		assert.NoError(t, res.Error)
	})

	t.Run("item not found", func(t *testing.T) {
		// get cart
		cartID := 1
		cart := sqlmock.NewRows([]string{"id", "session_id", "status"}).
			AddRow(cartID, sessionID, entity.CartOpen)
		mock.ExpectQuery("SELECT (.+)").WillReturnRows(cart)

		// get item
		itemID := 1
		mock.ExpectQuery("SELECT (.+)").WithArgs(itemID).WillReturnError(gorm.ErrRecordNotFound)

		res := calc.DeleteCartItem(context.Background(), sessionID, strconv.Itoa(itemID))
		assert.Nil(t, mock.ExpectationsWereMet())
		assert.NoError(t, res.Error)
	})
}

func NewMockDB() (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		log.Fatalf("An error '%s' was not expected when opening gorm database", err)
	}

	return gormDB, mock
}
