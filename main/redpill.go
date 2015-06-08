package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/infradash/redpill/pkg/redpill"
	"github.com/qorio/omni/auth"
	"github.com/qorio/omni/rest"
	"github.com/qorio/omni/runtime"
	"github.com/qorio/omni/version"
	"net/http"
	"os"
)

const (
	EnvPort = "PORT"
)

var (
	currentWorkingDir, _ = os.Getwd()

	port = flag.Int("port", runtime.EnvInt(EnvPort, 5050), "Server listening port")
)

func main() {

	buildInfo := version.BuildInfo()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", buildInfo.Notice())
		fmt.Fprintf(os.Stderr, "flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	glog.Infoln(buildInfo.Notice())

	redpillOptions := redpill.Options{
		WorkingDir: currentWorkingDir,
	}

	// This service implements a number of interfaces
	// Redpill service, webhook service and oauth2 service
	redpillService, sErr := redpill.NewService(redpillOptions)
	if sErr != nil {
		panic(sErr)
	}

	authService := auth.Init(auth.Settings{
		ErrorRenderer: rest.ErrorRenderer,
		AuthIntercept: redpill.MockAuthContext,
	})

	endpoint, err := redpill.NewApi(
		redpillOptions,
		authService,
		redpillService)

	if err != nil {
		panic(err)
	}

	runtime.StandardContainer(*port,
		func() http.Handler {
			return endpoint
		},
		func() error {
			glog.Infoln("Stopped endpoint")
			return nil
		})

	glog.Infoln("Bye")
}
