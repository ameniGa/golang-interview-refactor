package calculator_test

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"interview/pkg/calculator"
	"log"
	"testing"
)

func TestNewCalculator(t *testing.T) {
	dbMock, _ := NewMockDB()
	t.Run("nil prices map", func(t *testing.T) {
		_, err := calculator.NewCalculator(nil, dbMock)
		assert.Error(t, err)
	})
	t.Run("nil db", func(t *testing.T) {
		assert.Panics(t, func() {
			_, _ = calculator.NewCalculator(map[string]float64{"bag": 10}, nil)
		})
	})
	t.Run("valid ", func(t *testing.T) {
		h, _ := calculator.NewCalculator(map[string]float64{"bag": 10}, dbMock)
		assert.NotNil(t, h)
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
