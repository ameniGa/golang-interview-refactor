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
	repo := NewRepository(dbMock)
	t.Run("nil prices map", func(t *testing.T) {
		_, err := NewCartService(nil, repo)
		assert.Error(t, err)
	})
	t.Run("nil repo", func(t *testing.T) {
		assert.Panics(t, func() {
			_, _ = NewCartService(map[string]float64{"bag": 10}, nil)
		})
	})
	t.Run("valid ", func(t *testing.T) {
		h, _ := NewCartService(map[string]float64{"bag": 10}, repo)
		assert.NotNil(t, h)
	})
}

func TestCalculator_AddItemToCart(t *testing.T) {
	gormDB, mock := NewMockDB()
	repo := NewRepository(gormDB)

	calc, _ := NewCartService(itemPriceMapping, repo)
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

	t.Run("add new item to empty cartService successfully", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+)").WillReturnError(gorm.ErrRecordNotFound)

		// insert cartService
		cartID := int64(1)
		mock.ExpectBegin()
		mock.ExpectExec("INSERT (.+)").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sessionID, entity.CartOpen).WillReturnResult(sqlmock.NewResult(cartID, 1))
		mock.ExpectCommit()

		// insert item to cartService
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

	t.Run("add new item to existing cartService successfully", func(t *testing.T) {
		cartID := 1
		cart := sqlmock.NewRows([]string{"id", "session_id", "status"}).
			AddRow(cartID, sessionID, entity.CartOpen)
		mock.ExpectQuery("SELECT (.+)").WillReturnRows(cart)

		// check item in cartService
		mock.ExpectQuery("SELECT (.+)").WillReturnError(gorm.ErrRecordNotFound)
		// insert item to cartService
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

		// check item in cartService
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
	repo := NewRepository(gormDB)
	calc, _ := NewCartService(itemPriceMapping, repo)
	sessionID := "123456"

	t.Run("delete successfully", func(t *testing.T) {
		// get cartService
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

	t.Run("cartService already closed", func(t *testing.T) {
		// get cartService
		cartID := 1
		cart := sqlmock.NewRows([]string{"id", "session_id", "status"}).
			AddRow(cartID, sessionID, entity.CartClosed)
		mock.ExpectQuery("SELECT (.+)").WillReturnRows(cart)

		res := calc.DeleteCartItem(context.Background(), sessionID, "1")
		assert.Nil(t, mock.ExpectationsWereMet())
		assert.NoError(t, res.Error)
	})

	t.Run("cartService not found", func(t *testing.T) {
		// get cartService
		mock.ExpectQuery("SELECT (.+)").WillReturnError(gorm.ErrRecordNotFound)

		res := calc.DeleteCartItem(context.Background(), sessionID, "1")
		assert.Nil(t, mock.ExpectationsWereMet())
		assert.NoError(t, res.Error)
	})

	t.Run("item not found", func(t *testing.T) {
		// get cartService
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

func TestCalculator_GetCartData(t *testing.T) {
	gormDB, mock := NewMockDB()
	repo := NewRepository(gormDB)
	calc, _ := NewCartService(itemPriceMapping, repo)
	sessionID := "123456"

	t.Run("get cartService successfully", func(t *testing.T) {
		// get cartService
		cartID := 1
		cart := sqlmock.NewRows([]string{"id", "session_id"}).
			AddRow(cartID, sessionID)
		mock.ExpectQuery("SELECT (.+)").WillReturnRows(cart)

		// get items
		items := sqlmock.NewRows([]string{"id", "cart_id"}).
			AddRow(1, cartID).AddRow(2, cartID)
		mock.ExpectQuery("SELECT (.+)").WillReturnRows(items)

		res := calc.GetCartData(context.TODO(), sessionID)
		assert.Nil(t, mock.ExpectationsWereMet())
		assert.NotNil(t, res.Data)
	})

	t.Run("empty sessionID", func(t *testing.T) {
		res := calc.GetCartData(context.TODO(), "")
		assert.Nil(t, res.Data)
	})

	t.Run("failed to get cartService", func(t *testing.T) {
		// get cartService
		mock.ExpectQuery("SELECT (.+)").WillReturnError(gorm.ErrNotImplemented)

		res := calc.GetCartData(context.TODO(), sessionID)
		assert.Nil(t, mock.ExpectationsWereMet())
		assert.Nil(t, res.Data)
	})

	t.Run("non existing cartService", func(t *testing.T) {
		// get cartService
		mock.ExpectQuery("SELECT (.+)").WillReturnError(gorm.ErrRecordNotFound)

		res := calc.GetCartData(context.TODO(), sessionID)
		assert.Nil(t, mock.ExpectationsWereMet())
		assert.Nil(t, res.Data)
	})

	t.Run("failed to get cartService items", func(t *testing.T) {
		// get cartService
		cartID := 1
		cart := sqlmock.NewRows([]string{"id", "session_id"}).
			AddRow(cartID, sessionID)
		mock.ExpectQuery("SELECT (.+)").WillReturnRows(cart)

		// get items
		mock.ExpectQuery("SELECT (.+)").WillReturnError(gorm.ErrInvalidValueOfLength)

		res := calc.GetCartData(context.TODO(), sessionID)
		assert.Nil(t, mock.ExpectationsWereMet())
		assert.Nil(t, res.Data)
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
