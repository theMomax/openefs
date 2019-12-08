package cli

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/theMomax/openefs/cache"
	"github.com/theMomax/openefs/config"
	"github.com/theMomax/openefs/models"
	"github.com/theMomax/openefs/server"
)

func init() {
	config.RootCtx.Run = run
}

// Execute executes the root command.
func Execute() error {
	return config.RootCtx.Execute()
}

func run(cmd *cobra.Command, args []string) {
	models.Run()
	cache.Run()
	log.WithError(server.Run()).Panic("Unexpected panic!")
}
