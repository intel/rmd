package hospitality

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetByRequest(t *testing.T) {
	Convey("Test get hospitality score by request", t, func(c C) {
		h := &Hospitality{}
		req := &Request{}

		c.Convey("Test get hospitality score bad request min=0 ", func(c C) {
			req.MaxCache = 1
			req.MinCache = 0
			err := h.GetByRequest(req)
			c.So(err, ShouldNotBeNil)
		})
		c.Convey("Test get hospitality score bad request max=0", func(c C) {
			req.MaxCache = 0
			req.MinCache = 1
			err := h.GetByRequest(req)
			c.So(err, ShouldNotBeNil)
		})
		c.Convey("Test get hospitality score bad request max<min", func(c C) {
			req.MaxCache = 1
			req.MinCache = 2
			err := h.GetByRequest(req)
			c.So(err, ShouldNotBeNil)
		})

	})

}
