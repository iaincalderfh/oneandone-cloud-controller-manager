package oneandone

import (
	"fmt"
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type config struct {
}

// readConfig consumes the config Reader and constructs a Config object.
func readConfig(r io.Reader) (*config, error) {
	cfg := &config{}
	if r == nil {
		return cfg, nil
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("Error reading cloud-provider config: %v", err)
	}

	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling cloud-provider config: %v", err)
	}

	return cfg, nil
}
