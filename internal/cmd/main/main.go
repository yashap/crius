package main

import (
	"github.com/gin-gonic/gin"
	"github.com/yashap/crius/internal/controller"
	"github.com/yashap/crius/internal/db"
	"github.com/yashap/crius/internal/domain/repository"
	"log"
	"os"
)

func main() {
	// Wire up dependency graph
	database := db.Connect(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"),
	)
	serviceRepository := repository.NewServiceRespository(database)
	serviceController := controller.NewServiceController(&serviceRepository)

	// Run the server
	r := gin.Default()
	r.POST("/services", serviceController.Create)
	// TODO r.GET
	// TODO r.DELETE
	err := r.Run()
	if err != nil {
		log.Fatal("Failed to run server" + err.Error())
	}
}
