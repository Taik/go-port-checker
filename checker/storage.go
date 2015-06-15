package checker

import (
	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
)


type Storage struct {
	*bolt.DB
	DBPath     string
	BucketName []byte
}

func NewStorage(path string, bucket string) *Storage {
	s := &Storage{
		DBPath: path,
		BucketName: []byte(bucket),
	}
	return s
}

func (s *Storage) Init() error {
	logger := log.WithFields(log.Fields{
		"dbPath": s.DBPath,
		"bucketName": s.BucketName,
	})
	var err error

	s.DB, err = bolt.Open(s.DBPath, 0666, nil)
	if err != nil {
		logger.Errorf("Could not open database: %s", err.Error())
		return err
	}

	err = s.Update(func (tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(s.BucketName)
		return err
	})
	return err
}

func (s *Storage) GetBytes(key []byte) []byte {
	result := []byte{}
	s.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(s.BucketName)
		result = bucket.Get(key)
		return nil
	})
	return result
}

func (s *Storage) GetString(key string) string {
	return string(s.GetBytes([]byte(key)))
}

func (s *Storage) PutString(key string, value string) error {
	return s.PutBytes([]byte(key), []byte(value))
}

func (s *Storage) PutBytes(key []byte, value []byte) error {
	return s.Update(func (tx *bolt.Tx) error {
		bucket := tx.Bucket(s.BucketName)
		return bucket.Put(key, value)
	})
}
