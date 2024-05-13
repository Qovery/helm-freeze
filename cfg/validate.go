package cfg

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

func ValidateConfig(configFile string) (Config, error) {
	cfg := Config{}
	allReposName := make(map[string]bool)
	allReposUrl := make(map[string]bool)
	allDestinationsName := make(map[string]bool)
	allDestinationsPath := make(map[string]bool)

	data, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Error while reading config file %s: #%v ", configFile, err)
		return Config{}, err
	}

	err = yaml.Unmarshal([]byte(data), &cfg)
	if err != nil {
		fmt.Printf("Configuration file error: %v", err)
		return Config{}, err
	}

	// Ensure content is coherent
	for _, chart := range cfg.Charts {
		if _, ok := chart["name"]; !ok {
			return cfg, fmt.Errorf("name is missing in charts config for this element: %v\n", chart)
		}
		if _, ok := chart["version"]; !ok {
			return cfg, fmt.Errorf("version is missing in %s charts config\n", chart["name"])
		}
	}

	for _, repo := range cfg.Repos {
		if _, ok := repo["name"]; !ok {
			return cfg, fmt.Errorf("name is missing in repos config for this element: %v\n", repo)
		}
		if _, ok := repo["url"]; !ok {
			return cfg, fmt.Errorf("url is missing in %s repos config\n", repo["name"])
		}
	}

	for _, dest := range cfg.Destinations {
		if _, ok := dest["name"]; !ok {
			return cfg, fmt.Errorf("name is missing in destinations config for this element: %v\n", dest)
		}
		if _, ok := dest["path"]; !ok {
			return cfg, fmt.Errorf("path is missing in %s destinations config\n", dest["name"])
		}
	}

	// Find duplicates
	for _, repo := range cfg.Repos {
		if _, ok := allReposName[repo["name"]]; ok {
			return cfg, fmt.Errorf("Duplicate repo name found: %s\n", repo["name"])
		} else {
			allReposName[repo["name"]] = true
		}

		if _, ok := allReposUrl[repo["url"]]; ok {
			return cfg, fmt.Errorf("Duplicate url found: %s\n", repo["url"])
		} else {
			allReposUrl[repo["url"]] = true
		}
	}

	for _, dest := range cfg.Destinations {
		if _, ok := allDestinationsName[dest["name"]]; ok {
			return cfg, fmt.Errorf("Duplicate destination name found: %s\n", dest["name"])
		} else {
			allDestinationsName[dest["name"]] = true
		}

		if _, ok := allDestinationsPath[dest["path"]]; ok {
			return cfg, fmt.Errorf("Duplicate destination path found: %s\n", dest["path"])
		} else {
			allDestinationsPath[dest["path"]] = true
		}
	}

	// check if defaults elements are present
	if _, ok := allDestinationsName["default"]; !ok {
		return cfg, errors.New("missing default destination")
	}

	if _, ok := allReposName["stable"]; !ok {
		return cfg, errors.New("missing stable repository")
	}

	return cfg, nil
}
