package config

import (
	"sync"
	"time"

	"github.com/futurehomeno/cliffhanger/config"
	"github.com/futurehomeno/cliffhanger/storage"
)

// Config is a model containing all application configuration settings.
type Config struct {
	config.Default

    // TODO: Add specific configuration settings for your application. Don't forget to provide setters and getters where required.
}

// New creates new instance of a configuration object.
func New(workDir string) *Config {
	return &Config{
		Default: config.NewDefault(workDir),
	}
}

// NewConfigService creates a new configuration service.
func NewConfigService(workDir string) *Service {
	return &Service{
		Storage: config.NewStorage(New(workDir), workDir),
		lock:    &sync.RWMutex{},
	}
}

// Service is a configuration service responsible for:
// - providing concurrency safe access to settings,
// - persistence of settings.
type Service struct {
	storage.Storage
	lock *sync.RWMutex
}

// GetLogLevel allows to safely access a configuration setting.
func (cs *Service) GetLogLevel() string {
	cs.lock.RLock()
	defer cs.lock.RUnlock()

	return cs.Storage.Model().(*Config).LogLevel
}

// SetLogLevel allows to safely set and persist a configuration setting.
func (cs *Service) SetLogLevel(value string) error {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	cs.Storage.Model().(*Config).ConfiguredAt = time.Now().Format(time.RFC3339)
	cs.Storage.Model().(*Config).LogLevel = value

	return cs.Storage.Save()
}

// Factory is a factory method returning the configuration object without default settings.
func Factory() interface{} {
	return &Config{}
}
