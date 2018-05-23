package oneandone

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type FirewallConfig struct {
	Name string `yaml:"name"`
}

type Config struct {
	Firewall FirewallConfig `yaml:"firewall"`
}

// ReadConfig consumes the config Reader and constructs a Config object.
func ReadConfig(r io.Reader) (*Config, error) {
	if r == nil {
		return nil, errors.New("no cloud-provider config file given")
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("Error reading cloud-provider config: %v", err)
	}

	cfg := &Config{}
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling cloud-provider config: %v", err)
	}

	return cfg, nil
}
