package config

import (
	"errors"
	"flag"
	"sync"
)

var (
	cfg  = new(Config)
	once = sync.Once{}

	// ErrAlreadyInitialized indicates > 1 execution of Init()
	ErrAlreadyInitialized = errors.New("config already initialized")
)

// Config represents the application configuration
type Config struct {
	Auth    string
	Verbose bool
}

// Init initializes configuration nonce
func Init() (*Config, error) {
	err := ErrAlreadyInitialized
	once.Do(func() {
		err = nil
		cfg = initialize()
	})
	return cfg, err
}

func initialize() *Config {
	auth := flag.String("auth", "", "Basic auth: example user:password")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "enable verbose logging")

	flag.Parse()

	// TODO parse Basic Auth
	_ = auth

	return cfg
}
