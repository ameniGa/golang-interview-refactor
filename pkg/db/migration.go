package db

import (
	"gorm.io/gorm"
	"interview/pkg/entity"
)

func migrate(db *gorm.DB) {
	// AutoMigrate will create or update the tables based on the models
	err := db.AutoMigrate(&entity.CartEntity{}, &entity.CartItem{})
	if err != nil {
		panic(err)
	}
}
