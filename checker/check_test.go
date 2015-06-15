package checker
import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"time"
)


func TestCheckPortOpenSpec(t *testing.T) {
	Convey("Given google.com:80", t, func() {
		address := "www.google.com:80"
		result, err := GetAddrStatus(address)
		Convey("The status should be true", func() {
			So(result, ShouldBeTrue)
		})
		Convey("The error should be false", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("Given some undefined host", t, func() {
		address := "kladjsfkla.cadsfas.co:1029"
		result, err := GetAddrStatus(address)
		Convey("The status should be false", func() {
			So(result, ShouldBeFalse)
		})
		Convey("The error should not be nil", func() {
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Given a defined host but closed port", t, func() {
		address := "www.google.com:811"
		result, err := GetAddrStatus(address)
		Convey("The status should be false", func() {
			So(result, ShouldBeFalse)
		})
		Convey("The error should not be nil", func() {
			So(err, ShouldNotBeNil)
		})
	})
}


func TestGetAddrStatusFromCacheSpec(t *testing.T) {
	Convey("Given Storage instance", t, func() {
		file, err := ioutil.TempFile("", "bolt-")
		if err != nil {
			fmt.Errorf("Could not generate a new tempfile")
		}

		bucketName := "someCoolBucket"
		storage := NewStorage(file.Name(), bucketName)
		err = storage.Init()
		cacheInterval := 30 * time.Second
		So(err, ShouldBeNil)

		Convey("When the key is non-existant", func() {
			key := "some-non-existent-key"
			result, err := GetCachedAddrStatus(storage, key, cacheInterval)

			Convey("The error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("The result should be false", func() {
				So(result, ShouldBeFalse)
			})
		})

		Convey("When the key exists", func() {
			existingHost := StatusEntry{
				Address: "some-cool-site.com",
				Status: true,
				Timestamp: time.Now(),
			}
			serializedValue, err := existingHost.Serialize()
			So(err, ShouldBeNil)
			storage.PutBytes([]byte(existingHost.Address), serializedValue)

			result, err := GetCachedAddrStatus(storage, existingHost.Address, cacheInterval)
			Convey("The status should be true", func() {
				So(result, ShouldBeTrue)
			})
			Convey("The error should be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When the key exists and the timestamp expired", func() {
			existingHost := StatusEntry{
				Address: "expired-site.com",
				Status: true,
				Timestamp: time.Now().Add(-cacheInterval - 1 * time.Second),
			}
			serializedValue, err := existingHost.Serialize()
			So(err, ShouldBeNil)
			storage.PutBytes([]byte(existingHost.Address), serializedValue)

			result, err := GetCachedAddrStatus(storage, existingHost.Address, cacheInterval)
			Convey("The status should be false", func() {
				So(result, ShouldBeFalse)
			})
			Convey("The error should be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Reset(func() {
			storage.Close()
			os.Remove(file.Name())
		})
	})
}