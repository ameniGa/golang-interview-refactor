package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"interview/pkg/calculator"
	"interview/pkg/config"
	"interview/pkg/controllers"
	"interview/pkg/db"
	"net/http"
)

var configFile = flag.String("config", "cmd/web-api/config/config.dev.yml", "Path of the configuration file.")

func main() {
	flag.Parse()
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		panic(err)
	}

	fmt.Println(viper.GetString("DB_PASSWORD"))
	dbConn := db.Connect(&cfg.Database)

	ginEngine := gin.Default()
	itemPriceMapping := map[string]float64{
		"shoe":  100,
		"purse": 200,
		"bag":   300,
		"watch": 300,
	}
	cal, err := calculator.NewCalculator(itemPriceMapping, calculator.NewRepository(dbConn))
	if err != nil {
		panic(err)
	}
	taxController := controllers.NewTaxController(cal)

	ginEngine.GET("/", taxController.ShowAddItemForm)
	ginEngine.POST("/add-item", taxController.AddItem)
	ginEngine.GET("/remove-cart-item", taxController.DeleteCartItem)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%v:%v", cfg.Server.Host, cfg.Server.Port),
		Handler: ginEngine,
	}

	err = srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
