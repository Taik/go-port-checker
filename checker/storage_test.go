package checker

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"stablelib.com/v1/database/bolt"
)


func TestStorageSpec(t *testing.T) {
	Convey("Given database path and name", t, func() {
		file, err := ioutil.TempFile("", "bolt-")
		if err != nil {
			fmt.Errorf("Could not generate a new tempfile")
		}
		bucketName := "testBucket"

		Convey("When NewStorage is called", func() {
			s := NewStorage(file.Name(), bucketName)

			Convey("The bucket name should be set", func() {
				So(string(s.BucketName), ShouldEqual, bucketName)
			})
			Convey("The database path should be set", func() {
				So(s.DBPath, ShouldEqual, file.Name())
			})
		})

		Convey("When Init is called", func() {
			s := NewStorage(file.Name(), bucketName)
			s.Init()

			Convey("The underlying DB should be a BoltDB instance", func() {
				So(s.DB, ShouldHaveSameTypeAs, &bolt.DB{})
			})
			Convey("The bucket should be created", func() {
				s.View(func (tx *bolt.Tx) error {
					bucket := tx.Bucket([]byte(bucketName))
					So(bucket, ShouldNotBeNil)
					return nil
				})
			})
		})

		Convey("When GetBytes() is called", func() {
			s := NewStorage(file.Name(), bucketName)
			s.Init()
			key := []byte("some-existent-key")
			value := []byte("some-existent-value")
			s.Update(func (tx *bolt.Tx) error {
				bucket := tx.Bucket(s.BucketName)
				return bucket.Put(key, value)
			})

			Convey("The value should be nil for non-existent-keys", func() {
				nonexistentKey := []byte("some-non-existent-key")
				So(s.GetBytes(nonexistentKey), ShouldBeNil)
			})

			Convey("The value should exist for some existent-key", func() {
				// Workaround for ShouldEqual assertion not being able to assert on byte slices
				So(string(s.GetBytes(key)), ShouldEqual, string(value))
			})
		})

		Convey("When GetString() is called", func() {
			s := NewStorage(file.Name(), bucketName)
			s.Init()
			key := "some-existent-key"
			value := "some-existent-value"
			s.Update(func (tx *bolt.Tx) error {
				bucket := tx.Bucket(s.BucketName)
				return bucket.Put([]byte(key), []byte(value))
			})

			Convey("The value should be nil for non-existent keys", func() {
				So(s.GetString("some-non-existent-key"), ShouldBeEmpty)
			})
			Convey("The value should exist for some-existent-key", func() {
				So(s.GetString(key), ShouldEqual, value)
			})
		})

		Convey("When PutBytes() is called", func() {
			s := NewStorage(file.Name(), bucketName)
			s.Init()
			key := []byte{1, 2, 3}
			value := []byte{4, 5, 6}
			s.PutBytes(key, value)

			Convey("The value should exist", func() {
				s.View(func(tx *bolt.Tx) error {
					bucket := tx.Bucket([]byte(bucketName))
					result := bucket.Get(key)
					// Workaround for ShouldEqual assertion not being able to assert on byte slices
					So(string(result), ShouldEqual, string(value))
					return nil
				})
			})
		})

		Convey("When PutString() is called", func() {
			s := NewStorage(file.Name(), bucketName)
			s.Init()
			key := "some-existent-key"
			value := "some-existent-value"
			s.PutString(key, value)

			Convey("The value should exist", func() {
				s.View(func(tx *bolt.Tx) error {
					bucket := tx.Bucket([]byte(bucketName))
					result := bucket.Get([]byte(key))
					So(string(result), ShouldEqual, value)
					return nil
				})
			})
		})


		Reset(func() {
			os.Remove(file.Name())
		})
	})
}
