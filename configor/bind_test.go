package configor

import (
	"log"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type Service struct {
	Version string `env:"POLARIS_VERSION" validate:"required"`
	AppHost string `env:"APP_HOST" validate:"required"`
	AppPort string `env:"APP_PORT" validate:"required"`
}

type xConfig struct {
	Companion string  `env:"POLARIS_COMPANION" default:"companion_value" validate:"required"`
	Project   int64   `env:"POLARIS_PROJECT" validate:"required"`
	Group     float32 `env:"POLARIS_GROUP" validate:"required"`
	Service   Service
}

func TestBind(t *testing.T) {
	m := MMP{
		"POLARIS_PROJECT": "12345",
		"POLARIS_GROUP":   "3.14",
		"POLARIS_VERSION": "version_value",
		"APP_HOST":        "host_value",
		"APP_PORT":        "port_value",
	}
	m.SetEnv()
	defer m.ResetEnv()

	var cfg xConfig
	if err := BindEnv(&cfg); err != nil {
		log.Fatalf("Error binding env: %v\n", err)
	}

	assert.EqualValues(t, xConfig{
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
