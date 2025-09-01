package environ

import (
	"log"
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type Service struct {
	Version string `env:"POLARIS_VERSION" validate:"required"`
	AppHost string `env:"APP_HOST" validate:"required"`
	AppPort string `env:"APP_PORT" validate:"required"`
}

type Config struct {
	Companion string  `env:"POLARIS_COMPANION" validate:"required"`
	Project   int64   `env:"POLARIS_PROJECT" validate:"required"`
	Group     float32 `env:"POLARIS_GROUP" validate:"required"`
	Service   Service
}

func TestBindEnv(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("POLARIS_COMPANION", "companion_value")
	os.Setenv("POLARIS_PROJECT", "12345")
	os.Setenv("POLARIS_GROUP", "3.14")
	os.Setenv("POLARIS_VERSION", "version_value")
	os.Setenv("APP_HOST", "host_value")
	os.Setenv("APP_PORT", "port_value")

	var cfg Config
	if err := BindEnv(&cfg); err != nil {
		log.Fatalf("Error binding env: %v\n", err)
	}

	assert.EqualValues(t, Config{
		Companion: "companion_value",
		Project:   12345,
		Group:     3.14,
		Service: Service{
			Version: "version_value",
			AppHost: "host_value",
			AppPort: "port_value",
		},
	}, cfg)
	assert.NoError(t, validator.New().Struct(cfg))
}
