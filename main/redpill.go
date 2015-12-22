package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/infradash/redpill/pkg/conf"
	"github.com/infradash/redpill/pkg/console"
	"github.com/infradash/redpill/pkg/dockerapi"
	"github.com/infradash/redpill/pkg/domain"
	"github.com/infradash/redpill/pkg/env"
	"github.com/infradash/redpill/pkg/event"
	"github.com/infradash/redpill/pkg/mock"
	"github.com/infradash/redpill/pkg/orchestrate"
	"github.com/infradash/redpill/pkg/pkg"
	"github.com/infradash/redpill/pkg/redpill"
	"github.com/infradash/redpill/pkg/registry"
	"github.com/pkg/profile"
	"github.com/qorio/maestro/pkg/zk"
	"github.com/qorio/omni/auth"
	"github.com/qorio/omni/rest"
	"github.com/qorio/omni/runtime"
	"github.com/qorio/omni/version"
	"log"
	"net/http"
	_ "net/http/pprof"
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
	profiler   = flag.String("profile", "cpu", "cpu|mem|block|nil")
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

	s3 := new(conf.S3Bucket)
	s3.BindFlags()

	flag.Parse()

	pprof := profile.Start(profile.CPUProfile, profile.NoShutdownHook)
	// switch *profiler {
	// case "cpu":
	// 	defer profile.Start(profile.CPUProfile).Stop()
	// case "mem":
	// 	defer profile.Start(profile.MemProfile).Stop()
	// case "block":
	// 	defer profile.Start(profile.BlockProfile).Stop()
	// default:
	// }

	timeout, err := time.ParseDuration(*zk_timeout)
	must_not(err)

	glog.Infoln("Connecting to zookeeper:", *zk_hosts)
	zc, err := zk.Connect(strings.Split(*zk_hosts, ","), timeout)
	must_not(err)

	zk_pool := func() zk.ZK {
		return zc
	}

	conf_storage := func() conf.ConfStorage {
		if s3.IsRequested() {
			glog.Infoln("Checking S3 conf storage...")
			err := s3.Init(zk_pool)
			must_not(err)

			glog.Infoln("Using S3 conf storage")
			return s3
		} else {
			glog.Infoln("Using local BoltDB storage")
			return mock.ConfStorage()
		}
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
	service_pkg := pkg.NewService(zk_pool, service_domain)
	service_env := env.NewService(zk_pool, service_domain)
	service_console := console.NewService(zk_pool, service_domain)
	service_dockerapi := dockerapi.NewService(zk_pool, service_domain)
	service_confs := conf.NewService(zk_pool, conf_storage, service_domain)
	service_orchestrate := orchestrate.NewService(zk_pool,
		mock.OrchestrationModelStorage, mock.OrchestrationInstanceStorage)
	service_event := event.NewService(mock.GetEventFeed)

	endpoint, err := redpill.NewApi(
		redpillOptions,
		authService,
		service_env,
		service_domain,
		service_event,
		service_registry,
		service_orchestrate,
		service_confs,
		service_pkg,
		service_console,
		service_dockerapi,
	)

	if err != nil {
		panic(err)
	}

	// Mock
	endpoint.CreateServiceContext = mock.ServiceContext(endpoint.GetEngine())

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	runtime.StandardContainer(*port,
		func() http.Handler {
			return endpoint
		},
		func() error {
			glog.Infoln("Stopping pprof")
			pprof.Stop()

			glog.Infoln("Stopped endpoint")
			return nil
		})

	glog.Infoln("Bye")
}
