package cars

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// CustomValidator is a custom validator instance
var validate *validator.Validate
var translator ut.Translator

// Car represents a car entity with validation rules
type Car struct {
	ID        uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Brand     string `json:"brand" validate:"required,min=2,max=50,alpha_space"`
	Model     string `json:"model" validate:"required,min=1,max=50"`
	ModelYear int    `json:"year" validate:"required,min=1900,max=2026"`
	Color     string `json:"color" validate:"required,min=2,max=30,alpha_space"`
	MaxSpeed  int    `json:"max_speed" validate:"required,min=1,max=500"`
	TireSize  string `json:"tire_size" validate:"required"`
	Weight    int    `json:"weight" validate:"required,min=500,max=5000"`
	Body      string `json:"body" validate:"required,oneof=sedan suv hatchback coupe convertible wagon van pickup truck"`
	Price     int    `json:"price" validate:"required,min=1"`
}

// PaginatedResponse represents a paginated response structure
type PaginatedResponse struct {
	Data       []Car `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

// ValidationError represents a validation error with field-specific messages
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

var db *gorm.DB

// Init initializes the cars package and database connection
func Init() error {
	// Initialize validator
	initValidator()

	// Initialize database
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

// initValidator initializes the custom validator with custom validation rules
func initValidator() {
	validate = validator.New()
	
	// Set up translator
	english := en.New()
	uni := ut.New(english, english)
	translator, _ = uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, translator)

	// Register custom validation for alpha with spaces
	validate.RegisterValidation("alpha_space", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		matched, _ := regexp.MatchString(`^[a-zA-Z\s]+$`, value)
		return matched
	})

	// Register custom translations
	_ = validate.RegisterTranslation("alpha_space", translator,
		func(ut ut.Translator) error {
			return ut.Add("alpha_space", "{0} must contain only letters and spaces", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("alpha_space", fe.Field())
			return t
		},
	)
}

// validateCar performs validation on the car struct and returns any validation errors
func validateCar(car *Car) []ValidationError {
	var errors []ValidationError

	if err := validate.Struct(car); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			field := strings.ToLower(err.Field())
			message := getValidationErrorMessage(err)
			errors = append(errors, ValidationError{
				Field:   field,
				Message: message,
			})
		}
	}

	return errors
}

// getValidationErrorMessage returns a user-friendly error message for validation errors
func getValidationErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", err.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", err.Field(), err.Param())
	case "alpha_space":
		return fmt.Sprintf("%s must contain only letters and spaces", err.Field())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", err.Field(), err.Param())
	default:
		return fmt.Sprintf("%s failed validation: %s", err.Field(), err.Tag())
	}
}

func CreateCar(c *fiber.Ctx) error {
	var car Car
	if err := c.BodyParser(&car); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request payload",
			"details": err.Error(),
		})
	}

	// Validate car data
	if validationErrors := validateCar(&car); len(validationErrors) > 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Validation failed",
			"details": validationErrors,
		})
	}

	// Create car record in database
	result := db.Create(&car)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create car record",
			"details": result.Error.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Car created successfully",
		"data": car,
	})
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
