package service

import (
	"sync"
	"time"
)

type ServiceRegistry struct {
	services map[string]*Service
	regMu    *sync.Mutex
}

func NewServiceRegistry() *ServiceRegistry {
	registry := &ServiceRegistry{}
	registry.regMu = &sync.Mutex{}
	registry.services = make(map[string]*Service)
	return registry
}

func (self *ServiceRegistry) RegisterService(srv *Service) {
	self.regMu.Lock()
	defer self.regMu.Unlock()
	if _, ok := self.services[srv.Name]; ok {
		log.Info("[service] Service `%s` already registered", srv.Name)
		return
	}
	self.services[srv.Name] = srv
}

func (self *ServiceRegistry) GetServiceByName(name string) *Service {
	if srv, ok := self.services[name]; ok {
		return srv
	}
	return nil
}

type Service struct {
	Name           string
	Version        string
	Online         bool
	LastPing       time.Time
	RespondedCount int
}

func NewService() *Service {
	service := &Service{}
	return service
}
