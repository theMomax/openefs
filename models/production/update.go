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
	PathStepSize = "models.production.stepsize"
	PathSteps    = "models.production.steps"
)

func init() {
	config.RootCtx.PersistentFlags().Duration(PathStepSize, time.Hour, "the duration (in seconds) of a single time-step as required by the used production-forecasting-model")
	config.Viper.BindPFlag(PathStepSize, config.RootCtx.PersistentFlags().Lookup(PathStepSize))

	config.RootCtx.PersistentFlags().Uint(PathSteps, 120, "the amount of time-steps to predict for")
	config.Viper.BindPFlag(PathSteps, config.RootCtx.PersistentFlags().Lookup(PathSteps))
	config.OnInitialize(func() {
		log = config.NewLogger()
	})
}

var log *logrus.Logger

// Update is the typed equivalence to models.Update for production-updates.
type Update interface {
	Data() *Data
	// Time that Data is associated with. Time is rounded to the duration
	// defined in model.production.stepsize.
	Time() time.Time
	// Meta contains metadata about this update.
	Meta() metadata.Metadata
}

// Data contains the data required by this package's underlying
// production-forecasing-model.
type Data struct {
	// Power holds the average power produced by the system over some duration.
	Power float64 `csv:"relativePower"`
}

var weatherUpdates chan weather.Update
var incomingProductionUpdates chan Update
var outgoingProductionUpdates chan Update

var subscribers = make(map[int64]func(Update), 0)
var sm = &sync.RWMutex{}

// Run starts this model's update-cycle-goroutines.
func Run(bufferSize uint) {
	// start goroutine, that feeds into the model
	go func() {
		for {
			// TODO: implement (mock implementation for now)
			select {
			case u := <-incomingProductionUpdates:
				outgoingProductionUpdates <- u
			case wu := <-weatherUpdates:
				log.WithField("timestamp", wu.Time()).WithField("data", wu.Data()).WithField("meta", wu.Meta()).Info("received (and discarded) weather-update")
			}
		}
	}()

	// start goroutine, that updates the subscribers
	go func() {
		for {
			notify(<-outgoingProductionUpdates)
		}
	}()
}

// UpdateWeather receives a update on weather-data. This call may block if the
// system is overloaded. To prevent this, specify a timeout after with to abort.
func UpdateWeather(update weather.Update, timeout ...time.Duration) (ok bool) {
	if update != nil {
		if len(timeout) == 1 {
			select {
			case weatherUpdates <- update:
				return true
			case <-timeutils.After(timeout[0]):
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
			case <-timeutils.After(timeout[0]):
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
