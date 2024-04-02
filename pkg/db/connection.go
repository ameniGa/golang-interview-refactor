package db

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"interview/pkg/config"
)

func Connect(cfg *config.Database) *gorm.DB {
	// MySQL connection string
	// Update the username, password, host, port, and database name accordingly
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	// Open the connection to the database
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	migrate(db)

	return db
}
