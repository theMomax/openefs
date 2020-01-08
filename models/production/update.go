package production

import (
	"math/rand"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/theMomax/openefs/config"
	"github.com/theMomax/openefs/models/production/weather"
	"github.com/theMomax/openefs/utils/metadata"
	timeutils "github.com/theMomax/openefs/utils/time"
)

// Config paths
const (
	PathStepSize               = "models.production.stepsize"
	PathBatchSize              = "models.production.batchsize"
	PathInferenceBatchSize     = "models.production.inferencebatchsize"
	PathConsideredSteps        = "models.production.consideredsteps"
	PathMaximumProductionPower = "models.production.maximumpower"
	PathNormalizationMethod    = "models.production.normalizationmethod"
)

// Normalization methods
const (
	maxpower   = "maxpower"
	averageday = "averageday"
)

func init() {
	config.RootCtx.PersistentFlags().Duration(PathStepSize, time.Hour, "the duration (in seconds) of a single time-step as required by the used production-forecasting-model")
	config.Viper.BindPFlag(PathStepSize, config.RootCtx.PersistentFlags().Lookup(PathStepSize))

	config.RootCtx.PersistentFlags().Uint(PathBatchSize, 24, "the amount of new values required to start an update of the production-model")
	config.Viper.BindPFlag(PathBatchSize, config.RootCtx.PersistentFlags().Lookup(PathBatchSize))

	config.RootCtx.PersistentFlags().Uint(PathInferenceBatchSize, 24, "the amount of steps compiled into a single inference process")
	config.Viper.BindPFlag(PathInferenceBatchSize, config.RootCtx.PersistentFlags().Lookup(PathInferenceBatchSize))

	config.RootCtx.PersistentFlags().Uint(PathConsideredSteps, 2, "the amount of preceding time-steps required for making a prediction")
	config.Viper.BindPFlag(PathConsideredSteps, config.RootCtx.PersistentFlags().Lookup(PathConsideredSteps))

	config.RootCtx.PersistentFlags().Float64(PathMaximumProductionPower, 0, "the installed peak-production power (in Watts)")
	config.Viper.BindPFlag(PathMaximumProductionPower, config.RootCtx.PersistentFlags().Lookup(PathMaximumProductionPower))

	config.RootCtx.PersistentFlags().String(PathNormalizationMethod, maxpower, "the method used for normalizing the power value before passed into the production-model (one of: "+maxpower+", "+averageday+")")
	config.Viper.BindPFlag(PathNormalizationMethod, config.RootCtx.PersistentFlags().Lookup(PathNormalizationMethod))

	config.OnInitialize(func() {
		log = config.NewLogger()
	})

	config.OnInitialize(func() {
		maximumProductionPower = config.Viper.GetFloat64(PathMaximumProductionPower)
		if config.Viper.GetString(PathNormalizationMethod) == maxpower && maximumProductionPower <= 0 {
			config.InvalidConfiguration(PathMaximumProductionPower, "(0, +inf) W")
		}
	})

	config.OnInitialize(func() {
		switch config.Viper.GetString(PathNormalizationMethod) {
		case maxpower:
			normalize = normalizeByMaxPower
			denormalize = denormalizeByMaxPower
		case averageday:
			normalize = normalizeByAvgDay
			denormalize = denormalizeByAvgDay
		default:
			config.InvalidConfiguration(PathNormalizationMethod, maxpower+", "+averageday)
		}
	})
}

var log *logrus.Logger

var (
	maximumProductionPower float64
)

var normalize, denormalize func(float64, time.Time) float64

// Update is the typed equivalence to models.Update for production-updates.
type Update interface {
	Data() *Data
	// Time that Data is associated with. Time is rounded to the duration
	// defined in model.production.stepsize.
	Time() time.Time
	// Meta contains metadata about this update.
	Meta() metadata.Metadata
	// Returns false if the Update was provided by an external source and true
	// if the Update was derived from another Update by this system.
	IsDerived() bool
}

// Data contains the data required by this package's underlying
// production-forecasing-model.
type Data struct {
	// Power holds the average power produced by the system over some duration.
	Power float64 `csv:"production"`
}

var weatherUpdates chan weather.Update
var incomingProductionUpdates chan Update
var outgoingProductionUpdates chan Update

var subscribers = make(map[int64]func(Update), 0)
var sm = &sync.RWMutex{}

// Run starts this model's update-cycle-goroutines.
func Run(bufferSize uint) {
	weatherUpdates = make(chan weather.Update, bufferSize)
	incomingProductionUpdates = make(chan Update, bufferSize)
	outgoingProductionUpdates = make(chan Update, bufferSize)

	// start goroutine, that feeds into the model
	go func() {
		for {
			select {
			case u := <-incomingProductionUpdates:
				u.Data().Power = normalize(u.Data().Power, u.Time())
				handleProductionUpdate(u)
			case wu := <-weatherUpdates:
				handleWeatherUpdate(wu)
			}
		}
	}()

	// start goroutine, that updates the subscribers
	go func() {
		for {
			u := <-outgoingProductionUpdates
			u.Data().Power = denormalize(u.Data().Power, u.Time())
			notify(u)
		}
	}()
	RunAverage()
}

// UpdateWeather receives a update on weather-data. This call may block if the
// system is overloaded. To prevent this, specify a timeout after with to abort.
func UpdateWeather(update weather.Update, timeout ...time.Duration) (ok bool) {
	if update != nil {
		if len(timeout) == 1 {
			select {
			case weatherUpdates <- update:
				return true
			case <-time.After(timeout[0]): // Timeout must not be mocked!
				return false
			}
		} else {
			weatherUpdates <- update
			return true
		}
	}
	return false
}

// UpdateProduction receives a update on production-data. This call may block if
// the system is overloaded. To prevent this, specify a timeout after with to
// abort.
func UpdateProduction(update Update, timeout ...time.Duration) (ok bool) {
	if update != nil {
		if len(timeout) == 1 {
			select {
			case incomingProductionUpdates <- update:
				return true
			case <-time.After(timeout[0]): // Timeout must not be mocked!
				return false
			}
		} else {
			incomingProductionUpdates <- update
			return true
		}
	}
	return false
}

// Subscribe registers a callback to be called each time, when the underlying
// model creates new output. It returns the id required for unsubscribing. It
// returns -1, if callback is nil.
func Subscribe(callback func(Update)) int64 {
	if callback == nil {
		return -1
	}

	id := rand.Int63()

	sm.Lock()
	subscribers[id] = callback
	sm.Unlock()
	return id
}

// Unsubscribe the callback with the given id.
func Unsubscribe(id int64) {
	sm.Lock()
	delete(subscribers, id)
	sm.Unlock()
}

func notify(update Update) {
	sm.RLock()
	for _, s := range subscribers {
		go s(update)
	}
	sm.RUnlock()
}

// Round rounds the given time to the duration this model works on.
func Round(t time.Time) time.Time {
	return timeutils.Round(t, config.Viper.GetDuration(PathStepSize))
}

type update struct {
	data    *Data
	time    time.Time
	meta    metadata.Metadata
	derived bool
}

func (u *update) Data() *Data {
	return u.data
}

func (u *update) Time() time.Time {
	return u.time
}

func (u *update) Meta() metadata.Metadata {
	return u.meta
}

func (u *update) IsDerived() bool {
	return u.derived
}

func normalizeByMaxPower(p float64, t time.Time) float64 {
	return p / maximumProductionPower
}

func denormalizeByMaxPower(p float64, t time.Time) float64 {
	return p * maximumProductionPower
}

func normalizeByAvgDay(p float64, t time.Time) float64 {
	n := 1.0
	if p != 0 {
		n = p
	}
	if v, ok := GetNonDerived(0, uint(t.Hour())); ok && v != 0 {
		n = v
	}
	return p / n
}

func denormalizeByAvgDay(p float64, t time.Time) float64 {
	n := 1.0
	if p != 0 {
		n = p
	}
	if v, ok := GetNonDerived(0, uint(t.Hour())); ok && v != 0 {
		n = v
	}
	return p * n
}
