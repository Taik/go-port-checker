package main

import (
	"time"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/taik/go-port-checker/checker"
)


// StatusResource holds shared states across
type StatusResource struct {
	storage *checker.Storage
}


func (r *StatusResource) mainHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}


func (r *StatusResource) statusCheckHandler(c *gin.Context) {
	address := c.Param("address")
	if len(address) > 64 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"msg": "Address is too long",
		})
		return
	}

	cacheInterval := 2 * time.Second
	status, err := checker.GetCachedAddrStatus(r.storage, address, cacheInterval)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"msg": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"address": address,
		"online": status,
	})
}


func main() {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	storage := checker.NewStorage("/tmp/status.bolt", "status")
	if err := storage.Init(); err != nil {
		panic("Could not initialize storage")
	}
	statusResource := &StatusResource{storage: storage}

	router.GET("/", statusResource.mainHandler)
	router.GET("/status/:address", statusResource.statusCheckHandler)

	router.Run(":8008")
}
