package config

import (
	"errors"
	"io/ioutil"
	"regexp"

	"gopkg.in/yaml.v2"
)

var (
	errEmtpyContainer = errors.New("One of container_id, name or image should provide in container")
	errEmtpyLogFormat = errors.New("log_format should be regex of valid log")
)

type PrometheusMetricType string

const (
	Counter   PrometheusMetricType = "counter"
	Gauge     PrometheusMetricType = "gauge"
	Histogram PrometheusMetricType = "histogram"
	Summary   PrometheusMetricType = "summary"
)

type Container struct {
	ContainerID string `yaml:"container_id"`
	Name        string `yaml:"name"`
	Image       string `yaml:"image"`
}

type Config struct {
	Logprom []struct {
		Container         Container                       `yaml:"container,flow"`
		LogFormat         string                          `yaml:"log_format"`
		PrometheusMapping map[string]PrometheusMetricType `yaml:"prometheus_mapping,flow"`
	} `yaml:"logprom,flow"`
}

func LoadConfig(path string, c *Config) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = unmarshal(data, c)
	if err != nil {
		return err
	}
	err = validateConfig(c)
	if err != nil {
		return err
	}
	return nil
}

// func validatePromMapping()
func unmarshal(data []byte, c *Config) error {
	err := yaml.Unmarshal(data, c)
	return err
}

func validateConfig(c *Config) error {
	for _, l := range c.Logprom {
		if err := validateContainer(&l.Container); err != nil {
			return err
		}
		if err := validateLogFormat(&l.LogFormat); err != nil {
			return err
		}
		// TODO: improve validation
		// if err := validatePrometheusMapping(&l.LogFormat); err != nil {
		// 	return err
		// }

	}
	return nil
}

func validateContainer(c *Container) error {
	if c.ContainerID != "" {
		_, err := regexp.Compile(c.ContainerID)
		return err
	}
	if c.Name != "" {
		_, err := regexp.Compile(c.Name)
		return err
	}
	if c.Image != "" {
		_, err := regexp.Compile(c.Image)
		return err
	}
	return errEmtpyContainer
}

func validateLogFormat(l *string) error {
	if *l == "" {
		return errEmtpyLogFormat
	}
	_, err := regexp.Compile(*l)
	return err
}
