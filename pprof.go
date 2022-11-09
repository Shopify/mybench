package mybench

import (
	"fmt"
	"net/http"

	_ "net/http/pprof"

	"github.com/sirupsen/logrus"
)

func init() {
	go func() {
		addr := "localhost:6060"
		logrus.Infof("starting pprof server at %s", addr)
		fmt.Println(http.ListenAndServe(addr, nil))
	}()
}
