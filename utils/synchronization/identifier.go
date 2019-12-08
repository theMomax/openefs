package synchronization

import "sync"

var m = &sync.Mutex{}
var globalID uint64 = 0

// AttachID provides the idconsumer with a unique identifier, that is guranteed
// to be ordered by time of the call to this function.
func AttachID(idconsumer func(id uint64)) {
	m.Lock()
	idconsumer(globalID)
	globalID++
	m.Unlock()
}
