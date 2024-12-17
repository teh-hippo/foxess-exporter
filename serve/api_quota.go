package serve

import (
	"sync"

	"github.com/teh-hippo/foxess-exporter/foxess"
)

type ApiQuota struct {
	value *foxess.APIUsage
	cond  *sync.Cond
}

func NewApiCache() *ApiQuota {
	return &ApiQuota{
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

func (x *ApiQuota) Set(value *foxess.APIUsage) {
	x.cond.L.Lock()
	defer x.cond.L.Unlock()
	x.value = value
	x.cond.Broadcast()
}

func (x *ApiQuota) IsQuotaAvailable() bool {
	x.cond.L.Lock()
	defer x.cond.L.Unlock()
	if x.value == nil {
		x.cond.Wait()
	}
	return x.value.Remaining > 0
}
