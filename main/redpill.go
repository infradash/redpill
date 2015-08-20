package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/infradash/redpill/pkg/conf"
	"github.com/infradash/redpill/pkg/domain"
	"github.com/infradash/redpill/pkg/env"
	"github.com/infradash/redpill/pkg/event"
	"github.com/infradash/redpill/pkg/mock"
	"github.com/infradash/redpill/pkg/orchestrate"
	"github.com/infradash/redpill/pkg/redpill"
	"github.com/infradash/redpill/pkg/registry"
	"github.com/qorio/maestro/pkg/zk"
	"github.com/qorio/omni/auth"
	"github.com/qorio/omni/rest"
	"github.com/qorio/omni/runtime"
	"github.com/qorio/omni/version"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	EnvPort    = "REDPILL_PORT"
	EnvZkHosts = "REDPILL_ZK_HOSTS"
)

var (
	currentWorkingDir, _ = os.Getwd()

	port       = flag.Int("port", runtime.EnvInt(EnvPort, 5050), "Server listening port")
	zk_hosts   = flag.String("zk_hosts", runtime.EnvString(EnvZkHosts, "localhost:2181"), "ZK hosts")
	zk_timeout = flag.String("zk_timeout", "5s", "Zk timeout")
)

func must_not(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

	buildInfo := version.BuildInfo()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", buildInfo.Notice())
		fmt.Fprintf(os.Stderr, "flags:\n")
		flag.PrintDefaults()
	}

	glog.Infoln(buildInfo.Notice())

	flag.Parse()

	timeout, err := time.ParseDuration(*zk_timeout)
	must_not(err)

	zk_pool := func() zk.ZK {
		glog.Infoln("Connecting to zookeeper:", *zk_hosts)
		zc, err := zk.Connect(strings.Split(*zk_hosts, ","), timeout)
		must_not(err)
		return zc
	}

	redpillOptions := redpill.Options{
		WorkingDir: currentWorkingDir,
	}

	authService := auth.Init(auth.Settings{
		IsAuthOn:      func() bool { return false },
		AuthIntercept: mock.AuthContext,
		ErrorRenderer: rest.ErrorRenderer,
	})

	service_registry := registry.NewService(zk_pool)
	service_domain := domain.NewService(zk_pool)
	service_env := env.NewService(zk_pool, service_domain)
	service_event := event.NewService(mock.GetEventFeed)
	service_confs := conf.NewService(zk_pool, mock.ConfStorage, service_domain)
	service_orchestrate := orchestrate.NewService(zk_pool,
		mock.OrchestrationModelStorage, mock.OrchestrationInstanceStorage)

	endpoint, err := redpill.NewApi(
		redpillOptions,
		authService,
		service_env,
		service_domain,
		service_event,
		service_registry,
		service_orchestrate,
		service_confs,
	)

	if err != nil {
		panic(err)
	}

	// Mock
	endpoint.CreateServiceContext = mock.ServiceContext(endpoint.GetEngine())

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
