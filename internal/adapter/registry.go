package adapter

import "sync"

type Registry interface {
	RegisterAdapter(Adapter)
	GetAdaptersFor(AdapterType) map[string]Adapter
}

type registry struct {
	adapters map[AdapterType]map[string]Adapter
	sync.Mutex
}

func (r *registry) RegisterAdapter(a Adapter) {
	r.Lock()
	defer r.Unlock()

	if r.adapters == nil {
		r.adapters = make(map[AdapterType]map[string]Adapter)
	}

	if r.adapters[a.GetType()] == nil {
		r.adapters[a.GetType()] = make(map[string]Adapter)
	}

	r.adapters[a.GetType()][a.GetName()] = a
}

func (r *registry) GetAdaptersFor(at AdapterType) map[string]Adapter {
	if r.adapters == nil || r.adapters[at] == nil {
		return map[string]Adapter{}
	}

	return r.adapters[at]
}

func NewRegistry() *registry {
	return &registry{
		adapters: make(map[AdapterType]map[string]Adapter),
	}
}
