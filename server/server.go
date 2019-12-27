package server

import (
	"github.com/gin-gonic/gin"
	"github.com/theMomax/openefs/config"
	"github.com/theMomax/openefs/handlers"
)

// Config paths
const (
	PathIP   = "server.ip"
	PathPort = "server.port"
)

func init() {
	config.RootCtx.PersistentFlags().StringP(PathIP, "a", "localhost", "server address")
	config.Viper.BindPFlag(PathIP, config.RootCtx.PersistentFlags().Lookup(PathIP))

	config.RootCtx.PersistentFlags().UintP(PathPort, "p", 8080, "server port")
	config.Viper.BindPFlag(PathPort, config.RootCtx.PersistentFlags().Lookup(PathPort))
}

// Run starts the REST api server.
func Run() error {
	switch config.Env() {
	case config.Development:
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	r.Use(gin.Recovery())

	r.Use(config.GinLogrusLogger())

	handlers.Register(&r.RouterGroup)

	return r.Run(config.Viper.GetString(PathIP) + ":" + config.Viper.GetString(PathPort))
}
