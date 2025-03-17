package main

import (
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
)

// KeyStore holds the OpenAI API key in memory
type KeyStore struct {
	mu     sync.RWMutex
	APIKey string
}

// KeyRequest represents the API key request/response structure
type KeyRequest struct {
	APIKey string `json:"api_key" binding:"required"`
}

// KeyResponse represents the API key response structure with a disclaimer
type KeyResponse struct {
	APIKey     string `json:"api_key"`
	Disclaimer string `json:"disclaimer"`
}

// NewKeyStore creates a new in-memory key store
func NewKeyStore() *KeyStore {
	return &KeyStore{}
}

// Get returns the stored API key
func (ks *KeyStore) Get() string {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	return ks.APIKey
}

// Set updates the stored API key
func (ks *KeyStore) Set(key string) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	ks.APIKey = key
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create router and keystore
	r := gin.Default()
	keyStore := NewKeyStore()

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// API key endpoints
	r.GET("/key", func(c *gin.Context) {
		key := keyStore.Get()
		if key == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API key not set"})
			return
		}

		c.JSON(http.StatusOK, KeyResponse{
			APIKey:     key,
			Disclaimer: "This key is only for demo purposes",
		})
	})

	r.POST("/key", func(c *gin.Context) {
		var req KeyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if req.APIKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "API key cannot be empty"})
			return
		}

		keyStore.Set(req.APIKey)
		c.JSON(http.StatusOK, gin.H{
			"message":    "API key updated successfully",
			"disclaimer": "This key is only for demo purposes",
		})
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
