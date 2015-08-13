package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pmylund/go-cache"
	"github.com/spf13/viper"

	"github.com/taik/go-port-checker/checker"
	"runtime"
)

// StatusResource holds shared states across
type StatusResource struct {
	Cache *cache.Cache
}

func (r *StatusResource) mainHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

func (r *StatusResource) statusCheckHandler(c *gin.Context) {
	address := strings.Trim(c.Param("address"), " ")
	if len(address) > 64 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"msg":    "Address is too long",
		})
		return
	}

	statusChan := make(chan *checker.StatusEntry)
	status := &checker.StatusEntry{}

	cachedStatus, found := r.Cache.Get(address)
	if !found {
		// When error message is set to nil returned, this means that the key was not found.
		go func() {
			status, err := checker.GetAddrStatus(address)
			r.Cache.Set(address, status, cache.DefaultExpiration)
			statusChan <- &checker.StatusEntry{IsOnline: status, Error: err}
		}()
		status = <-statusChan
	} else {
		status.IsOnline = cachedStatus.(bool)
	}

	if status.Error != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"address": address,
			"online":  status.IsOnline,
			"msg":     status.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"address": address,
		"online":  status.IsOnline,
	})
}

func (r *StatusResource) loaderIOHandler(c *gin.Context) {
	c.Writer.WriteString("loaderio-9238b813ae75e7c43e9fc839c78fa7de")
	return
}

func initConfig() *viper.Viper {
	c := viper.New()
	c.SetConfigName("portChecker")
	c.SetEnvPrefix("checker")
	c.AutomaticEnv()

	// Overrides ListenPort value with a non-prefixed environment var
	c.BindEnv("ListenPort", "PORT")

	c.SetDefault("ListenPort", "8080")
	c.SetDefault("CacheExpirationMS", 30*1000)
	c.SetDefault("CacheCleanupIntervalMS", 10*1000)
	c.SetDefault("Threads", runtime.NumCPU())

	return c
}

func main() {
	config := initConfig()

	runtime.GOMAXPROCS(config.GetInt("Threads"))
	c := cache.New(
		time.Duration(config.GetInt("CacheExpirationMS"))*time.Millisecond,
		time.Duration(config.GetInt("CacheCleanupIntervalMS"))*time.Millisecond,
	)
	statusResource := &StatusResource{Cache: c}

	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/", statusResource.mainHandler)
	router.GET("/loaderio-9238b813ae75e7c43e9fc839c78fa7de", statusResource.loaderIOHandler)
	router.GET("/status/:address", statusResource.statusCheckHandler)

	router.Run(fmt.Sprintf(":%s", config.GetString("ListenPort")))
}
