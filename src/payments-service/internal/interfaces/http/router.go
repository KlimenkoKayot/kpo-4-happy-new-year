package http

import (
	"github.com/gin-gonic/gin"
)

func NewRouter(handlers *Handlers) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		accounts := api.Group("/accounts")
		{
			accounts.POST("", handlers.CreateAccount)
			accounts.POST("/deposit", handlers.Deposit)
			accounts.GET("/balance", handlers.GetBalance)
		}
	}

	return r
}
