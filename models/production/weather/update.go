package weather

import (
	"time"

	"github.com/theMomax/openefs/utils/metadata"
)

// Update is the typed equivalence to models.Update for weather-updates.
type Update interface {
	Data() *Data
	// Time that Data is associated with. Time is rounded to the duration
	// defined in model.production.stepsize.
	Time() time.Time
	// Meta contains metadata about this update.
	Meta() metadata.Metadata
}

// Data holds the weather-features required by this service.
type Data struct{}
