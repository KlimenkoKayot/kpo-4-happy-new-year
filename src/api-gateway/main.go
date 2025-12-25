package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

func main() {
	ordersURL := getEnv("ORDERS_SERVICE_URL", "http://localhost:5001")
	paymentsURL := getEnv("PAYMENTS_SERVICE_URL", "http://localhost:5002")
	port := getEnv("SERVER_PORT", "8080")

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, X-User-Id")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.Any("/api/orders", proxyHandler(ordersURL))
	r.Any("/api/orders/*path", proxyHandler(ordersURL))
	r.Any("/api/accounts", proxyHandler(paymentsURL))
	r.Any("/api/accounts/*path", proxyHandler(paymentsURL))

	log.Printf("API Gateway started on port %s", port)
	log.Printf("Orders Service: %s", ordersURL)
	log.Printf("Payments Service: %s", paymentsURL)

	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func proxyHandler(targetURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		url := targetURL + path
		if query != "" {
			url += "?" + query
		}

		req, err := http.NewRequest(c.Request.Method, url, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Service unavailable"})
			return
		}
		defer resp.Body.Close()

		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		c.Status(resp.StatusCode)
		io.Copy(c.Writer, resp.Body)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
