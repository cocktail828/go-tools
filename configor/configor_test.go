package configor

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type DB struct {
	Name     string        `env:"DB_NAME" default:"name"`
	User     string        `env:"DB_USER" default:"root"`
	Password string        `env:"DB_PASSWORD" validate:"required"`
	Port     uint          `env:"DB_PORT" default:"3306"`
	SSL      bool          `env:"DB_SSL" default:"true"`
	Timeout  time.Duration `env:"DB_TIMEOUT" default:"5s"`
}

type Config struct {
	APPName string `env:"APPNAME" default:"configor"`
	DB      DB
}

type MMP map[string]string

func (m MMP) SetEnv() {
	for k, v := range m {
		os.Setenv(k, v)
	}
}

func (m MMP) ResetEnv() {
	for k := range m {
		os.Setenv(k, "")
	}
}

func TestDifferentEnvPrefixes(t *testing.T) {
	tests := []struct {
		name        string
		envPrefix   string
		envVars     MMP
		expectedApp string
	}{
		{
			name:      "prefix_APP",
			envPrefix: "APP",
			envVars: MMP{
				"APP_APPNAME":   "app_prefix_app",
				"OTHER_APPNAME": "should_be_ignored",
			},
			expectedApp: "app_prefix_app",
		},
		{
			name:      "prefix_CUSTOM",
			envPrefix: "CUSTOM",
			envVars: MMP{
				"CUSTOM_APPNAME": "custom_prefix_app",
				"APP_APPNAME":    "should_be_ignored",
			},
			expectedApp: "custom_prefix_app",
		},
		{
			name:      "empty_prefix",
			envPrefix: "",
			envVars: MMP{
				"APPNAME":          "empty_prefix_app",
				"CONFIGOR_APPNAME": "should_be_ignored",
			},
			expectedApp: "empty_prefix_app",
		},
	}

	cfgor := &Configor{LoadEnv: true}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.envVars.SetEnv()
			defer tt.envVars.ResetEnv()

			cfgor.EnvPrefix = tt.envPrefix
			cfg := Config{}
			if err := cfgor.Load(&cfg); err != nil {
				t.Fatalf("test %q fail: %v", tt.name, err)
				return
			}
			assert.Equal(t, tt.expectedApp, cfg.APPName, "EnvPrefix handle env fail")
		})
	}
}

func TestLoadConfigurationWithLoadEnvFalse(t *testing.T) {
	m := MMP{
		"CONFIGOR_APPNAME":     "env_overridden_app",
		"CONFIGOR_DB_NAME":     "env_overridden_db",
		"CONFIGOR_DB_USER":     "env_user",
		"CONFIGOR_DB_PORT":     "6543",
		"CONFIGOR_DB_PASSWORD": "env_password",
		"CONFIGOR_DB_TIMEOUT":  "10s",
	}
	m.SetEnv()
	defer m.ResetEnv()

	c := Configor{
		LoadEnv:   false,
		EnvPrefix: "CONFIGOR",
	}

	cfg := Config{}
	if err := c.Load(&cfg); err != nil {
		assert.FailNow(t, "failed to load configuration", err.Error())
	}

	assert.Equal(t, cfg.APPName, "configor", "AppName should remain default when LoadEnv=false")
	assert.Equal(t, cfg.DB.Name, "name", "DB.Name should remain default when LoadEnv=false")
	assert.Equal(t, cfg.DB.User, "root", "DB.User should remain default when LoadEnv=false")
	assert.EqualValues(t, cfg.DB.Port, 3306, "DB.Port should remain default when LoadEnv=false")
	assert.Equal(t, cfg.DB.Password, "", "DB.Password should remain default when LoadEnv=false")
	assert.Equal(t, cfg.DB.Timeout, time.Duration(5)*time.Second, "DB.Timeout should remain default when LoadEnv=false")
}

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

	c := Configor{
		LoadEnv:   true,
		EnvPrefix: "",
	}

	var cfg xConfig
	if err := c.bindEnv(&cfg); err != nil {
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

	c := Configor{}
	if err := c.bindEnv(&nilPtrTest); err != nil {
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

	c.LoadEnv = true
	if err := c.bindEnv(&envPtrTest); err != nil {
		assert.FailNow(t, "failed to bind env to pointer fields", err.Error())
	}
	assert.Equal(t, "env_name", *envPtrTest.Name)
	assert.Equal(t, 30, *envPtrTest.Age)
	assert.True(t, *envPtrTest.Flag)
}
