package oneandone

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type firewallConfig struct {
	Name string `yaml:"name"`
}

type config struct {
	Firewall firewallConfig `yaml:"firewall"`
}

// readConfig consumes the config Reader and constructs a Config object.
func readConfig(r io.Reader) (*config, error) {
	if r == nil {
		return nil, errors.New("no cloud-provider config file given")
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("Error reading cloud-provider config: %v", err)
	}

	cfg := &config{}
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling cloud-provider config: %v", err)
	}

	return cfg, nil
}
