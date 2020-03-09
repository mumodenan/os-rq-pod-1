package command

import (
	"net/http"

	"github.com/cfhamlet/os-rq-pod/app/router"
	"github.com/cfhamlet/os-rq-pod/core"
	defaultConfig "github.com/cfhamlet/os-rq-pod/internal/config"
	"github.com/cfhamlet/os-rq-pod/pkg/command"
	"github.com/cfhamlet/os-rq-pod/pkg/config"
	"github.com/cfhamlet/os-rq-pod/pkg/ginserv"
	"github.com/cfhamlet/os-rq-pod/pkg/log"
	"github.com/cfhamlet/os-rq-pod/pkg/runner"
	"github.com/cfhamlet/os-rq-pod/pkg/utils"
	"github.com/gin-gonic/gin"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

func init() {
	Root.AddCommand(command.NewRunCommand("rq-pod", run))
}

var startFail chan error

func run(conf *viper.Viper) {
	loadConfig := func() (*viper.Viper, error) {
		err := config.LoadConfig(conf, defaultConfig.EnvPrefix, defaultConfig.DefaultConfig)
		return conf, err
	}

	newEngine := func(*core.Pod) *gin.Engine {
		return ginserv.NewEngine(conf)
	}
	newServer := func(engine *gin.Engine) *http.Server {
		return ginserv.NewServer(conf, engine)
	}

	app := fx.New(
		fx.Provide(
			loadConfig,
			utils.NewRedisClient,
			core.NewPod,
			newEngine,
			newServer,
			ginserv.NewAPIGroup,
			runner.HTTPServerLifecycle,
		),
		fx.Invoke(
			config.PrintDebugConfig,
			log.ConfigLogging,
			ginserv.LoadGlobalMiddlewares,
			router.InitAPIRouter,
			core.LoadQueues,
		),
		fx.Populate(&startFail),
	)

	runner.Run(app, startFail)
}
