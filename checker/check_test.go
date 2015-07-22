package checker
import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
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
