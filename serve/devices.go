package serve

import (
	"sync"

	"github.com/teh-hippo/foxess-exporter/foxess"
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

func (x *DeviceCache) Set(devices []foxess.Device) {
	x.cond.L.Lock()
	defer x.cond.L.Unlock()
	x.DeviceIds = make([]string, len(devices))

	for i, device := range devices {
		x.DeviceIds[i] = device.DeviceSerialNumber
	}

	x.cond.Broadcast()
}

func (x *DeviceCache) Get() *[]string {
	x.cond.L.Lock()
	defer x.cond.L.Unlock()
	if x.DeviceIds == nil {
		x.cond.Wait()
	}
	return &x.DeviceIds
}

func (x *DeviceCache) Initalise(filtered map[string]bool) {
	if len(filtered) == 0 {
		return
	}
	x.cond.L.Lock()
	defer x.cond.L.Unlock()
	x.DeviceIds = make([]string, 0, len(filtered))
	for deviceId := range filtered {
		x.DeviceIds = append(x.DeviceIds, deviceId)
	}
}
