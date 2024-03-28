package main

import (
	"github.com/gin-gonic/gin"
	"interview/pkg/calculator"
	"interview/pkg/controllers"
	"interview/pkg/db"
	"net/http"
)

func main() {
	db.MigrateDatabase()

	ginEngine := gin.Default()
	itemPriceMapping := map[string]float64{
		"shoe":  100,
		"purse": 200,
		"bag":   300,
		"watch": 300,
	}
	taxController := controllers.NewTaxController(calculator.NewCalculator(itemPriceMapping, db.GetDatabase()))
	ginEngine.GET("/", taxController.ShowAddItemForm)
	ginEngine.POST("/add-item", taxController.AddItem)
	ginEngine.GET("/remove-cart-item", taxController.DeleteCartItem)
	srv := &http.Server{
		Addr:    ":8088",
		Handler: ginEngine,
	}

	srv.ListenAndServe()
}
