package reflectx_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/cocktail828/go-tools/z/reflectx"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type FakeCloser struct{}

func (FakeCloser) Close() error { return nil }

func TestIsNil(t *testing.T) {
	var k io.Closer = func() *FakeCloser {
		return nil
	}()
	assert.Equal(t, false, k == nil)
	assert.Equal(t, true, reflectx.IsNil(k))
}

type Config struct {
	Companion string  `env:"POLARIS_COMPANION" validate:"required"`
	Project   int64   `env:"POLARIS_PROJECT" validate:"required"`
	Group     float32 `env:"POLARIS_GROUP" validate:"required"`
	Service   struct {
		Version string `env:"POLARIS_VERSION" validate:"required"`
		AppHost string `env:"APP_HOST" validate:"required"`
		AppPort string `env:"APP_PORT" validate:"required"`
	}
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
	if err := reflectx.BindEnv(&cfg); err != nil {
		fmt.Printf("Error binding env: %v\n", err)
	}

	if err := validator.New().Struct(cfg); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	} else {
		fmt.Printf("Validation passed: %+v\n", cfg)
	}
}
