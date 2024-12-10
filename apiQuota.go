package main

import "sync"

type ApiQuota struct {
	sync.Mutex
	apiUsage *ApiUsage
	updated  bool
	cond     *sync.Cond
}

func (a *ApiQuota) update() error {
	a.Lock()
	defer a.Unlock()
	if apiUsage, err := GetApiUsage(); err != nil {
		return err
	} else {
		a.apiUsage = apiUsage
		a.updated = true
		a.cond.Broadcast()
		return nil
	}
}

func (a *ApiQuota) current() *ApiUsage {
	a.Lock()
	defer a.Unlock()

	for !a.updated {
		a.cond.Wait()
	}
	return a.apiUsage
}
