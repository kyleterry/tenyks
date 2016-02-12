package service

import (
	"fmt"
	"sync"
	"time"
	"github.com/Xe/uuid"
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
	if _, ok := self.services[srv.UUID.String()]; ok {
		log.Info("[service] Service `%s` already registered", srv.UUID.String())
		srv, _ = self.services[srv.UUID.String()]
		srv.Online = ServiceOnline
		return
	}
	self.services[srv.UUID.String()] = srv
}

func (self *ServiceRegistry) GetServiceByUUID(uuid string) *Service {
	self.regMu.Lock()
	defer self.regMu.Unlock()
	if srv, ok := self.services[uuid]; ok {
		return srv
	}
	return nil
}

func (self *ServiceRegistry) GetServiceByName(name string) *Service {
	self.regMu.Lock()
	defer self.regMu.Unlock()
	for _, service := range self.services {
		if service.Name == name {
			return service
		}
	}
	return nil
}

func (self *ServiceRegistry) IsService(name string) bool {
	self.regMu.Lock()
	defer self.regMu.Unlock()
	for _, service := range self.services {
		if service.Name == name {
			return true
		}
	}
	return false
}

type Service struct {
	Name           string
	UUID           uuid.UUID
	Version        string
	Description    string
	Online         bool
	LastPing       time.Time
	LastPong       time.Time
	RespondedCount int
}

func NewService() *Service {
	service := &Service{}
	return service
}

func (self *Service) String() string {
	state := "offline"
	if self.Online {
		state = "online"
	}
	return fmt.Sprintf("%s (%s) - %s", self.Name, state, self.Description)
}
