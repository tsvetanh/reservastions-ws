package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORSandCSP() gin.HandlerFunc {
	return func(c *gin.Context) {
		//c.Header("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
		origin := c.Request.Header.Get("Origin")

		// Allow only trusted origins
		if origin == "http://localhost:8080" {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Credentials, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}
