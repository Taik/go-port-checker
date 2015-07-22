package checker

import (
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
)


type StatusEntry struct {
	IsOnline bool
	Error 	error
}


// GetAddrStatus returns a boolean representing the state of the address.
func GetAddrStatus(address string) (bool, error) {
	logger := log.WithFields(log.Fields{
		"address": address,
	})

	timeout := 200 * time.Millisecond
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		logger.Debug("Could not connect to server")
		return false, err
	}
	defer conn.Close()
	return true, nil
}
