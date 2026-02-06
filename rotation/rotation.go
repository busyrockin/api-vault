package rotation

import (
	"context"
	"sync"
	"time"
)

// RotatableField identifies which credential field a plugin can rotate.
type RotatableField string

const (
	FieldSecretKey RotatableField = "secret_key"
	FieldPublicKey RotatableField = "public_key"
	FieldURL       RotatableField = "url"
)

// CredentialInfo is the read-only view a plugin receives. No DB dependency.
type CredentialInfo struct {
	Name      string
	APIType   string
	SecretKey *string
	PublicKey *string
	URL       *string
	Config    map[string]string
}

// Result carries rotation output back to the caller.
type Result struct {
	NewSecretKey *string
	NewPublicKey *string
	NewURL       *string
	KeyID        string
	OldKeyGrace  time.Duration
	Metadata     map[string]string
}

// Config is the per-rotation configuration passed to a plugin.
type Config map[string]interface{}

// ConfigField describes one input a plugin needs.
type ConfigField struct {
	Name        string
	Description string
	Required    bool
	Secret      bool
}

// ConfigSchema describes all inputs a plugin needs.
type ConfigSchema struct {
	Fields []ConfigField
}

// Plugin is the interface every rotation provider implements.
type Plugin interface {
	Name() string
	RotatableFields() []RotatableField
	Rotate(ctx context.Context, cred CredentialInfo, cfg Config) (*Result, error)
	Validate(cred CredentialInfo) error
	ConfigSchema() ConfigSchema
}

// Registry holds registered rotation plugins keyed by API type.
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
}

func NewRegistry() *Registry {
	return &Registry{plugins: make(map[string]Plugin)}
}

func (r *Registry) Register(p Plugin) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.plugins[p.Name()] = p
}

func (r *Registry) Get(name string) (Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.plugins[name]
	return p, ok
}

func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.plugins))
	for k := range r.plugins {
		names = append(names, k)
	}
	return names
}

var globalRegistry = NewRegistry()

func GetGlobalRegistry() *Registry { return globalRegistry }
