package serve

import (
	"sync"
)

type DeviceCache struct {
	DeviceIds []string
	cond      *sync.Cond
}

func NewDeviceCache() *DeviceCache {
	return &DeviceCache{
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

func (x *DeviceCache) Set(deviceIds []string) {
	x.cond.L.Lock()
	defer x.cond.L.Unlock()
	x.DeviceIds = deviceIds
	x.cond.Broadcast()
}

func (x *DeviceCache) Get() []string {
	x.cond.L.Lock()
	defer x.cond.L.Unlock()
	if x.DeviceIds == nil {
		x.cond.Wait()
	}
	return x.DeviceIds
}
