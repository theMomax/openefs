package generic

import (
	"math/rand"
	"sync"
	"time"
)

type subscriber struct {
	callback       func(Element)
	observedHashes []interface{}
	observers      []func(interface{}) bool
}

type Outdating interface {
	Time() time.Time
}

type Element interface {
	Outdating
	Hash() interface{}
}

type Cache struct {
	cache map[interface{}]Element
	cm    *sync.RWMutex

	subscribers map[int64]*subscriber
	sm          *sync.RWMutex

	outdated func(interface{}) bool
}

// NewCache returns a Cache for Elements. The Cache automatically deletes
// Elements, that are outdated. I.e. if outdated returns true for an Element's
// Hash, it is deleted.
func NewCache(outdated func(interface{}) bool) *Cache {
	return &Cache{
		cache:       make(map[interface{}]Element),
		cm:          &sync.RWMutex{},
		subscribers: make(map[int64]*subscriber, 0),
		sm:          &sync.RWMutex{},
		outdated:    outdated,
	}
}

func (c *Cache) Update(e Element) {

	c.cm.Lock()
	c.cache[e.Hash()] = e
	c.cm.Unlock()

	go c.notify(e)

	c.cm.RLock()
	for h, v := range c.cache {
		if c.outdated(v.Time()) {
			c.cm.RUnlock()
			c.cm.Lock()
			delete(c.cache, h)
			c.cm.Unlock()
			c.cm.RLock()
		}
	}
	c.cm.RUnlock()
}

// Get returns the latest available value for hash.
func (c *Cache) Get(hash interface{}) Element {
	c.cm.RLock()
	defer c.cm.RUnlock()
	return c.cache[hash]
}

// Subscribe registers a callback to be called each time, when new input is
// cached and right after calling this function with the currently cached value.
// If there are observedHashes or observers given, the callback is only called,
// if the update is related to one of those hashes. It returns the id
// required for unsubscribing. It returns -1, if callback is nil.
func (c *Cache) Subscribe(callback func(Element), observedHashes []interface{}, observers []func(interface{}) bool) int64 {
	if callback == nil {
		return -1
	}

	if observedHashes == nil {
		observedHashes = make([]interface{}, 0)
	}
	if observers == nil {
		observers = make([]func(interface{}) bool, 0)
	}

	id := rand.Int63()

	s := subscriber{
		callback:       callback,
		observedHashes: observedHashes,
		observers:      observers,
	}

	c.sm.Lock()
	c.subscribers[id] = &s
	c.sm.Unlock()

	go func() {
		c.cm.RLock()
		for _, u := range c.cache {
			c.notify(u, &s)
		}
		c.cm.RUnlock()
	}()

	return id
}

// Unsubscribe the callback with the given id.
func (c *Cache) Unsubscribe(id int64) {
	c.sm.Lock()
	delete(c.subscribers, id)
	c.sm.Unlock()
}

func (c *Cache) notify(e Element, subs ...*subscriber) {
	if c.outdated(e.Time()) {
		return
	}

	// If no subs are given, take the global subscribers. Also, update the
	// global subscribers list by removing outdated ones.
	if len(subs) == 0 {
		c.sm.RLock()
		for id, s := range c.subscribers {
			outdated := len(s.observers) == 0
			for i := len(s.observedHashes); i >= 0; i-- {
				if !c.outdated(s.observedHashes[i]) {
					outdated = false
					break
				} else {
					s.observedHashes = append(s.observedHashes[:i], s.observedHashes[i+1:]...)
				}
			}
			if outdated {
				c.sm.RUnlock()
				c.Unsubscribe(id)
				c.sm.RLock()
			}

			subs = append(subs, s)
		}
		c.sm.RUnlock()
	}

	// check if subscriber subscribed to the update's hash and notify in case
	hash := e.Hash()

outer:
	for _, s := range subs {
		for _, a := range s.observedHashes {
			if a == hash {
				go s.callback(e)
				continue outer
			}
		}
		for _, r := range s.observers {
			if r(hash) {
				go s.callback(e)
				continue outer
			}
		}
	}
}
