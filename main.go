package main

import (
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

	statusChan := make(chan *checker.StatusEntry)

	go func() {
		status, err := checker.GetAddrStatus(address)
		statusChan <- &checker.StatusEntry{IsOnline: status, Error: err}
	}()

	status := <- statusChan

	if status.Error != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"address": address,
			"online": status.IsOnline,
			"msg": status.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"address": address,
		"online": status.IsOnline,
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
