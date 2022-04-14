package config

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// YamlConfig is used to handle yaml config file.
type YamlConfig struct {
	m map[string]any
}

// Create YamlConfig.
func NewYamlConfig() *YamlConfig {
	return &YamlConfig{
		m: make(map[string]any),
	}
}

// For printable.
func (c *YamlConfig) String() string {
	return fmt.Sprintf("YamlConfig: {%v}", c.m)
}

// Load and parse yaml from file.
func (c *YamlConfig) UnmarshalFromFile(filepath string) error {
	var f *os.File
	var err error

	f, err = os.Open(filepath)
	if err != nil {
		return errors.Wrap(err, "Open file fail")
	}

	return c.UnmarshalFromReader(f)
}

// Load and parse yaml from io.Reader.
func (c *YamlConfig) UnmarshalFromReader(reader io.Reader) error {
	var err error
	var decoder *yaml.Decoder

	decoder = yaml.NewDecoder(reader)
	err = decoder.Decode(&c.m)
	if err != nil {
		return errors.Wrap(err, "Decode fail")
	}
	return nil

}

// Get value with key, see Config.Get().
func (c *YamlConfig) Get(key string) *Value {
	var cur any = c.m
	names := strings.Split(key, ".")
	for _, name := range names {
		switch m := cur.(type) {
		case map[any]any:
			cur = m[name]
		case map[string]any:
			cur = m[name]
		default:
			cur = nil
		}
	}
	if cur == nil {
		return nil
	}
	return &Value{cur}
}
