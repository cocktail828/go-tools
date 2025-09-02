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

type TestStructWithPointer struct {
	Name *string `env:"NAME" default:"default_name"`
	Age  *int    `env:"AGE" default:"20"`
	Flag *bool   `env:"FLAG" default:"true"`
}

func TestBindEnvWithPointerFields(t *testing.T) {
	var nilPtrTest TestStructWithPointer
	assert.Nil(t, nilPtrTest.Name)
	assert.Nil(t, nilPtrTest.Age)
	assert.Nil(t, nilPtrTest.Flag)

	if err := BindEnv(&nilPtrTest, WithSkipEnv(true)); err != nil {
		assert.FailNow(t, "failed to bind nil pointer fields", err.Error())
	}
	assert.Equal(t, "default_name", *nilPtrTest.Name)
	assert.Equal(t, 20, *nilPtrTest.Age)
	assert.True(t, *nilPtrTest.Flag)

	envPtrTest := TestStructWithPointer{
		Name: new(string),
		Age:  new(int),
		Flag: new(bool),
	}
	m := MMP{
		"NAME": "env_name",
		"AGE":  "30",
		"FLAG": "true",
	}
	m.SetEnv()
	defer m.ResetEnv()

	if err := BindEnv(&envPtrTest); err != nil {
		assert.FailNow(t, "failed to bind env to pointer fields", err.Error())
	}
	assert.Equal(t, "env_name", *envPtrTest.Name)
	assert.Equal(t, 30, *envPtrTest.Age)
	assert.True(t, *envPtrTest.Flag)
}
