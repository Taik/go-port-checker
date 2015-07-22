package main

import (
	"github.com/gin-gonic/gin"
)


func main() {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "OK",
		})
	})

	router.Run(":8008")
}
