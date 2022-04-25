package echo

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestListenAndServeWithSignal(t *testing.T) {
	Convey("TestListenAndServeWithSignal...", t, func() {
		server := &Server{
			addr:     "localhost",
			port:     8000,
			protocol: "tcp",
		}
		handler := NewEchoHandler()
		err := ListenAndServeWithSignal(server, handler)
		So(err, ShouldBeNil)
	})
}
