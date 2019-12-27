package production

import (
	"math/rand"
	"sync"
	"time"

	"github.com/theMomax/openefs/config"
	models "github.com/theMomax/openefs/models/production"
	timeutils "github.com/theMomax/openefs/utils/time"
)

func init() {
	config.OnInitialize(func() {
		outdatedAfter = config.Viper.GetDuration(models.PathStepSize)
	})
}

type subscriber struct {
	callback           func(models.Update)
	timestamps         []time.Time
	relativeTimestamps []time.Duration
}

var outdatedAfter time.Duration

var cache = make(map[time.Time]models.Update)
var cm = &sync.RWMutex{}

var subscribers = make(map[int64]*subscriber, 0)
var sm = &sync.RWMutex{}

// Run initializes the caching package.
func Run() {
	models.Subscribe(handleUpdate)
}

func handleUpdate(update models.Update) {

	cm.Lock()
	cache[models.Round(update.Time())] = update
	cm.Unlock()

	go notify(update)

	cm.RLock()
	for t := range cache {
		if timeutils.Since(t) >= outdatedAfter {
			cm.RUnlock()
			cm.Lock()
			delete(cache, t)
			cm.Unlock()
			cm.RLock()
		}
	}
	cm.RUnlock()
}

// Update returns the latest available update for time t.
func Update(t time.Time) models.Update {
	return cache[models.Round(t)]
}

// Subscribe registers a callback to be called each time, when the underlying
// model creates new output and immediately with the currently cached value. If
// there are (relative) timestamps given, the callback is only called, if the
// update is related to one of those timestamps. The callback is automatically
// unsubscribed when all the given absolute timestamps are in the past. It
// returns the id required for unsubscribing. It returns -1, if callback is nil.
func Subscribe(callback func(models.Update), absolute []time.Time, relative []time.Duration) int64 {
	if callback == nil {
		return -1
	}

	if absolute == nil {
		absolute = make([]time.Time, 0)
	}
	if relative == nil {
		relative = make([]time.Duration, 0)
	}

	id := rand.Int63()

	// round timestamps
	for i, t := range absolute {
		absolute[i] = models.Round(t)
	}

	// remove outdated relative timestamps
	for i, d := range relative {
		if d <= -1*outdatedAfter {
			relative = append(relative[:i], relative[i+1:]...)
		}
	}

	s := subscriber{
		callback:           callback,
		timestamps:         absolute,
		relativeTimestamps: relative,
	}

	sm.Lock()
	subscribers[id] = &s
	sm.Unlock()

	go func() {
		cm.RLock()
		for _, u := range cache {
			notify(u, &s)
		}
		cm.RUnlock()
	}()

	return id
}

// Unsubscribe the callback with the given id.
func Unsubscribe(id int64) {
	sm.Lock()
	delete(subscribers, id)
	sm.Unlock()
}

func notify(update models.Update, subs ...*subscriber) {
	if timeutils.Since(update.Time()) >= outdatedAfter {
		return
	}

	// If no subs are given, take the global subscribers. Also, update the
	// global subscribers list by removing outdated ones.
	if len(subs) == 0 {
		sm.RLock()
		for id, s := range subscribers {
			outdated := len(s.relativeTimestamps) == 0
			for i := len(s.timestamps); i >= 0; i-- {
				if timeutils.Since(s.timestamps[i]) < outdatedAfter {
					outdated = false
					break
				} else {
					s.timestamps = append(s.timestamps[:i], s.timestamps[i+1:]...)
				}
			}
			if outdated {
				sm.RUnlock()
				Unsubscribe(id)
				sm.RLock()
			}

			subs = append(subs, s)
		}
		sm.Unlock()
	}

	// check if subscriber subscribed to the update's timestamp and notify in case
	updatetime := models.Round(update.Time())

outer:
	for _, s := range subs {
		for _, a := range s.timestamps {
			if a == updatetime {
				go s.callback(update)
				continue outer
			}
		}
		for _, r := range s.relativeTimestamps {
			if models.Round(timeutils.Now().Add(r)) == updatetime {
				go s.callback(update)
				continue outer
			}
		}
	}
}
