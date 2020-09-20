package exec

import (
	"errors"
	"fmt"
	"github.com/Qovery/helm-freeze/cfg"
	"os"
	"os/exec"
	"path/filepath"
)

func AddAllRepos(config cfg.Config) error {
	fmt.Println("\n[+] Adding helm repos")
	for _, repo := range config.Repos {
		fmt.Printf(" -> %s\n", repo["name"])
		err := helmRepoAdd(repo["name"], repo["url"])
		if err != nil {
			return err
		}
	}
	return nil
}

func GetAllCharts(config cfg.Config, configPath string) error {
	// Move to folder path to correctly manage non absolute paths
	filenamePath, _ := filepath.Abs(configPath)
	absolutePath, _ := filepath.Split(filenamePath)
	err := os.Chdir(absolutePath)
	if err != nil {
		return err
	}

	// convert destinations to map for easier usage
	destinations := make(map[string]string)
	for _, dest := range config.Destinations {
		destinations[dest["name"]] = dest["path"]
	}

	// convert repos to map for easier usage
	repos := make(map[string]string)
	for _, dest := range config.Repos {
		repos[dest["name"]] = dest["path"]
	}

	// download and extract all charts
	fmt.Println("\n[+] Downloading charts")
	for _, chart := range config.Charts {
		if noSync, ok := chart["no_sync"]; ok {
			if noSync == "true" {
				continue
			}
		}

		// set default values
		chartUrl := "stable/" + chart["name"]
		destinationFolder := destinations["default"]

		// use user defined repo if specified
		if repoNameDefined, ok := chart["repo_name"]; ok {
			chartUrl = repoNameDefined + "/" + chart["name"]
		}

		// use user defined destinationFolder if specified
		if destinationIsDefined, ok := chart["dest"]; ok {
			destinationFolder = destinations[destinationIsDefined]
		}

		fmt.Printf(" -> %s %s\n", chartUrl, chart["version"])

		// move the current folder to avoid helm failure
		chartFolderDestName := destinationFolder + "/" + chart["name"]
		oldChartFolderDestName := chartFolderDestName + ".old"

		chartExists := false
		if _, err := os.Stat(chartFolderDestName); !os.IsNotExist(err) {
			chartExists = true
		}

		if chartExists {
			err := os.Rename(chartFolderDestName, oldChartFolderDestName)
			if err != nil {
				return err
			}
		}

		err = helmDownload(chartUrl, chart["version"], destinationFolder)
		if err != nil {
			if chartExists {
				// restore old chart on failure
				_ = os.Rename(oldChartFolderDestName, chartFolderDestName)
			}
			return err
		}

		// finally remove the old chart
		if chartExists {
			err = os.RemoveAll(oldChartFolderDestName)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func helmDownload(chartUrl string, version string, dest string) error {
	cmd := exec.Command("helm", "pull", "--untar", "-d", dest, chartUrl, "--version", version)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(out))
	}
	return nil
}

func helmRepoAdd(name string, url string) error {
	cmd := exec.Command("helm", "repo", "add", "--force-update", name, url)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(out))
	}
	return nil
}