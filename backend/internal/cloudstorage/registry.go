package cloudstorage

import (
	"fmt"
	"sort"
	"sync"
)

// Registry holds the set of known drivers. Drivers register themselves at
// startup; runtime code only reads.
type Registry struct {
	mu      sync.RWMutex
	drivers map[string]Driver
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{drivers: map[string]Driver{}}
}

// Register adds a driver under its Type(). Returns ErrDuplicateDriver if a
// driver with that type is already registered.
func (r *Registry) Register(d Driver) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.drivers[d.Type()]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateDriver, d.Type())
	}
	r.drivers[d.Type()] = d
	return nil
}

// Get returns the driver for providerType or ErrUnknownProvider.
func (r *Registry) Get(providerType string) (Driver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.drivers[providerType]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownProvider, providerType)
	}
	return d, nil
}

// DriverMeta is the minimal metadata returned to the admin UI.
type DriverMeta struct {
	Type        string `json:"type"`
	DisplayName string `json:"display_name"`
	AuthFlow    string `json:"auth_flow"`
}

// List returns registered drivers in stable order for admin-UI dropdowns.
func (r *Registry) List() []DriverMeta {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]DriverMeta, 0, len(r.drivers))
	for _, d := range r.drivers {
		out = append(out, DriverMeta{
			Type:        d.Type(),
			DisplayName: d.DisplayName(),
			AuthFlow:    d.AuthFlow().String(),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DisplayName < out[j].DisplayName })
	return out
}
