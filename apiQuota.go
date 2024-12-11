package main

import (
	"fmt"
	"sync"
)

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
	apiUsage, err := GetApiUsage()
	if err != nil {
		return fmt.Errorf("failed to get API usage: %w", err)
	}

	a.apiUsage = apiUsage
	a.updated = true
	a.cond.Broadcast()
	return nil
}

func (a *ApiQuota) current() *ApiUsage {
	a.cond.L.Lock()
	defer a.cond.L.Unlock()
	for !a.updated {
		a.cond.Wait()
	}
	return a.apiUsage
}
