package server

import (
	"errors"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	viper.BindPFlag(PathIP, config.RootCtx.PersistentFlags().Lookup(PathIP))

	config.RootCtx.PersistentFlags().UintP(PathPort, "p", 8080, "server port")
	viper.BindPFlag(PathPort, config.RootCtx.PersistentFlags().Lookup(PathPort))
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

	r.Use(logrusLogger())

	handlers.Register(r.RouterGroup)

	return r.Run(viper.GetString(PathIP) + ":" + viper.GetString(PathPort))
}

func logrusLogger() gin.HandlerFunc {
	logger := log.New()
	logger.SetFormatter(config.LogFormatter())
	logger.SetLevel(log.GetLevel())

	return gin.LoggerWithFormatter(func(p gin.LogFormatterParams) string {
		fields := logger.WithFields(log.Fields{
			"status_code":  p.StatusCode,
			"latency_time": p.Latency,
			"client_ip":    p.ClientIP,
			"req_method":   p.Method,
			"req_uri":      p.Request.RequestURI,
		})

		if p.ErrorMessage != "" {
			fields.WithError(errors.New(p.ErrorMessage)).Error("GIN")
		}

		fields.Info("GIN")

		return ""
	})
}
