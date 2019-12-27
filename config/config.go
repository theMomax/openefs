package config

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ========== Injectable Constants =============================================

// To be injected during build
const (
	Version = "development"

	GitCommit = "unknown"

	Author = "Max Obermeier (themomax@icloud.com)"

	License = "MIT"
)

const (
	// ApplicationName is used for configuration paths and environment
	// variables.
	ApplicationName = "openefs"

	// ConfigName is the configuration file's name without prefix.
	ConfigName = "config"
)

var (
	// ConfigPaths specifies where to look for configuration files.
	ConfigPaths = [...]string{".", "/etc/" + ApplicationName, "$HOME/" + ApplicationName}
)

// ========== Config setup =====================================================

// Config paths
const (
	PathConfig      = "config"
	PathConfigPaths = "config_paths"
	PathEnv         = "env"
	PathAuthor      = "author"
	PathLicense     = "license"
)

var (
	// RootCtx is the root command. It may be used by other packages to register
	// flags and bind them to the viper configuration.
	RootCtx = &cobra.Command{
		Use:     ApplicationName,
		Short:   "Open Energy Forecasting Service",
		Long:    `OpenEFS is an Open Energy Forecasting Service providing predictions on the local production and consumption of energy.`,
		Version: Version + " (" + GitCommit + ")",
	}
)

// Viper instance to be used by everything that relies on this config-package.
var Viper = viper.New()

func init() {
	// initialize config flags
	RootCtx.PersistentFlags().StringP(PathEnv, "e", Development, "application context")
	Viper.BindPFlag(PathEnv, RootCtx.PersistentFlags().Lookup(PathEnv))

	RootCtx.PersistentFlags().StringP(PathConfig, "c", ConfigName, "configuration file's name (without extension)")
	Viper.BindPFlag(PathConfig, RootCtx.PersistentFlags().Lookup(PathConfig))

	RootCtx.PersistentFlags().StringArray(PathConfigPaths, ConfigPaths[:], "directories in which to look for config files")
	Viper.BindPFlag(PathConfigPaths, RootCtx.PersistentFlags().Lookup(PathConfigPaths))

	RootCtx.PersistentFlags().String(PathAuthor, Author, "author name for copyright attribution")
	Viper.BindPFlag(PathAuthor, RootCtx.PersistentFlags().Lookup(PathAuthor))

	RootCtx.PersistentFlags().String(PathLicense, License, "name of license for the project")
	Viper.BindPFlag(PathLicense, RootCtx.PersistentFlags().Lookup(PathLicense))

	OnInitialize(func() {
		log = NewLogger()
	})
	OnInitialize(loadConfiguration)
	OnInitialize(initializeLogrus)
}

var log *logrus.Logger

// OnInitialize registers a function to be called after all the configuration-
// parameters have been collected, but before the command is executed.
func OnInitialize(callbacks ...func()) {
	cobra.OnInitialize(callbacks...)
}

// InvalidConfiguration is a public helper-function, that is to be used for
// complaining about invalid configuration.
func InvalidConfiguration(identifier string, expected interface{}) {
	log.WithFields(logrus.Fields{
		"identifier": identifier,
		"expected":   expected,
		"actual":     Viper.Get(identifier),
	}).Fatal("Invalid configuration!")
}

func loadConfiguration() {
	// search for environment variables
	Viper.SetEnvPrefix(strings.ToUpper(ApplicationName))
	Viper.AutomaticEnv()

	// read config file
	Viper.SetConfigName(Viper.GetString(PathConfig))
	for _, p := range ConfigPaths {
		Viper.AddConfigPath(p)
	}

	if err := Viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.WithField("checked_directories", ConfigPaths[:]).Debug("No config file found!")
		} else {
			log.WithError(err).Fatal("Could not read file!")
		}
	}
}

// ========== Config API =======================================================

// Environment is the type specifying the application's context.
type Environment string

// Environment contexts.
const (
	Production  = "prod"
	Development = "dev"
)

// Env returns the application's context.
func Env() Environment {
	return Environment(Viper.GetString(PathEnv))
}
