package config

import (
	"testing"
)

var emptyConfig = []string{
	``,
	`---`,
	`---
logprom:`,
}

const (
	validConfig_0 = `---
logprom:
- container:
    container_id: akbar
    name: mirza
    image: ridi
  log_format: asdf
  prometheus_mapping:
    request_duration: gauge
`

	invalidContainerConfig_0 = `---
logprom:
- container:
    container_id: akbar
      name: mirza
      image: ridi
      log_format: asdf
    prometheus_mapping:
      request_duration: gauge`
)

func TestUnmarshal(t *testing.T) {
	c := Config{}
	err := unmarshal([]byte(validConfig_0), &c)
	if err != nil {
		t.Error("Unmarshal failed", err.Error())
	}
}

func TestValidateConfig(t *testing.T) {

}
func TestValidateContainer(t *testing.T) {
}
func TestValidateLogFormat(t *testing.T) {
}
