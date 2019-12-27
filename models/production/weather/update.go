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
type Data struct {
	CloudCover               float64 `csv:"cloudCover"`
	PrecipitationProbability float64 `csv:"precipitationProbability"`
	PrecipitationIntensity   float64 `csv:"precipitationIntensity"`
	WindSpeed                float64 `csv:"windSpeed"`
	WindGust                 float64 `csv:"windGust"`
	ApparentTemperature      float64 `csv:"apparentTemperature"`
	Temperature              float64 `csv:"temperature"`
	Humidity                 float64 `csv:"humidity"`
	DewPoint                 float64 `csv:"dewPoint"`
	Visibility               float64 `csv:"visibility"`
	UVIndex                  float64 `csv:"uvIndex"`
}
