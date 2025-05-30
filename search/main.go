package main

import (
	"os"
	"search-service/cars"
	"search-service/elastic"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize Logger
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)

	// Initialize Fiber app
	app := fiber.New()

	// Initialize Elasticsearch service
	service := elastic.Service()
	app.Get("/search", service.SearchProducts)

	// Initialize Cars service
	if err := cars.Init(); err != nil {
		logrus.Fatal("Failed to initialize cars service: ", err)
	}

	// Setup cars routes
	app.Get("/cars", cars.GetCars)
	app.Post("/cars", cars.CreateCar)

	// Start the server
	logrus.Info("Starting Search Service on port 8081")
	if err := app.Listen(":8081"); err != nil {
		logrus.Fatal("Failed to start server: ", err)
	}
}
