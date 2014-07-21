package service

import (
	"time"
	"sync"
)

type ServiceRegistry struct {
	services map[string]*Service
	regMu *sync.Mutex
}

func NewServiceRegistry() *ServiceRegistry {
	registry := &ServiceRegistry{}
	registry.regMu = &sync.Mutex{}
	registry.services = make(map[string]*Service)
	return registry
}

func (self *ServiceRegistry) RegisterService(name string, srv *Service) {
	self.regMu.Lock()
	defer self.regMu.Unlock()
	if _, ok := self.services[name]; ok {
		log.Info("[service] Service `%s` already registered", name)
		return
	}
	self.services[name] = srv
}

type Service struct {
	Name string
	Online bool
	LastPing time.Time
	RespondedCount int
}

func NewService() *Service {
	service := &Service{}
	return service
}
