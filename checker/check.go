package checker
import (
	"net"
	"time"
	"errors"

	log "github.com/Sirupsen/logrus"
	"encoding/json"
)


type StatusEntry struct {
	Address string
	Status bool
	Timestamp time.Time
}

func deserializeToStatusEntry(data []byte) (*StatusEntry, error) {
	entry := &StatusEntry{}
	err := entry.Deserialize(data)
	return entry, err
}

func (e *StatusEntry) Deserialize(bytes []byte) error {
	return json.Unmarshal(bytes, e)
}

func (e *StatusEntry) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

func GetCachedAddrStatus(db *Storage, address string, cachedInterval time.Duration) (bool, error) {
	logFields := log.Fields{
		"address": address,
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
	} else {
		defer conn.Close()
		return true, nil
	}
}