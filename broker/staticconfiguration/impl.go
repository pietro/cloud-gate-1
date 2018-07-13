package staticconfiguration

import (
	"errors"
	"os"

	"github.com/Symantec/cloud-gate/lib/constants"
	"gopkg.in/yaml.v2"
)

func LoadVerifyConfigFile(configFilename string) (*StaticConfiguration, error) {
	var config StaticConfiguration
	if _, err := os.Stat(configFilename); os.IsNotExist(err) {
		err = errors.New("mising config file failure")
		return nil, err
	}
	source, err := os.Open(configFilename)
	if err != nil {
		return nil, err

	}
	err = yaml.NewDecoder(source).Decode(&config)
	if err != nil {
		return nil, err
	}
	// setup defaults
	if config.Base.StatusPort == 0 {
		config.Base.StatusPort = constants.DefaultStatusPort
	}
	if config.Base.ServicePort == 0 {
		config.Base.ServicePort = constants.DefaultServicePort
	}
	if len(config.Base.AccountConfigurationUrl) == 0 {
		config.Base.AccountConfigurationUrl =
			constants.DefaultAccountConfigurationUrl
	}
	if config.Base.AccountConfigurationCheckInterval == 0 {
		config.Base.AccountConfigurationCheckInterval =
			constants.DefaultAccountConfigurationCheckInterval
	}

	return &config, nil
}