package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"stablelib.com/v1/database/redis"

	"github.com/taik/go-port-checker/checker"
)

func newRedisPool(server string, password string) *redis.Pool {
	return &redis.Pool{
		IdleTimeout: 5 * time.Minute,
		MaxIdle:     5,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// StatusResource holds shared states across
type StatusResource struct {
	Cache *redis.Pool
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
			"msg":    "Address is too long",
		})
		return
	}

	cacheConn := r.Cache.Get()
	defer cacheConn.Close()

	statusChan := make(chan *checker.StatusEntry)
	status := &checker.StatusEntry{}

	cachedStatus, err := redis.Bool(cacheConn.Do("GET", address))
	if err != nil {
		// When error message is set to nil returned, this means that the key was not found.
		go func() {
			status, err := checker.GetAddrStatus(address)
			cacheConn.Do("SETEX", address, 120, true)
			statusChan <- &checker.StatusEntry{IsOnline: status, Error: err}
		}()
		status = <-statusChan
	} else {
		status.IsOnline = cachedStatus
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

func initConfig() *viper.Viper {
	c := viper.New()
	c.SetConfigName("portChecker")
	c.SetEnvPrefix("checker")
	c.AutomaticEnv()

	c.SetDefault("Port", "8080")
	return c
}

func main() {
	config := initConfig()

	cache := newRedisPool(config.GetString("RedisAddr"), config.GetString("RedisPass"))
	statusResource := &StatusResource{Cache: cache}
	defer statusResource.Cache.Close()

	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/", statusResource.mainHandler)
	router.GET("/status/:address", statusResource.statusCheckHandler)

	router.Run(fmt.Sprintf(":%s", config.GetString("Port")))
}
