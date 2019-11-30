package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Log formatter options
const (
	json   = "json"
	logfmt = "logfmt"
	tty    = "tty"
)

// Config paths
const (
	PathLevel     = "loglevel"
	PathFormatter = "logformatter"
)

func init() {
	// provide configuration
	RootCtx.PersistentFlags().UintP(PathLevel, "l", uint(log.InfoLevel), "log level (Panic: 0, Fatal: 1, Error: 2, Warning: 3, Info: 4, Debug: 5, Trace: 6)")
	viper.BindPFlag(PathLevel, RootCtx.PersistentFlags().Lookup(PathLevel))

	RootCtx.PersistentFlags().String(PathFormatter, tty, "log format")
	viper.BindPFlag(PathFormatter, RootCtx.PersistentFlags().Lookup(PathFormatter))
}

func initializeLogrus() {
	// choose formatter
	log.SetFormatter(LogFormatter())

	// set logging level
	lvl := viper.GetUint32(PathLevel)
	if uint32(log.TraceLevel) < lvl {
		InvalidConfiguration(PathLevel, [...]log.Level{log.PanicLevel, log.FatalLevel, log.ErrorLevel, log.WarnLevel, log.InfoLevel, log.DebugLevel, log.TraceLevel})
	}
	log.SetLevel(log.Level(lvl))
}

// LogFormatter returns the configured logrus formatter.
func LogFormatter() log.Formatter {
	switch viper.GetString(PathFormatter) {
	case json:
		return &log.JSONFormatter{}
	case logfmt:
		return &log.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		}
	case tty:
		return &log.TextFormatter{}
	default:
		InvalidConfiguration(PathFormatter, [...]string{json, logfmt, tty})
		return nil
	}
}
