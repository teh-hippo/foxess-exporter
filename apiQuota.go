package main

import "sync"

type ApiQuota struct {
	apiUsage *ApiUsage
	updated  bool
	cond     *sync.Cond
}

func NewApiQuota() *ApiQuota {
	return &ApiQuota{cond: sync.NewCond(&sync.Mutex{})}
}

func (a *ApiQuota) update() error {
	a.cond.L.Lock()
	defer a.cond.L.Unlock()
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
	a.cond.L.Lock()
	defer a.cond.L.Unlock()
	for !a.updated {
		a.cond.Wait()
	}
	return a.apiUsage
}
