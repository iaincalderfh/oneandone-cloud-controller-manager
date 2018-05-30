package oneandone

import (
	"strings"
	"testing"
)

const validConfig = `
firewall:
  name: spc-1and1-fw
`

func TestValidConfigFirewall(t *testing.T) {
	config, err := readConfig(strings.NewReader(validConfig))
	if err != nil {
		t.Fatal(err)
	}
	expected := "spc-1and1-fw"
	if config.Firewall.Name != expected {
		t.Errorf("expected firewall name %s but got %s", expected, config.Firewall.Name)
	}
}
