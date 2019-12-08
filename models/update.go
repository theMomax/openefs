package models

import (
	"time"

	"github.com/spf13/viper"
	"github.com/theMomax/openefs/config"
	"github.com/theMomax/openefs/models/production"
	"github.com/theMomax/openefs/utils/metadata"
)

// Config paths
const (
	PathBufferSize = "models.buffersize"
)

func init() {
	config.RootCtx.PersistentFlags().Uint(PathBufferSize, 100, "the amount of model-update-requests per update-type, that can be buffered")
	viper.BindPFlag(PathBufferSize, config.RootCtx.PersistentFlags().Lookup(PathBufferSize))
}

// Update is any type, that contains update-information for a model. This can be
// new training-data or new data for inference. Use the typed equivalents
// contained in subpackages where possible.
type Update interface {
	Data() interface{}
	// Time that Data is associated with. Time may be rounded. This depends on
	// the implementation.
	Time() time.Time
	// Meta contains metadata about this update.
	Meta() metadata.Metadata
}

// Run parametrizes and starts the update-cycle-goroutines of all subpackages.
func Run() {
	production.Run(viper.GetUint(PathBufferSize))
}
