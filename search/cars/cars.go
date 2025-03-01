package cars

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

type Car struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Brand    string `json:"brand"`
	Model    string `json:"model"`
	Year     int    `json:"year"`
	Color    string `json:"color"`
	MaxSpeed int    `json:"max_speed"`
	TireSize int    `json:"tire_size"`
	Weight   int    `json:"weight"`
	BodyType string `json:"body_type"`
}

var db *gorm.DB

func Test() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)
	fmt.Println(string(dsn))

}
func initDB() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	db.AutoMigrate(&Car{})
}

func createCar(c *fiber.Ctx) error {
	var car Car
	if err := c.BodyParser(&car); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request payload"})
	}
	db.Create(&car)
	return c.Status(201).JSON(car)
}

func getCars(c *fiber.Ctx) error {
	var cars []Car
	db.Find(&cars)
	return c.JSON(cars)
}

func getCar(c *fiber.Ctx) error {
	id := c.Params("id")
	var car Car
	if err := db.First(&car, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Car not found"})
	}
	return c.JSON(car)
}

func updateCar(c *fiber.Ctx) error {
	id := c.Params("id")
	var car Car
	if err := db.First(&car, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Car not found"})
	}

	if err := c.BodyParser(&car); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request payload"})
	}
	db.Save(&car)
	return c.JSON(car)
}

func deleteCar(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := db.Delete(&Car{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete car"})
	}
	return c.SendStatus(204)
}

func main() {
	initDB()
	app := fiber.New()

	app.Post("/cars", createCar)
	app.Get("/cars", getCars)
	app.Get("/cars/:id", getCar)
	app.Put("/cars/:id", updateCar)
	app.Delete("/cars/:id", deleteCar)

	log.Fatal(app.Listen(":3000"))
}
