package cfg

func FilterCharts(config Config, charts []string) Config {
	chartsToFilter := make(map[string]bool)
	reposToFilter := make(map[string]bool)

	for _, name := range charts {
		chartsToFilter[name] = true
	}

	var filteredCharts []map[string]string
	for _, chart := range config.Charts {
		if chartsToFilter[chart["name"]] {
			filteredCharts = append(filteredCharts, chart)
			reposToFilter[chart["repo_name"]] = true
		}
	}

	var filteredRepos []map[string]string
	for _, repo := range config.Repos {
		if reposToFilter[repo["name"]] {
			filteredRepos = append(filteredRepos, repo)
		}
	}

	config.Charts = filteredCharts
	config.Repos = filteredRepos

	return config
}
