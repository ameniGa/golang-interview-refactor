//go:build integration
// +build integration

package calculator_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"interview/pkg/calculator"
	"interview/pkg/config"
	"interview/pkg/db"
	"interview/pkg/entity"
	"math/rand"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestRepository_AddCart(t *testing.T) {
	in := entity.CartEntity{
		SessionID: "1234556",
		Status:    entity.CartOpen,
	}
	repo := calculator.NewRepository(getDatabase(t))

	t.Run("add cartService successfully", func(t *testing.T) {
		cleanUp, err := addCart(t, repo, &in)
		defer cleanUp()

		assert.NoError(t, err)
		assert.NotEmpty(t, in.ID)
	})

	t.Run("missing sessionID", func(t *testing.T) {
		in = entity.CartEntity{}
		err := repo.AddCart(context.TODO(), &in)
		assert.Error(t, err)
		assert.Empty(t, in.ID)
	})

	t.Run("nil input", func(t *testing.T) {
		err := repo.AddCart(context.TODO(), nil)
		assert.Error(t, err)
	})
}

func TestRepository_GetCart(t *testing.T) {
	repo := calculator.NewRepository(getDatabase(t))

	in := entity.CartEntity{
		SessionID: "123456",
		Status:    entity.CartOpen,
	}
	cleanup, err := addCart(t, repo, &in)
	require.NoError(t, err)
	defer cleanup()

	t.Run("get existing cartService successfully", func(t *testing.T) {
		cart, err := repo.GetCart(context.TODO(), in.SessionID)
		assert.NoError(t, err)
		assert.Equal(t, in.ID, cart.ID)
	})

	t.Run("non existing cartService", func(t *testing.T) {
		cart, err := repo.GetCart(context.TODO(), "_")
		assert.IsType(t, gorm.ErrRecordNotFound, err)
		assert.Empty(t, cart.ID)
	})

	t.Run("db error", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(context.TODO(), time.Now())
		defer cancel()
		cart, err := repo.GetCart(ctx, in.SessionID)
		assert.Error(t, err)
		assert.Empty(t, cart.ID)
	})
}

func TestRepository_SaveItem(t *testing.T) {
	repo := calculator.NewRepository(getDatabase(t))

	in := entity.CartEntity{
		SessionID: "123456",
		Status:    entity.CartOpen,
	}
	cleanup, err := addCart(t, repo, &in)
	require.NoError(t, err)
	defer cleanup()

	t.Run("add item successfully", func(t *testing.T) {
		item := entity.CartItem{
			CartID:      in.ID,
			ProductName: "purse",
			Quantity:    1,
			Price:       10,
		}
		cleanUp, err := addItem(t, repo, &item)
		defer cleanUp()
		assert.NoError(t, err)
		assert.NotEmpty(t, item.ID)
	})

	t.Run("update item successfully", func(t *testing.T) {
		item := entity.CartItem{
			CartID:      in.ID,
			ProductName: "bag",
			Quantity:    1,
			Price:       10,
		}
		cleanUp, err := addItem(t, repo, &item)
		defer cleanUp()

		updated := item
		updated.Quantity = 3
		updated.Price = 30
		err = repo.SaveItem(context.TODO(), &updated)
		assert.NoError(t, err)
		assert.Equal(t, item.ID, updated.ID)
	})

	t.Run("db error", func(t *testing.T) {
		item := entity.CartItem{
			CartID:      in.ID,
			ProductName: "purse",
			Quantity:    1,
			Price:       10,
		}
		ctx, cancel := context.WithDeadline(context.TODO(), time.Now())
		defer cancel()
		err := repo.SaveItem(ctx, &item)
		assert.Error(t, err)
	})
}

func TestRepository_GetItem(t *testing.T) {
	repo := calculator.NewRepository(getDatabase(t))
	in := entity.CartEntity{
		SessionID: "123456",
		Status:    entity.CartOpen,
	}
	cleanUp, err := addCart(t, repo, &in)
	require.NoError(t, err)
	defer cleanUp()

	item := entity.CartItem{
		CartID:      in.ID,
		ProductName: "purse",
		Quantity:    1,
		Price:       10,
	}

	cleanUp, err = addItem(t, repo, &item)
	require.NoError(t, err)
	defer cleanUp()

	t.Run("get item successfully", func(t *testing.T) {
		res, err := repo.GetItem(context.TODO(), in.ID, "purse")
		assert.NoError(t, err)
		assert.Equal(t, item.ID, res.ID)
	})

	t.Run("item not found", func(t *testing.T) {
		_, err := repo.GetItem(context.TODO(), 0, "some-prod")
		assert.Error(t, err)
		assert.IsType(t, gorm.ErrRecordNotFound, err)
	})
}

func TestRepository_GetItemByID(t *testing.T) {
	repo := calculator.NewRepository(getDatabase(t))
	in := entity.CartEntity{
		SessionID: "123456",
		Status:    entity.CartOpen,
	}
	cleanUp, err := addCart(t, repo, &in)
	require.NoError(t, err)
	defer cleanUp()

	item := entity.CartItem{
		CartID:      in.ID,
		ProductName: "purse",
		Quantity:    1,
		Price:       10,
	}

	cleanUp, err = addItem(t, repo, &item)
	require.NoError(t, err)
	defer cleanUp()

	t.Run("get item successfully", func(t *testing.T) {
		res, err := repo.GetItemByID(context.TODO(), item.ID)
		assert.NoError(t, err)
		assert.Equal(t, item.ID, res.ID)
	})

	t.Run("item not found", func(t *testing.T) {
		_, err := repo.GetItemByID(context.TODO(), 0)
		assert.Error(t, err)
		assert.IsType(t, gorm.ErrRecordNotFound, err)
	})
}

func TestRepository_GetItems(t *testing.T) {
	repo := calculator.NewRepository(getDatabase(t))
	cart1 := entity.CartEntity{
		SessionID: "123456",
		Status:    entity.CartOpen,
	}
	cleanUp, err := addCart(t, repo, &cart1)
	require.NoError(t, err)
	defer cleanUp()

	cart2 := entity.CartEntity{
		SessionID: "7896",
		Status:    entity.CartOpen,
	}

	cleanUp, err = addCart(t, repo, &cart2)
	require.NoError(t, err)
	defer cleanUp()

	item1 := entity.CartItem{
		CartID:      cart1.ID,
		ProductName: "purse",
		Quantity:    1,
		Price:       10,
	}
	cleanUp, err = addItem(t, repo, &item1)

	item2 := entity.CartItem{
		CartID:      cart2.ID,
		ProductName: "purse",
		Quantity:    1,
		Price:       10,
	}
	cleanUp, err = addItem(t, repo, &item2)
	require.NoError(t, err)
	defer cleanUp()

	t.Run("get items successfully", func(t *testing.T) {
		res, err := repo.GetItems(context.TODO(), cart1.ID)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(res))
		assert.Equal(t, item1.ID, res[0].ID)
	})

	t.Run("non exiting cartService", func(t *testing.T) {
		res, err := repo.GetItems(context.TODO(), 0)
		assert.NoError(t, err)
		assert.Empty(t, res)
	})
}

func TestRepository_DeleteItem(t *testing.T) {
	repo := calculator.NewRepository(getDatabase(t))

	in := entity.CartEntity{
		SessionID: "123456",
		Status:    entity.CartOpen,
	}
	cleanUp, err := addCart(t, repo, &in)
	require.NoError(t, err)
	defer cleanUp()

	item := entity.CartItem{
		CartID:      in.ID,
		ProductName: "purse",
		Quantity:    1,
		Price:       10,
	}
	cleanUp, err = addItem(t, repo, &item)
	require.NoError(t, err)
	defer cleanUp()

	t.Run("delete item successfully", func(t *testing.T) {
		err := repo.DeleteItem(context.TODO(), item)
		assert.NoError(t, err)
	})

	t.Run("item not found", func(t *testing.T) {
		err := repo.DeleteItem(context.TODO(), entity.CartItem{
			Model: gorm.Model{ID: uint(rand.Int())},
		})
		assert.NoError(t, err)
	})
}

func addCart(t *testing.T, repo calculator.Repository, in *entity.CartEntity) (func(), error) {
	err := repo.AddCart(context.TODO(), in)

	return func() {
		err := repo.DeleteCart(context.TODO(), *in)
		require.NoError(t, err)
	}, err
}

func addItem(t *testing.T, repo calculator.Repository, in *entity.CartItem) (func(), error) {
	err := repo.SaveItem(context.TODO(), in)

	return func() {
		err := repo.DeleteItem(context.TODO(), *in)
		require.NoError(t, err)
	}, err
}

func getDatabase(t *testing.T) *gorm.DB {
	_, filename, _, _ := runtime.Caller(0)
	cfg, err := config.LoadConfig(filepath.Join(filepath.Dir(filename), "../../cmd/web-api/config/config.test.yml"))
	require.NoError(t, err)

	dbConn := db.Connect(&cfg.Database)

	return dbConn
}
