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
	_, err := readConfig(strings.NewReader(validConfig))
	if err != nil {
		t.Fatal(err)
	}
}
