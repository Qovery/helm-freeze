package cfg

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func ValidateConfig(configFile string) (Config, error) {
	cfg := Config{}

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Error while reading config file %s: #%v ", configFile, err)
		return Config{}, err
	}

	err = yaml.Unmarshal([]byte(data), &cfg)
	if err != nil {
		fmt.Printf("Configuration file error: %v", err)
		return Config{}, err
	}

	// Todo: Ensure content is coherent
	//for _, chart := range cfg.Charts {
	//	fmt.Println(chart["name"])
	//}

	return cfg, nil
}

// Todo: find duplicates charts, repos and destinations
// Todo: check default destination is present