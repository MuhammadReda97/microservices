package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type SearchService struct {
	ESClient *elasticsearch.Client
}

func Service() *SearchService {
	logrus.Info("host" + os.Getenv("ELASTIC_HOST"))
	logrus.Info("user" + os.Getenv("ELASTIC_USER"))
	logrus.Info("password" + os.Getenv("ELASTIC_PASSWORD"))
	cfg := elasticsearch.Config{
		Addresses: []string{
			os.Getenv("ELASTIC_HOST"),
		},
		Username: os.Getenv("ELASTIC_USER"),
		Password: os.Getenv("ELASTIC_PASSWORD"),
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		logrus.Fatal("Failed to initialize Elasticsearch: ", err)
	}

	return &SearchService{
		ESClient: es,
	}
}

func (s *SearchService) SearchProducts(c *fiber.Ctx) error {
	logrus.Info("SearchProducts called")
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

func (s *SearchService) UpdateProductIndex(c *fiber.Ctx) error {
	id := c.Params("id")
	var updateData map[string]interface{}
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	updateDoc := map[string]interface{}{"doc": updateData}
	data, err := json.Marshal(updateDoc)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error"})
	}

	res, err := s.ESClient.Update("products", id, bytes.NewReader(data))
	if err != nil {
		logrus.WithError(err).Error("Error updating document")
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error"})
	}
	defer res.Body.Close()

	return c.JSON(fiber.Map{"message": "Product index updated successfully"})
}

func (s *SearchService) DeleteProductIndex(c *fiber.Ctx) error {
	id := c.Params("id")
	res, err := s.ESClient.Delete("products", id)
	if err != nil {
		logrus.WithError(err).Error("Error deleting document")
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error"})
	}
	defer res.Body.Close()

	return c.JSON(fiber.Map{"message": "Product index deleted successfully"})
}
