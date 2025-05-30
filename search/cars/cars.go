package cars

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Car struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Brand     string `json:"brand"`
	Model     string `json:"model"`
	ModelYear int    `json:"year"`
	Color     string `json:"color"`
	MaxSpeed  int    `json:"max_speed"`
	TireSize  string `json:"tire_size"`
	Weight    int    `json:"weight"`
	Body      string `json:"body"`
	Price     int    `json:"price"`
}

var db *gorm.DB

// Init initializes the cars package and database connection
func Init() error {
	if err := initDB(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	return nil
}

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

func initDB() error {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
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
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	return nil
}

// PaginatedResponse represents a paginated response structure
type PaginatedResponse struct {
	Data       []Car `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

func CreateCar(c *fiber.Ctx) error {
	var car Car
	if err := c.BodyParser(&car); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request payload"})
	}
	db.Create(&car)
	return c.Status(201).JSON(car)
}

func GetCars(c *fiber.Ctx) error {
	// Get pagination parameters from query
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	// Ensure page and limit are valid
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Maximum limit to prevent excessive data fetching
	}

	var cars []Car
	var total int64

	// Count total records
	if err := db.Model(&Car{}).Count(&total).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count cars"})
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get paginated data
	if err := db.Offset(offset).Limit(limit).Find(&cars).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch cars"})
	}

	// Calculate total pages
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	// Return paginated response
	return c.JSON(PaginatedResponse{
		Data:       cars,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

func GetCar(c *fiber.Ctx) error {
	id := c.Params("id")
	var car Car
	if err := db.First(&car, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Car not found"})
	}
	return c.JSON(car)
}

func UpdateCar(c *fiber.Ctx) error {
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

func DeleteCar(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := db.Delete(&Car{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete car"})
	}
	return c.SendStatus(204)
}

func main() {
	app := fiber.New()
	Init()
	log.Println("db initialized")

	app.Get("/cars", GetCars)
	app.Post("/cars", CreateCar)
	app.Get("/cars/:id", GetCar)
	app.Put("/cars/:id", UpdateCar)
	app.Delete("/cars/:id", DeleteCar)

	log.Fatal(app.Listen(":3000"))
}
