package configor

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// unmatchedTomlKeysError errors are returned by the Load function when
// ErrorOnUnmatchedKeys is set to true and there are unmatched keys in the input
// toml dst file. The string returned by Error() contains the names of the
// missing keys.
type unmatchedTomlKeysError struct {
	Keys []toml.Key
}

func (e *unmatchedTomlKeysError) Error() string {
	return fmt.Sprintf("There are keys in the dst file that do not match any field in the given struct: %v", e.Keys)
}

func processFile(dst interface{}, file string, errorOnUnmatchedKeys bool) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	switch {
	case strings.HasSuffix(file, ".yaml") || strings.HasSuffix(file, ".yml"):
		if errorOnUnmatchedKeys {
			return yaml.UnmarshalStrict(data, dst)
		}
		return yaml.Unmarshal(data, dst)
	case strings.HasSuffix(file, ".toml"):
		return unmarshalToml(data, dst, errorOnUnmatchedKeys)
	case strings.HasSuffix(file, ".json"):
		return unmarshalJSON(data, dst, errorOnUnmatchedKeys)
	default:
		if err := unmarshalToml(data, dst, errorOnUnmatchedKeys); err == nil {
			return nil
		} else if errUnmatchedKeys, ok := err.(*unmatchedTomlKeysError); ok {
			return errUnmatchedKeys
		}

		if err := unmarshalJSON(data, dst, errorOnUnmatchedKeys); err == nil {
			return nil
		} else if strings.Contains(err.Error(), "json: unknown field") {
			return err
		}

		var yamlError error
		if errorOnUnmatchedKeys {
			yamlError = yaml.UnmarshalStrict(data, dst)
		} else {
			yamlError = yaml.Unmarshal(data, dst)
		}

		if yamlError == nil {
			return nil
		} else if yErr, ok := yamlError.(*yaml.TypeError); ok {
			return yErr
		}
		return errors.Errorf("failed to decode %v", file)
	}
}

func unmarshalToml(data []byte, dst interface{}, errorOnUnmatchedKeys bool) error {
	metadata, err := toml.Decode(string(data), dst)
	if err == nil && len(metadata.Undecoded()) > 0 && errorOnUnmatchedKeys {
		return &unmatchedTomlKeysError{Keys: metadata.Undecoded()}
	}
	return err
}

func unmarshalTomlString(data string, dst interface{}, errorOnUnmatchedKeys bool) error {
	metadata, err := toml.Decode(data, dst)
	if err == nil && len(metadata.Undecoded()) > 0 && errorOnUnmatchedKeys {
		return &unmatchedTomlKeysError{Keys: metadata.Undecoded()}
	}
	return err
}
