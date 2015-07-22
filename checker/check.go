package checker

import (
	"errors"
	"net"
	"time"

	"encoding/json"

	log "github.com/Sirupsen/logrus"
)

// StatusEntry stores the status and expiration time of the address.
type StatusEntry struct {
	Address   string
	Status    bool
	Timestamp time.Time
}

func deserializeToStatusEntry(data []byte) (*StatusEntry, error) {
	entry := &StatusEntry{}
	err := entry.Deserialize(data)
	return entry, err
}

// Deserialize unmarshals the StatusEntry instance into a struct.
func (e *StatusEntry) Deserialize(bytes []byte) error {
	return json.Unmarshal(bytes, e)
}

// Serialize returns the byte array after serializing the StatusEntry struct.
func (e *StatusEntry) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

// GetCachedAddrStatus returns the status of the host.
// If the timestamp falls within `cachedInterval`, the cached status is returned. Otherwise,
// a live status check is queued and a StatusEntry is created.
func GetCachedAddrStatus(db *Storage, address string, cachedInterval time.Duration) (bool, error) {
	logFields := log.Fields{
		"address":        address,
		"cachedInterval": cachedInterval,
	}

	data := db.GetBytes([]byte(address))
	if data == nil {
		return false, nil
	}

	entry, err := deserializeToStatusEntry(data)
	if err != nil {
		log.WithFields(logFields).Debug("Error deserializing to StatusEntry struct")
		return false, err
	}

	expiresAt := entry.Timestamp.Add(cachedInterval)

	logFields["statusEntry"] = entry
	logFields["expiresAt"] = expiresAt

	if expiresAt.Before(time.Now()) {
		log.WithFields(logFields).Debug("Entry timestamp is past expiration")
		return false, nil
	}

	return entry.Status, nil
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
		return false, errors.New("Could not connect to server")
	}
	defer conn.Close()
	return true, nil
}
