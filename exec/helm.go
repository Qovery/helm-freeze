package exec

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Qovery/helm-freeze/cfg"
	"github.com/Qovery/helm-freeze/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/otiai10/copy"
)

func AddAllRepos(config cfg.Config) error {
	fmt.Println("\n[+] Adding helm repos")
	for _, repo := range config.Repos {
		fmt.Printf(" -> %s\n", repo["name"])

		// ignore if git repo
		if repo["type"] == "git" {
			continue
		}

		// handle OCI repos: login if credentials are provided, do not add as classic helm repo
		isOCI := repo["type"] == "oci" || isOCIURL(repo["url"])
		if isOCI {
			username := repo["username"]
			password := repo["password"]
			if username != "" {
				if err := helmRegistryLogin(repo["url"], username, password); err != nil {
					return err
				}
			}
			continue
		}

		err := helmRepoAdd(repo["name"], repo["url"])
		if err != nil {
			return err
		}
	}
	return nil
}

type repo struct {
	name string
	url  string
	kind string
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
	var repos []repo
	for _, dest := range config.Repos {
		repoType := "chart"
		if val, ok := dest["type"]; ok {
			repoType = val
		} else if isOCIURL(dest["url"]) {
			repoType = "oci"
		}

		repos = append(repos, repo{
			name: dest["name"],
			url:  dest["url"],
			kind: repoType,
		})
	}

	// download and extract all charts
	fmt.Println("\n[+] Downloading charts")
	for _, chart := range config.Charts {
		if noSync, ok := chart["no_sync"]; ok {
			if noSync == "true" {
				continue
			}
		}

		chartType := "chart"
		if _, ok := chart["chart_path"]; ok {
			chartType = "git"
		}

		if chartType == "chart" {
			err = getHelmChart(chart, destinations, repos)
			if err != nil {
				return err
			}
		} else if chartType == "git" {
			err = getGitChart(chart, destinations, repos)
			if err != nil {
				return err
			}
		} else {
			return errors.New("unknown chart type")
		}
	}

	return nil
}

// isOCIURL returns true if the provided repo url indicates an OCI registry (supports optional leading '@')
func isOCIURL(url string) bool {
	u := strings.TrimPrefix(url, "@")
	return strings.HasPrefix(u, "oci://")
}

func getHelmChart(chart map[string]string, destinations map[string]string, repos []repo) error {
	// set default values
	chartUrl := "stable/" + chart["name"]
	destinationFolder := destinations["default"]

	// use user defined repo if specified
	if repoNameDefined, ok := chart["repo_name"]; ok {
		// detect if the referenced repo is OCI
		repoIsOCI := false
		repoBaseUrl := ""
		for _, r := range repos {
			if r.name == repoNameDefined {
				if r.kind == "oci" {
					repoIsOCI = true
					repoBaseUrl = r.url
				}
				break
			}
		}
		if repoIsOCI || strings.HasPrefix(repoBaseUrl, "oci://") {
			chartUrl = buildOCIChartURL(repoBaseUrl, chart["name"])
		} else {
			chartUrl = repoNameDefined + "/" + chart["name"]
		}
	}

	// use user defined destinationFolder if specified
	if destinationIsDefined, ok := chart["dest"]; ok {
		destinationFolder = destinations[destinationIsDefined]
	}

	fmt.Printf(" -> %s %s\n", chartUrl, chart["version"])

	// move the current folder to avoid helm failure
	chartFolderDestName := destinationFolder + "/" + chart["name"]
	destOverride := ""
	if _, ok := chart["dest_folder_override"]; ok {
		chartFolderDestName = destinationFolder + "/" + chart["dest_folder_override"]
		destOverride = chart["dest_folder_override"]
	}
	oldChartFolderDestName := chartFolderDestName + ".old"

	chartExists := false
	if _, err := os.Stat(chartFolderDestName); !os.IsNotExist(err) {
		chartExists = true
	}

	_, err := os.Stat(oldChartFolderDestName)
	if err == nil {
		return fmt.Errorf("%s already exists, please fix manually then retry", oldChartFolderDestName)
	}

	if chartExists {
		err := os.Rename(chartFolderDestName, oldChartFolderDestName)
		if err != nil {
			return err
		}
	}

	err = helmDownload(chart["name"], chartUrl, chart["version"], destinationFolder, destOverride)
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

	return nil
}

func getGitChart(chart map[string]string, destinations map[string]string, repos []repo) error {
	// set default values
	chartName := chart["name"]
	chartUrl := ""
	destinationFolder := destinations["default"]
	repoPath := "/"

	fmt.Printf(" -> git/%s %s\n", chartName, chart["version"])

	if rPath, ok := chart["chart_path"]; ok {
		repoPath = rPath
	}

	// find chart Url
	for _, repo := range repos {
		if repo.kind == "git" {
			if repo.name == chart["repo_name"] {
				chartUrl = repo.url
				break
			}
		}
	}
	if chartUrl == "" {
		return errors.New("no url defined for this git repo")
	}

	// use user defined destinationFolder if specified
	if destinationIsDefined, ok := chart["dest"]; ok {
		destinationFolder = destinations[destinationIsDefined]
	}

	// move the current folder to avoid helm failure
	chartFolderDestName := destinationFolder + "/" + chart["name"]
	if _, ok := chart["dest_folder_override"]; ok {
		chartFolderDestName = destinationFolder + "/" + chart["dest_folder_override"]
	}
	oldChartFolderDestName := chartFolderDestName + ".helm_freeze_old"

	chartExists := false
	if _, err := os.Stat(chartFolderDestName); !os.IsNotExist(err) {
		chartExists = true
	}

	_, err := os.Stat(oldChartFolderDestName)
	if err == nil {
		return fmt.Errorf("%s already exists, please fix manually then retry", oldChartFolderDestName)
	}

	if chartExists {
		err := os.Rename(chartFolderDestName, oldChartFolderDestName)
		if err != nil {
			return err
		}
	}

	clonedRepoDir, err := gitClone(chartUrl, chart["version"])
	if err != nil {
		if chartExists {
			// restore old chart on failure
			_ = os.Rename(oldChartFolderDestName, chartFolderDestName)
		}
		return err
	}

	err = copy.Copy(clonedRepoDir+"/"+repoPath, chartFolderDestName)
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
	_ = os.RemoveAll(clonedRepoDir)

	return nil
}

func gitClone(gitUrl string, commitSha string) (string, error) {
	// create tmp dir to clone to this dir
	tmpDir, err := util.MkdirTemp()
	if err != nil {
		return "", err
	}
	r, err := git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL: gitUrl,
	})
	if err != nil {
		return "", err
	}

	w, err := r.Worktree()
	if err != nil {
		return "", err
	}

	// git checkout
	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(commitSha),
	})
	if err != nil {
		return "", err
	}

	return tmpDir, nil
}

func helmDownload(chartName string, chartUrl string, version string, dest string, destOverride string) error {
	destination := dest
	if destOverride != "" {
		destination = dest + "/" + destOverride + ".tmp"
		if _, err := os.Stat(destination); !os.IsNotExist(err) {
			os.RemoveAll(destination)
		}
	}

	cmd := exec.Command("helm", "pull", "--untar", "-d", destination, chartUrl, "--version", version)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(out))
	}

	if destOverride != "" {
		err := os.Rename(destination+"/"+chartName, dest+"/"+destOverride)
		if err != nil {
			return err
		}
		os.RemoveAll(destination)
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

func HelmRepoUpdate() error {
	fmt.Println("\n[+] Updating helm repos")
	cmd := exec.Command("helm", "repo", "update")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(out))
	}
	return nil
}

// helmRegistryLogin performs `helm registry login` for an OCI registry host if credentials are provided.
func helmRegistryLogin(repoURL string, username string, password string) error {
	host := sanitizeRegistryHost(repoURL)
	cmd := exec.Command("helm", "registry", "login", host, "-u", username, "-p", password)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(out))
	}
	return nil
}

// buildOCIChartURL builds a proper oci:// URL regardless of how the repo URL is provided
func buildOCIChartURL(repoURL string, chartName string) string {
	// remove leading @ if present and known schemes
	cleaned := strings.TrimPrefix(repoURL, "@")
	cleaned = strings.TrimPrefix(cleaned, "oci://")
	cleaned = strings.TrimPrefix(cleaned, "https://")
	cleaned = strings.TrimPrefix(cleaned, "http://")
	cleaned = strings.TrimSuffix(cleaned, "/")

	// if the last path segment already equals chartName, do not append it
	segments := strings.Split(cleaned, "/")
	if len(segments) > 0 {
		last := segments[len(segments)-1]
		if last == chartName {
			return "oci://" + cleaned
		}
	}
	return "oci://" + cleaned + "/" + chartName
}

// sanitizeRegistryHost extracts the registry host for helm registry login
func sanitizeRegistryHost(repoURL string) string {
	cleaned := strings.TrimPrefix(repoURL, "@")
	cleaned = strings.TrimPrefix(cleaned, "oci://")
	cleaned = strings.TrimPrefix(cleaned, "https://")
	cleaned = strings.TrimPrefix(cleaned, "http://")
	if idx := strings.Index(cleaned, "/"); idx != -1 {
		return cleaned[:idx]
	}
	return cleaned
}
