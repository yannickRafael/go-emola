package config

import (
	"errors"
	"time"
)

// Environment definition to differentiate between UAT and Production.
type Environment string

const (
	EnvUAT  Environment = "UAT"
	EnvPROD Environment = "PROD"
)

// Config holds all necessary credentials and settings to communicate with e-Mola.
type Config struct {
	// Provided by Movitel
	PartnerCode string
	PartnerKey  string
	Username    string
	Password    string

	// Operational Settings
	Environment Environment
	Timeout     time.Duration
}

// URL resolves the endpoint based on the selected environment.
func (c *Config) URL() string {
	if c.Environment == EnvPROD {
		return "http://10.229.16.30:9821/BCCSGateway/BCCSGateway" // Using primary Prod IP
	}
	// Default to UAT
	return "http://10.229.16.29:8520/BCCSGateway/BCCSGateway"
}

// Validate ensures all required configuration is present before starting.
func (c *Config) Validate() error {
	if c.PartnerCode == "" {
		return errors.New("emola config: PartnerCode is required")
	}
	if c.PartnerKey == "" {
		return errors.New("emola config: PartnerKey is required")
	}
	if c.Username == "" {
		return errors.New("emola config: Username is required")
	}
	if c.Password == "" {
		return errors.New("emola config: Password is required")
	}
	if c.Environment != EnvPROD && c.Environment != EnvUAT {
		return errors.New("emola config: Environment must be EnvUAT or EnvPROD")
	}
	return nil
}
