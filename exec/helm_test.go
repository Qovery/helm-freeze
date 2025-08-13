package exec

import (
	"testing"
)

func TestIsOCIURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "OCI URL with oci:// scheme",
			url:      "oci://ghcr.io/hyperspike/valkey-operator",
			expected: true,
		},
		{
			name:     "OCI URL with @ prefix",
			url:      "@oci://ghcr.io/hyperspike/valkey-operator",
			expected: true,
		},
		{
			name:     "HTTP URL",
			url:      "https://charts.helm.sh/stable",
			expected: false,
		},
		{
			name:     "Git URL",
			url:      "https://github.com/Qovery/pleco.git",
			expected: false,
		},
		{
			name:     "Empty URL",
			url:      "",
			expected: false,
		},
		{
			name:     "OCI URL with trailing slash",
			url:      "oci://ghcr.io/hyperspike/valkey-operator/",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOCIURL(tt.url)
			if result != tt.expected {
				t.Errorf("isOCIURL(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestBuildOCIChartURL(t *testing.T) {
	tests := []struct {
		name      string
		repoURL   string
		chartName string
		expected  string
	}{
		{
			name:      "Basic OCI URL",
			repoURL:   "oci://ghcr.io/hyperspike/valkey-operator",
			chartName: "valkey-operator",
			expected:  "oci://ghcr.io/hyperspike/valkey-operator",
		},
		{
			name:      "OCI URL with @ prefix",
			repoURL:   "@oci://ghcr.io/hyperspike/valkey-operator",
			chartName: "valkey-operator",
			expected:  "oci://ghcr.io/hyperspike/valkey-operator",
		},
		{
			name:      "OCI URL with trailing slash",
			repoURL:   "oci://ghcr.io/hyperspike/valkey-operator/",
			chartName: "valkey-operator",
			expected:  "oci://ghcr.io/hyperspike/valkey-operator",
		},
		{
			name:      "OCI URL with different chart name",
			repoURL:   "oci://ghcr.io/hyperspike/valkey-operator",
			chartName: "different-chart",
			expected:  "oci://ghcr.io/hyperspike/valkey-operator/different-chart",
		},
		{
			name:      "OCI URL with HTTP scheme prefix",
			repoURL:   "https://ghcr.io/hyperspike/valkey-operator",
			chartName: "valkey-operator",
			expected:  "oci://ghcr.io/hyperspike/valkey-operator",
		},
		{
			name:      "OCI URL with HTTP scheme prefix and @",
			repoURL:   "@https://ghcr.io/hyperspike/valkey-operator",
			chartName: "valkey-operator",
			expected:  "oci://ghcr.io/hyperspike/valkey-operator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildOCIChartURL(tt.repoURL, tt.chartName)
			if result != tt.expected {
				t.Errorf("buildOCIChartURL(%q, %q) = %v, want %v", tt.repoURL, tt.chartName, result, tt.expected)
			}
		})
	}
}

func TestSanitizeRegistryHost(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		expected string
	}{
		{
			name:     "Basic OCI URL",
			repoURL:  "oci://ghcr.io/hyperspike/valkey-operator",
			expected: "ghcr.io",
		},
		{
			name:     "OCI URL with @ prefix",
			repoURL:  "@oci://ghcr.io/hyperspike/valkey-operator",
			expected: "ghcr.io",
		},
		{
			name:     "OCI URL with HTTP scheme",
			repoURL:  "https://ghcr.io/hyperspike/valkey-operator",
			expected: "ghcr.io",
		},
		{
			name:     "OCI URL with HTTP scheme and @",
			repoURL:  "@https://ghcr.io/hyperspike/valkey-operator",
			expected: "ghcr.io",
		},
		{
			name:     "Simple registry host",
			repoURL:  "ghcr.io",
			expected: "ghcr.io",
		},
		{
			name:     "Registry with port",
			repoURL:  "oci://localhost:5000/my-charts",
			expected: "localhost:5000",
		},
		{
			name:     "Registry with subdomain",
			repoURL:  "oci://registry.example.com:5000/charts",
			expected: "registry.example.com:5000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeRegistryHost(tt.repoURL)
			if result != tt.expected {
				t.Errorf("sanitizeRegistryHost(%q) = %v, want %v", tt.repoURL, result, tt.expected)
			}
		})
	}
}
