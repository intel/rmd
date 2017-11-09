package hospitality

import (
	//"fmt"
	"testing"

	//. "github.com/prashantv/gostub"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetByRequest(t *testing.T) {
	Convey("Test get hosptility score by request", t, func() {
		h := &Hospitality{}
		req := &Request{}

		Convey("Test get hosptility score bad request min=0 ", func() {
			req.MaxCache = 1
			req.MinCache = 0
			err := h.GetByRequest(req)
			So(err, ShouldNotBeNil)
		})
		Convey("Test get hosptility score bad request max=0", func() {
			req.MaxCache = 0
			req.MinCache = 1
			err := h.GetByRequest(req)
			So(err, ShouldNotBeNil)
		})
		Convey("Test get hosptility score bad request max<min", func() {
			req.MaxCache = 1
			req.MinCache = 2
			err := h.GetByRequest(req)
			So(err, ShouldNotBeNil)
		})

	})

}
