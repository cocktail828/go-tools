package configor_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/cocktail828/go-tools/configor"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type Anonymous struct {
	Description string
}

type testConfig struct {
	APPName string `default:"configor" json:",omitempty"`
	Hosts   []string
	DB      struct {
		Name     string
		User     string `default:"root"`
		Password string `env:"DBPassword" validate:"required"`
		Port     uint   `default:"3306" json:",omitempty"`
		SSL      bool   `default:"true" json:",omitempty"`
	}
	Contacts []struct {
		Name  string
		Email string `validate:"required"`
	}
	Anonymous `anonymous:"true"`
	private   string
}

func generateDefaultConfig() testConfig {
	return testConfig{
		APPName: "configor",
		Hosts:   []string{"http://example.org", "http://configor.me"},
		DB: struct {
			Name     string
			User     string `default:"root"`
			Password string `env:"DBPassword" validate:"required"`
			Port     uint   `default:"3306" json:",omitempty"`
			SSL      bool   `default:"true" json:",omitempty"`
		}{
			Name:     "configor",
			User:     "configor",
			Password: "configor",
			Port:     3306,
			SSL:      true,
		},
		Contacts: []struct {
			Name  string
			Email string `validate:"required"`
		}{
			{
				Name:  "example",
				Email: "wosmvp@gmail.com",
			},
		},
		Anonymous: Anonymous{
			Description: "This is an anonymous embedded struct whose environment variables should NOT include 'ANONYMOUS'",
		},
	}
}

func TestLoadNormaltestConfig(t *testing.T) {
	config := generateDefaultConfig()
	if bytes, err := json.Marshal(config); err == nil {
		if file, err := os.CreateTemp("/tmp", "configor-*.json"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)

			var result testConfig
			if err := configor.LoadFile(&result, file.Name()); err != nil {
				t.Errorf("configor.LoadFile fail for %v", err)
			}
			assert.Equal(t, result, config)
			if !reflect.DeepEqual(result, config) {
				t.Errorf("result should equal to original configuration")
			}
		}
	} else {
		t.Errorf("failed to marshal config")
	}
}

func TestLoadtestConfigFromTomlWithExtension(t *testing.T) {
	var (
		config = generateDefaultConfig()
		buffer bytes.Buffer
	)

	if err := toml.NewEncoder(&buffer).Encode(config); err == nil {
		if file, err := os.CreateTemp("/tmp", "configor-*.toml"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(buffer.Bytes())

			var result testConfig
			if err := configor.LoadFile(&result, file.Name()); err != nil {
				t.Errorf("configor.LoadFile fail for %v", err)
			}
			if !reflect.DeepEqual(result, config) {
				t.Errorf("result should equal to original configuration")
			}
		}
	} else {
		t.Errorf("failed to marshal config")
	}
}

func TestLoadtestConfigFromTomlWithoutExtension(t *testing.T) {
	var (
		config = generateDefaultConfig()
		buffer bytes.Buffer
	)

	if err := toml.NewEncoder(&buffer).Encode(config); err == nil {
		if file, err := os.CreateTemp("/tmp", "configor-*.toml"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(buffer.Bytes())

			var result testConfig
			if err := configor.LoadFile(&result, file.Name()); err != nil {
				t.Errorf("configor.LoadFile fail for %v", err)
			}
			if !reflect.DeepEqual(result, config) {
				t.Errorf("result should equal to original configuration")
			}
		}
	} else {
		t.Errorf("failed to marshal config")
	}
}

func TestDefaultValue(t *testing.T) {
	config := generateDefaultConfig()
	config.APPName = ""
	config.DB.Port = 0
	config.DB.SSL = false

	if bytes, err := json.Marshal(config); err == nil {
		if file, err := os.CreateTemp("/tmp", "configor-*.json"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)

			var result testConfig
			if err := configor.LoadFile(&result, file.Name()); err != nil {
				t.Errorf("configor.LoadFile fail for %v", err)
			}

			if !reflect.DeepEqual(result, generateDefaultConfig()) {
				t.Errorf("result should be set default value correctly")
			}
		}
	} else {
		t.Errorf("failed to marshal config")
	}
}

func TestMissingRequiredValue(t *testing.T) {
	config := generateDefaultConfig()
	config.DB.Password = ""

	if bytes, err := json.Marshal(config); err == nil {
		if file, err := os.CreateTemp("/tmp", "configor-*.json"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)

			var result testConfig
			if err := configor.LoadFile(&result, file.Name()); err == nil {
				t.Errorf("configor.LoadFile should fail")
			}
		}
	} else {
		t.Errorf("failed to marshal config")
	}
}

func TestUnmatchedKeyInTomltestConfigFile(t *testing.T) {
	type configStruct struct {
		Name string
	}
	type configFile struct {
		Name string
		Test string
	}
	config := configFile{Name: "test", Test: "ATest"}

	file, err := os.CreateTemp("/tmp", "configor-*.toml")
	if err != nil {
		t.Fatal("Could not create temp file")
	}
	defer os.Remove(file.Name())
	defer file.Close()

	filename := file.Name()

	if err := toml.NewEncoder(file).Encode(config); err != nil {
		t.Errorf("failed to marshal config")
	}

	var result configStruct
	// Do not return error when there are unmatched keys but ErrorOnUnmatchedKeys is false
	if err := configor.LoadFile(&result, filename); err != nil {
		t.Errorf("Should NOT get error when loading configuration with extra keys. Error: %v", err)
	}
}

func TestUnmatchedKeyInYamltestConfigFile(t *testing.T) {
	type configStruct struct {
		Name string
	}
	type configFile struct {
		Name string
		Test string
	}
	config := configFile{Name: "test", Test: "ATest"}

	file, err := os.CreateTemp("/tmp", "configor-*.yml")
	if err != nil {
		t.Fatal("Could not create temp file")
	}

	defer os.Remove(file.Name())
	defer file.Close()

	filename := file.Name()

	if data, err := yaml.Marshal(config); err == nil {
		file.WriteString(string(data))
	} else {
		t.Errorf("failed to marshal config")
	}

	var result configStruct

	// Do not return error when there are unmatched keys but ErrorOnUnmatchedKeys is false
	if err := configor.LoadFile(&result, filename); err != nil {
		t.Errorf("Should NOT get error when loading configuration with extra keys. Error: %v", err)
	}
}

func TestUnmatchedKeyInJsonConfigFile(t *testing.T) {
	type configStruct struct {
		Name string
	}
	type configFile struct {
		Name string
		Test string
	}
	config := configFile{Name: "test", Test: "ATest"}

	file, err := os.CreateTemp("/tmp", "configor-*.json")
	if err != nil {
		t.Fatal("Could not create temp file")
	}
	defer os.Remove(file.Name())
	defer file.Close()

	if err := json.NewEncoder(file).Encode(config); err != nil {
		t.Errorf("failed to marshal config")
	}

	var result configStruct
	// Do not return error when there are unmatched keys but ErrorOnUnmatchedKeys is false
	if err := configor.LoadFile(&result, file.Name()); err != nil {
		t.Errorf("Should NOT get error when loading configuration with extra keys. Error: %v", err)
	}
}

func TestOverwritetestConfigurationWithEnvironmentWithDefaultPrefix(t *testing.T) {
	config := generateDefaultConfig()

	if bytes, err := json.Marshal(config); err == nil {
		if file, err := os.CreateTemp("/tmp", "configor-*.json"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)
			var result testConfig
			os.Setenv("CONFIGOR_ENV_PREFIX", "CONFIGOR")
			os.Setenv("CONFIGOR_APPNAME", "config2")
			os.Setenv("CONFIGOR_HOSTS", "- http://example.org\n- http://configor.me")
			os.Setenv("CONFIGOR_DB_NAME", "db_name")
			defer os.Setenv("CONFIGOR_APPNAME", "")
			defer os.Setenv("CONFIGOR_HOSTS", "")
			defer os.Setenv("CONFIGOR_DB_NAME", "")
			if err := configor.LoadFile(&result, file.Name()); err != nil {
				t.Errorf("configor.LoadFile fail for %v", err)
			}

			var defaultConfig = generateDefaultConfig()
			defaultConfig.APPName = "config2"
			defaultConfig.Hosts = []string{"http://example.org", "http://configor.me"}
			defaultConfig.DB.Name = "db_name"
			if !reflect.DeepEqual(result, defaultConfig) {
				t.Errorf("result should equal to original configuration")
			}
		}
	}
}

func TestOverwritetestConfigurationWithEnvironment(t *testing.T) {
	config := generateDefaultConfig()

	if bytes, err := json.Marshal(config); err == nil {
		if file, err := os.CreateTemp("/tmp", "configor-*.json"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)
			var result testConfig
			os.Setenv("CONFIGOR_ENV_PREFIX", "app")
			os.Setenv("APP_APPNAME", "config2")
			os.Setenv("APP_DB_NAME", "db_name")
			defer os.Setenv("APP_APPNAME", "")
			defer os.Setenv("APP_DB_NAME", "")
			if err := configor.LoadFile(&result, file.Name()); err != nil {
				t.Errorf("configor.LoadFile fail for %v", err)
			}

			var defaultConfig = generateDefaultConfig()
			defaultConfig.APPName = "config2"
			defaultConfig.DB.Name = "db_name"
			if !reflect.DeepEqual(result, defaultConfig) {
				t.Errorf("result should equal to original configuration")
			}
		}
	}
}

func TestOverwritetestConfigurationWithEnvironmentThatSetBytestConfig(t *testing.T) {
	config := generateDefaultConfig()

	if bytes, err := json.Marshal(config); err == nil {
		if file, err := os.CreateTemp("/tmp", "configor-*.json"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)
			os.Setenv("CONFIGOR_ENV_PREFIX", "app1")
			os.Setenv("APP1_APPName", "config2")
			os.Setenv("APP1_DB_Name", "db_name")
			defer os.Setenv("APP1_APPName", "")
			defer os.Setenv("APP1_DB_Name", "")

			var result testConfig
			var Configor = &configor.Configor{EnvPrefix: "APP1"}
			Configor.LoadFile(&result, file.Name())

			var defaultConfig = generateDefaultConfig()
			defaultConfig.APPName = "config2"
			defaultConfig.DB.Name = "db_name"
			if !reflect.DeepEqual(result, defaultConfig) {
				t.Errorf("result should equal to original configuration")
			}
		}
	}
}

func TestResetPrefixToBlank(t *testing.T) {
	config := generateDefaultConfig()

	if bytes, err := json.Marshal(config); err == nil {
		if file, err := os.CreateTemp("/tmp", "configor-*.json"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)
			var result testConfig
			os.Setenv("CONFIGOR_ENV_PREFIX", "")
			os.Setenv("APPNAME", "config2")
			os.Setenv("DB_NAME", "db_name")
			defer os.Setenv("APPNAME", "")
			defer os.Setenv("DB_NAME", "")
			if err := configor.LoadFile(&result, file.Name()); err != nil {
				t.Errorf("configor.LoadFile fail for %v", err)
			}

			var defaultConfig = generateDefaultConfig()
			defaultConfig.APPName = "config2"
			defaultConfig.DB.Name = "db_name"
			if !reflect.DeepEqual(result, defaultConfig) {
				t.Errorf("result should equal to original configuration")
			}
		}
	}
}

func TestResetPrefixToBlank2(t *testing.T) {
	config := generateDefaultConfig()

	if bytes, err := json.Marshal(config); err == nil {
		if file, err := os.CreateTemp("/tmp", "configor-*.json"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)
			var result testConfig
			os.Setenv("APPName", "config2")
			os.Setenv("DB_Name", "db_name")
			defer os.Setenv("APPName", "")
			defer os.Setenv("DB_Name", "")
			if err := configor.LoadFile(&result, file.Name()); err != nil {
				t.Errorf("configor.LoadFile fail for %v", err)
			}

			var defaultConfig = generateDefaultConfig()
			defaultConfig.APPName = "config2"
			defaultConfig.DB.Name = "db_name"
			if !reflect.DeepEqual(result, defaultConfig) {
				t.Errorf("result should equal to original configuration")
			}
		}
	}
}

func TestReadFromEnvironmentWithSpecifiedEnvName(t *testing.T) {
	config := generateDefaultConfig()

	if bytes, err := json.Marshal(config); err == nil {
		if file, err := os.CreateTemp("/tmp", "configor-*.json"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)
			var result testConfig
			os.Setenv("DBPassword", "db_password")
			defer os.Setenv("DBPassword", "")
			if err := configor.LoadFile(&result, file.Name()); err != nil {
				t.Errorf("configor.LoadFile fail for %v", err)
			}

			var defaultConfig = generateDefaultConfig()
			defaultConfig.DB.Password = "db_password"
			if !reflect.DeepEqual(result, defaultConfig) {
				t.Errorf("result should equal to original configuration")
			}
		}
	}
}

func TestAnonymousStruct(t *testing.T) {
	config := generateDefaultConfig()

	if bytes, err := json.Marshal(config); err == nil {
		if file, err := os.CreateTemp("/tmp", "configor-*.json"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)
			var result testConfig
			os.Setenv("CONFIGOR_ENV_PREFIX", "CONFIGOR")
			os.Setenv("CONFIGOR_DESCRIPTION", "environment description")
			defer os.Setenv("CONFIGOR_DESCRIPTION", "")
			if err := configor.LoadFile(&result, file.Name()); err != nil {
				t.Errorf("configor.LoadFile fail for %v", err)
			}

			var defaultConfig = generateDefaultConfig()
			defaultConfig.Anonymous.Description = "environment description"
			assert.Equal(t, result, defaultConfig)
			if !reflect.DeepEqual(result, defaultConfig) {
				t.Errorf("result should equal to original configuration")
			}
		}
	}
}

type slicetestConfig struct {
	Test1 int
	Test2 []struct {
		Test2Ele1 int
		Test2Ele2 int
	}
}

func TestSliceFromEnv(t *testing.T) {
	var tc = slicetestConfig{
		Test1: 1,
		Test2: []struct {
			Test2Ele1 int
			Test2Ele2 int
		}{
			{
				Test2Ele1: 1,
				Test2Ele2: 2,
			},
			{
				Test2Ele1: 3,
				Test2Ele2: 4,
			},
		},
	}

	var result slicetestConfig
	os.Setenv("CONFIGOR_ENV_PREFIX", "CONFIGOR")
	os.Setenv("CONFIGOR_TEST1", "1")
	os.Setenv("CONFIGOR_TEST2_0_TEST2ELE1", "1")
	os.Setenv("CONFIGOR_TEST2_0_TEST2ELE2", "2")

	os.Setenv("CONFIGOR_TEST2_1_TEST2ELE1", "3")
	os.Setenv("CONFIGOR_TEST2_1_TEST2ELE2", "4")
	if err := configor.Load(&result); err != nil {
		t.Fatalf("configor.Load from env err:%v", err)
	}

	if !reflect.DeepEqual(result, tc) {
		t.Fatalf("unexpected result:%+v", result)
	}
}

func TestConfigFromEnv(t *testing.T) {
	type config struct {
		LineBreakString string `required:"true"`
		Count           int64
		Slient          bool
	}

	cfg := &config{}

	os.Setenv("CONFIGOR_ENV_PREFIX", "CONFIGOR")
	os.Setenv("CONFIGOR_LineBreakString", "Line one\nLine two\nLine three\nAnd more lines")
	os.Setenv("CONFIGOR_Slient", "1")
	os.Setenv("CONFIGOR_Count", "10")
	if err := configor.Load(cfg); err != nil {
		t.Fatalf("configor.Load err:%v", err)
	}

	if os.Getenv("CONFIGOR_LineBreakString") != cfg.LineBreakString {
		t.Error("Failed to load value has line break from env")
	}

	if !cfg.Slient {
		t.Error("Failed to load bool from env")
	}

	if cfg.Count != 10 {
		t.Error("Failed to load number from env")
	}
}

type Menu struct {
	Key      string `json:"key" yaml:"key"`
	Name     string `json:"name" yaml:"name"`
	Icon     string `json:"icon" yaml:"icon"`
	Children []Menu `json:"children" yaml:"children"`
}

type MenuList struct {
	Top []Menu `json:"top"  yaml:"top"`
}

func TestLoadNestedConfig(t *testing.T) {
	adminConfig := MenuList{}
	if err := configor.LoadFile(&adminConfig, "test/admin.yml"); err != nil {
		t.Fatalf("configor.LoadFile err:%v", err)
	}
}

func TestLoad_FS(t *testing.T) {
	type testEmbedConfig struct {
		Foo string
	}
	var result testEmbedConfig
	if err := configor.LoadFile(&result, "test/config.yaml"); err != nil {
		t.Fatalf("configor.LoadFile err:%v", err)
	}
	if result.Foo != "bar" {
		t.Error("expected to have foo: bar in config")
	}
}

func TestValidateDefault(t *testing.T) {
	type Obj struct {
		V bool `toml:"v" default:"false" validate:"required"`
	}
	o := Obj{}
	assert.Equal(t, nil, configor.Load(&o, []byte(`v=false`)).Error())
}
