package agent

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"time"
)

type Configuration struct {
	Monitor string
	Token   string
	Global  EndpointConfiguration            `yaml:"global,omitempty"`
	Hosts   map[string]EndpointConfiguration `yaml:"hosts,omitempty"`
}

type EndpointConfiguration struct {
	Skip             bool          `yaml:"skip,omitempty"`
	Interval         time.Duration `yaml:"interval,omitempty"`
	Method           string        `yaml:"method,omitempty"`
	ValidStatusCodes []int         `yaml:"valid-status-codes,omitempty"`
}

var (
	DefaultConfiguration = Configuration{
		Global: DefaultGlobalConfiguration,
	}
	DefaultGlobalConfiguration = EndpointConfiguration{
		Interval:         5 * time.Minute,
		Method:           http.MethodGet,
		ValidStatusCodes: []int{http.StatusOK},
	}
)

func Load(r io.Reader) (Configuration, error) {
	configuration := DefaultConfiguration
	err := yaml.NewDecoder(r).Decode(&configuration)
	return configuration, err
}

func LoadFromFile(filename string) (Configuration, error) {
	f, err := os.Open(filename)
	if err != nil {
		return Configuration{}, fmt.Errorf("open: %w", err)
	}
	defer func() { _ = f.Close() }()
	return Load(f)
}
