package serve

import (
	"sync"
)

type DeviceCache struct {
	DeviceIDs []string
	cond      *sync.Cond
}

func NewDeviceCache() *DeviceCache {
	return &DeviceCache{
		cond:      sync.NewCond(&sync.Mutex{}),
		DeviceIDs: nil,
	}
}

func (x *DeviceCache) Set(deviceIDs []string) {
	x.cond.L.Lock()
	defer x.cond.L.Unlock()
	x.DeviceIDs = deviceIDs
	x.cond.Broadcast()
}

func (x *DeviceCache) Get() []string {
	x.cond.L.Lock()
	defer x.cond.L.Unlock()

	if x.DeviceIDs == nil {
		x.cond.Wait()
	}

	return x.DeviceIDs
}
