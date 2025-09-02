package configor

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type DB struct {
	Name     string `env:"DB_NAME" default:"name"`
	User     string `env:"DB_USER" default:"root"`
	Password string `env:"DB_PASSWORD" validate:"required"`
	Port     uint   `env:"DB_PORT" default:"3306"`
	SSL      bool   `env:"DB_SSL" default:"true"`
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
}
