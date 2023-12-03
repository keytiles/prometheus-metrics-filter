package conf

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// All configurations
var All AllConfig

// InitConfig Initilization for configuration file
func InitConfig(path string) error {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("yamlFile.Get err   #%v ", err)
		return err
	}
	err = yaml.Unmarshal(yamlFile, &All)
	if err != nil {
		fmt.Printf("Unmarshal: %v", err)
		return err
	}
	return nil
}

// AllConfig all config struct
type AllConfig struct {
	HttpService HttpServiceConfig    `yaml:"httpService"`
	ProxyRules  map[string]ProxyRule `yaml:"proxyRules"`
}

// HttpServiceConfig influxdb config struct
type HttpServiceConfig struct {
	Address                string `yaml:"address"`
	Port                   string `yaml:"port"`
	DefaultProxyRule       string `yaml:"defaultProxyRule"`
	DefaultMetricsFetchUrl string `yaml:"defaultMetricsFetchUrl"`
}

type ProxyRule struct {
	Remove  []LineMatchRule `yaml:"remove"`
	Include []LineMatchRule `yaml:"include"`
}

type LineMatchRule struct {
	Regexp string          `yaml:"regexp"`
	Negate bool            `yaml:"negate"`
	And    []LineMatchRule `yaml:"and"`
	Or     []LineMatchRule `yaml:"or"`
}
