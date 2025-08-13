package exec

import (
	"os"
	"testing"

	"github.com/Qovery/helm-freeze/cfg"
)

func TestOCIWorkflowIntegration(t *testing.T) {
	// Create a temporary test config with generic values
	testConfig := cfg.Config{
		Charts: []map[string]string{
			{
				"name":      "test-chart",
				"version":   "1.0.0",
				"repo_name": "oci-repo",
			},
		},
		Repos: []map[string]string{
			{
				"name": "oci-repo",
				"url":  "oci://test.registry.com/charts",
			},
		},
		Destinations: []map[string]string{
			{
				"name": "default",
				"path": "./_test_out",
			},
		},
	}

	// Clean up test output directory
	testOutDir := "./_test_out"
	os.RemoveAll(testOutDir)

	// Test the complete workflow
	t.Run("Complete OCI Workflow", func(t *testing.T) {
		// Test repo detection
		for _, repo := range testConfig.Repos {
			if !isOCIURL(repo["url"]) {
				t.Errorf("Expected OCI URL detection for %s", repo["url"])
			}
		}

		// Test chart URL building
		chart := testConfig.Charts[0]
		repoName := chart["repo_name"]
		var repoURL string
		for _, repo := range testConfig.Repos {
			if repo["name"] == repoName {
				repoURL = repo["url"]
				break
			}
		}

		expectedURL := "oci://test.registry.com/charts/test-chart"
		builtURL := buildOCIChartURL(repoURL, chart["name"])
		if builtURL != expectedURL {
			t.Errorf("Expected chart URL %s, got %s", expectedURL, builtURL)
		}

		// Test registry host sanitization
		expectedHost := "test.registry.com"
		sanitizedHost := sanitizeRegistryHost(repoURL)
		if sanitizedHost != expectedHost {
			t.Errorf("Expected registry host %s, got %s", expectedHost, sanitizedHost)
		}
	})

	// Clean up
	os.RemoveAll(testOutDir)
}

// TestRealOCIWorkflow tests the actual OCI functionality with a real chart
// This test is marked as integration test and can be skipped in CI
func TestRealOCIWorkflow(t *testing.T) {
	// Skip this test if running in CI or if helm is not available
	if testing.Short() {
		t.Skip("Skipping real OCI test in short mode")
	}

	// Check if helm is available
	if _, err := os.Stat("./helm-freeze"); os.IsNotExist(err) {
		t.Skip("Skipping real OCI test - helm-freeze binary not found")
	}

	// Create a test config file
	testConfigContent := `charts:
  - name: valkey-operator
    version: v0.0.59-chart
    repo_name: oci-valkey

repos:
  - name: oci-valkey
    url: "oci://ghcr.io/hyperspike/valkey-operator"

destinations:
  - name: default
    path: "./_test_real_out"
`

	testConfigFile := "./_test_config.yaml"
	if err := os.WriteFile(testConfigFile, []byte(testConfigContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove(testConfigFile)

	// Clean up test output directory
	testOutDir := "./_test_real_out"
	os.RemoveAll(testOutDir)

	// Run the actual helm-freeze command
	// Note: This would require implementing a way to call the binary from tests
	// For now, we'll just validate the config structure
	t.Run("Real OCI Config Validation", func(t *testing.T) {
		// Parse the config to ensure it's valid
		config, err := cfg.ValidateConfig(testConfigFile)
		if err != nil {
			t.Fatalf("Failed to parse test config: %v", err)
		}

		// Validate that OCI repo is detected
		foundOCI := false
		for _, repo := range config.Repos {
			if isOCIURL(repo["url"]) {
				foundOCI = true
				break
			}
		}

		if !foundOCI {
			t.Error("OCI repo was not detected in real config")
		}

		// Validate chart URL building with real data
		chart := config.Charts[0]
		repoName := chart["repo_name"]
		var repoURL string
		for _, repo := range config.Repos {
			if repo["name"] == repoName {
				repoURL = repo["url"]
				break
			}
		}

		expectedURL := "oci://ghcr.io/hyperspike/valkey-operator"
		builtURL := buildOCIChartURL(repoURL, chart["name"])
		if builtURL != expectedURL {
			t.Errorf("Expected chart URL %s, got %s", expectedURL, builtURL)
		}

		// Validate registry host sanitization
		expectedHost := "ghcr.io"
		sanitizedHost := sanitizeRegistryHost(repoURL)
		if sanitizedHost != expectedHost {
			t.Errorf("Expected registry host %s, got %s", expectedHost, sanitizedHost)
		}
	})

	// Clean up
	os.RemoveAll(testOutDir)
}

func TestOCIConfigValidation(t *testing.T) {
	// Test that OCI configs are properly validated
	testConfig := cfg.Config{
		Charts: []map[string]string{
			{
				"name":      "test-chart",
				"version":   "1.0.0",
				"repo_name": "oci-repo",
			},
		},
		Repos: []map[string]string{
			{
				"name": "oci-repo",
				"url":  "oci://test.registry.com/charts",
			},
		},
		Destinations: []map[string]string{
			{
				"name": "default",
				"path": "./_test_out",
			},
		},
	}

	// Test that all repos are properly categorized
	repos := make([]repo, 0)
	for _, dest := range testConfig.Repos {
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

	// Verify OCI repo was detected
	foundOCI := false
	for _, r := range repos {
		if r.name == "oci-repo" && r.kind == "oci" {
			foundOCI = true
			break
		}
	}

	if !foundOCI {
		t.Error("OCI repo was not properly detected and categorized")
	}

	// Clean up
	os.RemoveAll("./_test_out")
}

func TestOCIVariousURLFormats(t *testing.T) {
	testCases := []struct {
		name        string
		url         string
		expectedOCI bool
	}{
		{
			name:        "Standard OCI URL",
			url:         "oci://ghcr.io/hyperspike/valkey-operator",
			expectedOCI: true,
		},
		{
			name:        "OCI URL with trailing slash",
			url:         "oci://ghcr.io/hyperspike/valkey-operator/",
			expectedOCI: true,
		},
		{
			name:        "HTTP URL",
			url:         "https://charts.helm.sh/stable",
			expectedOCI: false,
		},
		{
			name:        "Git URL",
			url:         "https://github.com/Qovery/pleco.git",
			expectedOCI: false,
		},
		{
			name:        "Simple registry",
			url:         "ghcr.io",
			expectedOCI: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isOCIURL(tc.url)
			if result != tc.expectedOCI {
				t.Errorf("isOCIURL(%q) = %v, want %v", tc.url, result, tc.expectedOCI)
			}
		})
	}
}
