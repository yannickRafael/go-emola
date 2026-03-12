package config

import (
	"errors"
	"os"
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
	// 1. Check for a global override
	if url := os.Getenv("EMOLA_BASE_URL"); url != "" {
		return url
	}

	// 2. Resolve based on Environment
	if c.Environment == EnvPROD {
		return os.Getenv("EMOLA_PROD_URL")
	}

	return os.Getenv("EMOLA_UAT_URL")
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

	// Ensure a URL is available
	if c.URL() == "" {
		if c.Environment == EnvPROD {
			return errors.New("emola config: EMOLA_PROD_URL (or EMOLA_BASE_URL) is required for Production")
		}
		return errors.New("emola config: EMOLA_UAT_URL (or EMOLA_BASE_URL) is required for UAT")
	}

	return nil
}
