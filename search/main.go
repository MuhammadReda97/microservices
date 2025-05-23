package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"search-service/cars"
	"search-service/elastic"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type SearchService struct {
	ESClient *elasticsearch.Client
}

// func main() {
// 	cars.Testing()
// 	elastic.Service()
// 	// es password is elastic,oauFcQ-WiwCakfbv*id1
// 	println("started...")
// 	// Initialize Logger
// 	logrus.SetFormatter(&logrus.JSONFormatter{})
// 	logrus.SetOutput(os.Stdout)

// 	// Initialize Elasticsearch
// 	cfg := elasticsearch.Config{
// 		Addresses: []string{
// 			"http://localhost:9200",
// 		},
// 		Username: "elastic",
// 		Password: "=AR4wCe6XEO0lPWq-fsb",
// 	}

// 	es, err := elasticsearch.NewClient(cfg)
// 	if err != nil {
// 		logrus.Fatal("Failed to initialize Elasticsearch: ", err)
// 	}

// 	service := &SearchService{
// 		ESClient: es,
// 	}

// 	// Initialize Fiber app
// 	app := fiber.New()

// 	app.Get("/search", service.SearchProducts)
// 	//app.Post("/search/index", service.IndexPartner)
// 	// app.Put("/search/index/:id", service.UpdateProductIndex)
// 	// app.Delete("/search/index/:id", service.DeleteProductIndex)

// 	// Start the server
// 	logrus.Info("Starting Search Service on port 8081")
// 	if err := app.Listen(":8081"); err != nil {
// 		logrus.Fatal("Failed to start server: ", err)
// 	}
// }

func main() {
	cars.Testing()
	// es password is elastic,oauFcQ-WiwCakfbv*id1
	println("started...")
	// Initialize Logger
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	// Initialize Fiber app
	app := fiber.New()
	service := elastic.Service()
	app.Get("/search", service.SearchProducts)
	// Start the server
	logrus.Info("Starting Search Service on port 8081")
	if err := app.Listen(":8081"); err != nil {
		logrus.Fatal("Failed to start server: ", err)
	}
}

func (s *SearchService) SearchProducts(c *fiber.Ctx) error {
	logrus.Info("here")
	var queryParams map[string]interface{}
	if err := c.BodyParser(&queryParams); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Initialize Elasticsearch query
	esQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{},
		},
	}

	// Process "must" conditions
	if mustConditions, ok := queryParams["must"].([]interface{}); ok {
		esQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = mustConditions
	}

	// Process "should" conditions
	if shouldConditions, ok := queryParams["should"].([]interface{}); ok {
		esQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["should"] = shouldConditions

		// Optional: Ensure at least one "should" condition matches
		if len(shouldConditions) > 0 {
			esQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["minimum_should_match"] = 1
		}
	}

	// Convert query to JSON
	data, err := json.Marshal(esQuery)
	if err != nil {
		logrus.WithError(err).Error("Error marshaling query")
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error"})
	}

	// Execute Elasticsearch search query
	res, err := s.ESClient.Search(
		s.ESClient.Search.WithContext(context.Background()),
		s.ESClient.Search.WithIndex("products"),
		s.ESClient.Search.WithBody(bytes.NewReader(data)),
		s.ESClient.Search.WithPretty(),
	)
	if err != nil {
		logrus.WithError(err).Error("Error searching Elasticsearch")
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error"})
	}
	defer res.Body.Close()

	// Parse Elasticsearch response
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		logrus.WithError(err).Error("Error parsing search response")
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error"})
	}

	// Format and return response
	formattedResponse := formatElasticsearchResponse(result)
	return c.JSON(formattedResponse)
}

func formatElasticsearchResponse(esResponse map[string]interface{}) []map[string]interface{} {
	var formattedResults []map[string]interface{}

	hits, ok := esResponse["hits"].(map[string]interface{})
	if !ok {
		return formattedResults
	}

	hitsArray, ok := hits["hits"].([]interface{})
	if !ok {
		return formattedResults
	}

	for _, hit := range hitsArray {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		source, ok := hitMap["_source"].(map[string]interface{})
		if !ok {
			continue
		}

		formattedHit := map[string]interface{}{
			"id":       hitMap["_id"],
			"score":    hitMap["_score"],
			"name":     source["name"],
			"price":    source["price"],
			"status":   source["status"],
			"brand":    source["brand"],
			"category": source["category"],
		}

		formattedResults = append(formattedResults, formattedHit)
	}

	return formattedResults
}
