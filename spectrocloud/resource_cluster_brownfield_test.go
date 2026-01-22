package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

func TestToBrownfieldClusterSpecGeneric(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1SpectroGenericClusterImportEntitySpec
	}{
		{
			name:  "default values",
			input: map[string]interface{}{},
			expected: &models.V1SpectroGenericClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy:      nil,
				},
			},
		},
		{
			name: "import_mode full",
			input: map[string]interface{}{
				"import_mode": "full",
			},
			expected: &models.V1SpectroGenericClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy:      nil,
				},
			},
		},
		{
			name: "import_mode read_only",
			input: map[string]interface{}{
				"import_mode": "read_only",
			},
			expected: &models.V1SpectroGenericClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "read-only",
					Proxy:      nil,
				},
			},
		},
		{
			name: "with proxy fields",
			input: map[string]interface{}{
				"import_mode":          "full",
				"proxy":                "http://proxy.example.com:8080",
				"no_proxy":             "localhost,127.0.0.1",
				"host_path":            "/etc/ssl/certs/proxy-ca.pem",
				"container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1SpectroGenericClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy:            "http://proxy.example.com:8080",
						NoProxy:              "localhost,127.0.0.1",
						CaHostPath:           "/etc/ssl/certs/proxy-ca.pem",
						CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem",
					},
				},
			},
		},
		{
			name: "import_mode read_only with proxy",
			input: map[string]interface{}{
				"import_mode": "read_only",
				"proxy":       "http://proxy.example.com:8080",
				"no_proxy":    "localhost",
			},
			expected: &models.V1SpectroGenericClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "read-only",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy: "http://proxy.example.com:8080",
						NoProxy:   "localhost",
					},
				},
			},
		},
		{
			name: "partial proxy fields",
			input: map[string]interface{}{
				"import_mode": "full",
				"proxy":       "http://proxy.example.com:8080",
			},
			expected: &models.V1SpectroGenericClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy: "http://proxy.example.com:8080",
					},
				},
			},
		},
		{
			name: "only host_path and container_mount_path",
			input: map[string]interface{}{
				"import_mode":          "full",
				"host_path":            "/etc/ssl/certs/proxy-ca.pem",
				"container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1SpectroGenericClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						CaHostPath:           "/etc/ssl/certs/proxy-ca.pem",
						CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create schema for brownfield cluster resource
			schemaMap := resourceClusterBrownfield().Schema

			// Create ResourceData from input
			d := schema.TestResourceDataRaw(t, schemaMap, tt.input)

			// Call the function under test
			result := toBrownfieldClusterSpecGeneric(d)

			// Assert the result
			assert.NotNil(t, result)
			assert.NotNil(t, result.ClusterConfig)
			assert.Equal(t, tt.expected.ClusterConfig.ImportMode, result.ClusterConfig.ImportMode)

			// Assert proxy configuration
			if tt.expected.ClusterConfig.Proxy == nil {
				assert.Nil(t, result.ClusterConfig.Proxy)
			} else {
				assert.NotNil(t, result.ClusterConfig.Proxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.HTTPProxy, result.ClusterConfig.Proxy.HTTPProxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.NoProxy, result.ClusterConfig.Proxy.NoProxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.CaHostPath, result.ClusterConfig.Proxy.CaHostPath)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.CaContainerMountPath, result.ClusterConfig.Proxy.CaContainerMountPath)
			}
		})
	}
}

func TestToBrownfieldClusterSpecCloudStack(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1SpectroCloudStackClusterImportEntitySpec
	}{
		{
			name:  "default values",
			input: map[string]interface{}{},
			expected: &models.V1SpectroCloudStackClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy:      nil,
				},
			},
		},
		{
			name: "import_mode full",
			input: map[string]interface{}{
				"import_mode": "full",
			},
			expected: &models.V1SpectroCloudStackClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy:      nil,
				},
			},
		},
		{
			name: "import_mode read_only",
			input: map[string]interface{}{
				"import_mode": "read_only",
			},
			expected: &models.V1SpectroCloudStackClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "read-only",
					Proxy:      nil,
				},
			},
		},
		{
			name: "with proxy fields",
			input: map[string]interface{}{
				"import_mode":          "full",
				"proxy":                "http://proxy.example.com:8080",
				"no_proxy":             "localhost,127.0.0.1",
				"host_path":            "/etc/ssl/certs/proxy-ca.pem",
				"container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1SpectroCloudStackClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy:            "http://proxy.example.com:8080",
						NoProxy:              "localhost,127.0.0.1",
						CaHostPath:           "/etc/ssl/certs/proxy-ca.pem",
						CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem",
					},
				},
			},
		},
		{
			name: "import_mode read_only with proxy",
			input: map[string]interface{}{
				"import_mode": "read_only",
				"proxy":       "http://proxy.example.com:8080",
				"no_proxy":    "localhost",
			},
			expected: &models.V1SpectroCloudStackClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "read-only",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy: "http://proxy.example.com:8080",
						NoProxy:   "localhost",
					},
				},
			},
		},
		{
			name: "partial proxy fields",
			input: map[string]interface{}{
				"import_mode": "full",
				"proxy":       "http://proxy.example.com:8080",
			},
			expected: &models.V1SpectroCloudStackClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy: "http://proxy.example.com:8080",
					},
				},
			},
		},
		{
			name: "only host_path and container_mount_path",
			input: map[string]interface{}{
				"import_mode":          "full",
				"host_path":            "/etc/ssl/certs/proxy-ca.pem",
				"container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1SpectroCloudStackClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						CaHostPath:           "/etc/ssl/certs/proxy-ca.pem",
						CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create schema for brownfield cluster resource
			schemaMap := resourceClusterBrownfield().Schema

			// Create ResourceData from input
			d := schema.TestResourceDataRaw(t, schemaMap, tt.input)

			// Call the function under test
			result := toBrownfieldClusterSpecCloudStack(d)

			// Assert the result
			assert.NotNil(t, result)
			assert.NotNil(t, result.ClusterConfig)
			assert.Equal(t, tt.expected.ClusterConfig.ImportMode, result.ClusterConfig.ImportMode)

			// Assert proxy configuration
			if tt.expected.ClusterConfig.Proxy == nil {
				assert.Nil(t, result.ClusterConfig.Proxy)
			} else {
				assert.NotNil(t, result.ClusterConfig.Proxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.HTTPProxy, result.ClusterConfig.Proxy.HTTPProxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.NoProxy, result.ClusterConfig.Proxy.NoProxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.CaHostPath, result.ClusterConfig.Proxy.CaHostPath)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.CaContainerMountPath, result.ClusterConfig.Proxy.CaContainerMountPath)
			}
		})
	}
}

func TestToBrownfieldClusterSpecMaas(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1SpectroMaasClusterImportEntitySpec
	}{
		{
			name:  "default values",
			input: map[string]interface{}{},
			expected: &models.V1SpectroMaasClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy:      nil,
				},
			},
		},
		{
			name: "import_mode full",
			input: map[string]interface{}{
				"import_mode": "full",
			},
			expected: &models.V1SpectroMaasClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy:      nil,
				},
			},
		},
		{
			name: "import_mode read_only",
			input: map[string]interface{}{
				"import_mode": "read_only",
			},
			expected: &models.V1SpectroMaasClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "read-only",
					Proxy:      nil,
				},
			},
		},
		{
			name: "with proxy fields",
			input: map[string]interface{}{
				"import_mode":          "full",
				"proxy":                "http://proxy.example.com:8080",
				"no_proxy":             "localhost,127.0.0.1",
				"host_path":            "/etc/ssl/certs/proxy-ca.pem",
				"container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1SpectroMaasClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy:            "http://proxy.example.com:8080",
						NoProxy:              "localhost,127.0.0.1",
						CaHostPath:           "/etc/ssl/certs/proxy-ca.pem",
						CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem",
					},
				},
			},
		},
		{
			name: "import_mode read_only with proxy",
			input: map[string]interface{}{
				"import_mode": "read_only",
				"proxy":       "http://proxy.example.com:8080",
				"no_proxy":    "localhost",
			},
			expected: &models.V1SpectroMaasClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "read-only",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy: "http://proxy.example.com:8080",
						NoProxy:   "localhost",
					},
				},
			},
		},
		{
			name: "partial proxy fields",
			input: map[string]interface{}{
				"import_mode": "full",
				"proxy":       "http://proxy.example.com:8080",
			},
			expected: &models.V1SpectroMaasClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy: "http://proxy.example.com:8080",
					},
				},
			},
		},
		{
			name: "only host_path and container_mount_path",
			input: map[string]interface{}{
				"import_mode":          "full",
				"host_path":            "/etc/ssl/certs/proxy-ca.pem",
				"container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1SpectroMaasClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						CaHostPath:           "/etc/ssl/certs/proxy-ca.pem",
						CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create schema for brownfield cluster resource
			schemaMap := resourceClusterBrownfield().Schema

			// Create ResourceData from input
			d := schema.TestResourceDataRaw(t, schemaMap, tt.input)

			// Call the function under test
			result := toBrownfieldClusterSpecMaas(d)

			// Assert the result
			assert.NotNil(t, result)
			assert.NotNil(t, result.ClusterConfig)
			assert.Equal(t, tt.expected.ClusterConfig.ImportMode, result.ClusterConfig.ImportMode)

			// Assert proxy configuration
			if tt.expected.ClusterConfig.Proxy == nil {
				assert.Nil(t, result.ClusterConfig.Proxy)
			} else {
				assert.NotNil(t, result.ClusterConfig.Proxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.HTTPProxy, result.ClusterConfig.Proxy.HTTPProxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.NoProxy, result.ClusterConfig.Proxy.NoProxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.CaHostPath, result.ClusterConfig.Proxy.CaHostPath)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.CaContainerMountPath, result.ClusterConfig.Proxy.CaContainerMountPath)
			}
		})
	}
}

func TestToBrownfieldClusterSpecEdgeNative(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1SpectroEdgeNativeClusterImportEntitySpec
	}{
		{
			name:  "default values",
			input: map[string]interface{}{},
			expected: &models.V1SpectroEdgeNativeClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy:      nil,
				},
			},
		},
		{
			name: "import_mode full",
			input: map[string]interface{}{
				"import_mode": "full",
			},
			expected: &models.V1SpectroEdgeNativeClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy:      nil,
				},
			},
		},
		{
			name: "import_mode read_only",
			input: map[string]interface{}{
				"import_mode": "read_only",
			},
			expected: &models.V1SpectroEdgeNativeClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "read-only",
					Proxy:      nil,
				},
			},
		},
		{
			name: "with proxy fields",
			input: map[string]interface{}{
				"import_mode":          "full",
				"proxy":                "http://proxy.example.com:8080",
				"no_proxy":             "localhost,127.0.0.1",
				"host_path":            "/etc/ssl/certs/proxy-ca.pem",
				"container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1SpectroEdgeNativeClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy:            "http://proxy.example.com:8080",
						NoProxy:              "localhost,127.0.0.1",
						CaHostPath:           "/etc/ssl/certs/proxy-ca.pem",
						CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem",
					},
				},
			},
		},
		{
			name: "import_mode read_only with proxy",
			input: map[string]interface{}{
				"import_mode": "read_only",
				"proxy":       "http://proxy.example.com:8080",
				"no_proxy":    "localhost",
			},
			expected: &models.V1SpectroEdgeNativeClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "read-only",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy: "http://proxy.example.com:8080",
						NoProxy:   "localhost",
					},
				},
			},
		},
		{
			name: "partial proxy fields",
			input: map[string]interface{}{
				"import_mode": "full",
				"proxy":       "http://proxy.example.com:8080",
			},
			expected: &models.V1SpectroEdgeNativeClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy: "http://proxy.example.com:8080",
					},
				},
			},
		},
		{
			name: "only host_path and container_mount_path",
			input: map[string]interface{}{
				"import_mode":          "full",
				"host_path":            "/etc/ssl/certs/proxy-ca.pem",
				"container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1SpectroEdgeNativeClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						CaHostPath:           "/etc/ssl/certs/proxy-ca.pem",
						CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create schema for brownfield cluster resource
			schemaMap := resourceClusterBrownfield().Schema

			// Create ResourceData from input
			d := schema.TestResourceDataRaw(t, schemaMap, tt.input)

			// Call the function under test
			result := toBrownfieldClusterSpecEdgeNative(d)

			// Assert the result
			assert.NotNil(t, result)
			assert.NotNil(t, result.ClusterConfig)
			assert.Equal(t, tt.expected.ClusterConfig.ImportMode, result.ClusterConfig.ImportMode)

			// Assert proxy configuration
			if tt.expected.ClusterConfig.Proxy == nil {
				assert.Nil(t, result.ClusterConfig.Proxy)
			} else {
				assert.NotNil(t, result.ClusterConfig.Proxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.HTTPProxy, result.ClusterConfig.Proxy.HTTPProxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.NoProxy, result.ClusterConfig.Proxy.NoProxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.CaHostPath, result.ClusterConfig.Proxy.CaHostPath)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.CaContainerMountPath, result.ClusterConfig.Proxy.CaContainerMountPath)
			}
		})
	}
}

func TestToBrownfieldClusterSpecAws(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1SpectroAwsClusterImportEntitySpec
	}{
		{
			name:  "default values",
			input: map[string]interface{}{},
			expected: &models.V1SpectroAwsClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy:      nil,
				},
			},
		},
		{
			name: "import_mode full",
			input: map[string]interface{}{
				"import_mode": "full",
			},
			expected: &models.V1SpectroAwsClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy:      nil,
				},
			},
		},
		{
			name: "import_mode read_only",
			input: map[string]interface{}{
				"import_mode": "read_only",
			},
			expected: &models.V1SpectroAwsClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "read-only",
					Proxy:      nil,
				},
			},
		},
		{
			name: "with proxy fields",
			input: map[string]interface{}{
				"import_mode":          "full",
				"proxy":                "http://proxy.example.com:8080",
				"no_proxy":             "localhost,127.0.0.1",
				"host_path":            "/etc/ssl/certs/proxy-ca.pem",
				"container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1SpectroAwsClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy:            "http://proxy.example.com:8080",
						NoProxy:              "localhost,127.0.0.1",
						CaHostPath:           "/etc/ssl/certs/proxy-ca.pem",
						CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem",
					},
				},
			},
		},
		{
			name: "import_mode read_only with proxy",
			input: map[string]interface{}{
				"import_mode": "read_only",
				"proxy":       "http://proxy.example.com:8080",
				"no_proxy":    "localhost",
			},
			expected: &models.V1SpectroAwsClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "read-only",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy: "http://proxy.example.com:8080",
						NoProxy:   "localhost",
					},
				},
			},
		},
		{
			name: "partial proxy fields",
			input: map[string]interface{}{
				"import_mode": "full",
				"proxy":       "http://proxy.example.com:8080",
			},
			expected: &models.V1SpectroAwsClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy: "http://proxy.example.com:8080",
					},
				},
			},
		},
		{
			name: "only host_path and container_mount_path",
			input: map[string]interface{}{
				"import_mode":          "full",
				"host_path":            "/etc/ssl/certs/proxy-ca.pem",
				"container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1SpectroAwsClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						CaHostPath:           "/etc/ssl/certs/proxy-ca.pem",
						CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create schema for brownfield cluster resource
			schemaMap := resourceClusterBrownfield().Schema

			// Create ResourceData from input
			d := schema.TestResourceDataRaw(t, schemaMap, tt.input)

			// Call the function under test
			result := toBrownfieldClusterSpecAws(d)

			// Assert the result
			assert.NotNil(t, result)
			assert.NotNil(t, result.ClusterConfig)
			assert.Equal(t, tt.expected.ClusterConfig.ImportMode, result.ClusterConfig.ImportMode)

			// Assert proxy configuration
			if tt.expected.ClusterConfig.Proxy == nil {
				assert.Nil(t, result.ClusterConfig.Proxy)
			} else {
				assert.NotNil(t, result.ClusterConfig.Proxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.HTTPProxy, result.ClusterConfig.Proxy.HTTPProxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.NoProxy, result.ClusterConfig.Proxy.NoProxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.CaHostPath, result.ClusterConfig.Proxy.CaHostPath)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.CaContainerMountPath, result.ClusterConfig.Proxy.CaContainerMountPath)
			}
		})
	}
}

func TestToBrownfieldClusterSpecAzure(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1SpectroAzureClusterImportEntitySpec
	}{
		{
			name:  "default values",
			input: map[string]interface{}{},
			expected: &models.V1SpectroAzureClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy:      nil,
				},
			},
		},
		{
			name: "import_mode full",
			input: map[string]interface{}{
				"import_mode": "full",
			},
			expected: &models.V1SpectroAzureClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy:      nil,
				},
			},
		},
		{
			name: "import_mode read_only",
			input: map[string]interface{}{
				"import_mode": "read_only",
			},
			expected: &models.V1SpectroAzureClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "read-only",
					Proxy:      nil,
				},
			},
		},
		{
			name: "with proxy fields",
			input: map[string]interface{}{
				"import_mode":          "full",
				"proxy":                "http://proxy.example.com:8080",
				"no_proxy":             "localhost,127.0.0.1",
				"host_path":            "/etc/ssl/certs/proxy-ca.pem",
				"container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1SpectroAzureClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy:            "http://proxy.example.com:8080",
						NoProxy:              "localhost,127.0.0.1",
						CaHostPath:           "/etc/ssl/certs/proxy-ca.pem",
						CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem",
					},
				},
			},
		},
		{
			name: "import_mode read_only with proxy",
			input: map[string]interface{}{
				"import_mode": "read_only",
				"proxy":       "http://proxy.example.com:8080",
				"no_proxy":    "localhost",
			},
			expected: &models.V1SpectroAzureClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "read-only",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy: "http://proxy.example.com:8080",
						NoProxy:   "localhost",
					},
				},
			},
		},
		{
			name: "partial proxy fields",
			input: map[string]interface{}{
				"import_mode": "full",
				"proxy":       "http://proxy.example.com:8080",
			},
			expected: &models.V1SpectroAzureClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy: "http://proxy.example.com:8080",
					},
				},
			},
		},
		{
			name: "only host_path and container_mount_path",
			input: map[string]interface{}{
				"import_mode":          "full",
				"host_path":            "/etc/ssl/certs/proxy-ca.pem",
				"container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1SpectroAzureClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						CaHostPath:           "/etc/ssl/certs/proxy-ca.pem",
						CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create schema for brownfield cluster resource
			schemaMap := resourceClusterBrownfield().Schema

			// Create ResourceData from input
			d := schema.TestResourceDataRaw(t, schemaMap, tt.input)

			// Call the function under test
			result := toBrownfieldClusterSpecAzure(d)

			// Assert the result
			assert.NotNil(t, result)
			assert.NotNil(t, result.ClusterConfig)
			assert.Equal(t, tt.expected.ClusterConfig.ImportMode, result.ClusterConfig.ImportMode)

			// Assert proxy configuration
			if tt.expected.ClusterConfig.Proxy == nil {
				assert.Nil(t, result.ClusterConfig.Proxy)
			} else {
				assert.NotNil(t, result.ClusterConfig.Proxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.HTTPProxy, result.ClusterConfig.Proxy.HTTPProxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.NoProxy, result.ClusterConfig.Proxy.NoProxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.CaHostPath, result.ClusterConfig.Proxy.CaHostPath)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.CaContainerMountPath, result.ClusterConfig.Proxy.CaContainerMountPath)
			}
		})
	}
}

func TestToBrownfieldClusterSpecGcp(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1SpectroGcpClusterImportEntitySpec
	}{
		{
			name:  "default values",
			input: map[string]interface{}{},
			expected: &models.V1SpectroGcpClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy:      nil,
				},
			},
		},
		{
			name: "import_mode full",
			input: map[string]interface{}{
				"import_mode": "full",
			},
			expected: &models.V1SpectroGcpClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy:      nil,
				},
			},
		},
		{
			name: "import_mode read_only",
			input: map[string]interface{}{
				"import_mode": "read_only",
			},
			expected: &models.V1SpectroGcpClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "read-only",
					Proxy:      nil,
				},
			},
		},
		{
			name: "with proxy fields",
			input: map[string]interface{}{
				"import_mode":          "full",
				"proxy":                "http://proxy.example.com:8080",
				"no_proxy":             "localhost,127.0.0.1",
				"host_path":            "/etc/ssl/certs/proxy-ca.pem",
				"container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1SpectroGcpClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy:            "http://proxy.example.com:8080",
						NoProxy:              "localhost,127.0.0.1",
						CaHostPath:           "/etc/ssl/certs/proxy-ca.pem",
						CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem",
					},
				},
			},
		},
		{
			name: "import_mode read_only with proxy",
			input: map[string]interface{}{
				"import_mode": "read_only",
				"proxy":       "http://proxy.example.com:8080",
				"no_proxy":    "localhost",
			},
			expected: &models.V1SpectroGcpClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "read-only",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy: "http://proxy.example.com:8080",
						NoProxy:   "localhost",
					},
				},
			},
		},
		{
			name: "partial proxy fields",
			input: map[string]interface{}{
				"import_mode": "full",
				"proxy":       "http://proxy.example.com:8080",
			},
			expected: &models.V1SpectroGcpClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy: "http://proxy.example.com:8080",
					},
				},
			},
		},
		{
			name: "only host_path and container_mount_path",
			input: map[string]interface{}{
				"import_mode":          "full",
				"host_path":            "/etc/ssl/certs/proxy-ca.pem",
				"container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1SpectroGcpClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						CaHostPath:           "/etc/ssl/certs/proxy-ca.pem",
						CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create schema for brownfield cluster resource
			schemaMap := resourceClusterBrownfield().Schema

			// Create ResourceData from input
			d := schema.TestResourceDataRaw(t, schemaMap, tt.input)

			// Call the function under test
			result := toBrownfieldClusterSpecGcp(d)

			// Assert the result
			assert.NotNil(t, result)
			assert.NotNil(t, result.ClusterConfig)
			assert.Equal(t, tt.expected.ClusterConfig.ImportMode, result.ClusterConfig.ImportMode)

			// Assert proxy configuration
			if tt.expected.ClusterConfig.Proxy == nil {
				assert.Nil(t, result.ClusterConfig.Proxy)
			} else {
				assert.NotNil(t, result.ClusterConfig.Proxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.HTTPProxy, result.ClusterConfig.Proxy.HTTPProxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.NoProxy, result.ClusterConfig.Proxy.NoProxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.CaHostPath, result.ClusterConfig.Proxy.CaHostPath)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.CaContainerMountPath, result.ClusterConfig.Proxy.CaContainerMountPath)
			}
		})
	}
}

func TestToBrownfieldClusterSpecVsphere(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1SpectroVsphereClusterImportEntitySpec
	}{
		{
			name:  "default values",
			input: map[string]interface{}{},
			expected: &models.V1SpectroVsphereClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy:      nil,
				},
			},
		},
		{
			name: "import_mode full",
			input: map[string]interface{}{
				"import_mode": "full",
			},
			expected: &models.V1SpectroVsphereClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy:      nil,
				},
			},
		},
		{
			name: "import_mode read_only",
			input: map[string]interface{}{
				"import_mode": "read_only",
			},
			expected: &models.V1SpectroVsphereClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "read-only",
					Proxy:      nil,
				},
			},
		},
		{
			name: "with proxy fields",
			input: map[string]interface{}{
				"import_mode":          "full",
				"proxy":                "http://proxy.example.com:8080",
				"no_proxy":             "localhost,127.0.0.1",
				"host_path":            "/etc/ssl/certs/proxy-ca.pem",
				"container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1SpectroVsphereClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy:            "http://proxy.example.com:8080",
						NoProxy:              "localhost,127.0.0.1",
						CaHostPath:           "/etc/ssl/certs/proxy-ca.pem",
						CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem",
					},
				},
			},
		},
		{
			name: "import_mode read_only with proxy",
			input: map[string]interface{}{
				"import_mode": "read_only",
				"proxy":       "http://proxy.example.com:8080",
				"no_proxy":    "localhost",
			},
			expected: &models.V1SpectroVsphereClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "read-only",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy: "http://proxy.example.com:8080",
						NoProxy:   "localhost",
					},
				},
			},
		},
		{
			name: "partial proxy fields",
			input: map[string]interface{}{
				"import_mode": "full",
				"proxy":       "http://proxy.example.com:8080",
			},
			expected: &models.V1SpectroVsphereClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						HTTPProxy: "http://proxy.example.com:8080",
					},
				},
			},
		},
		{
			name: "only host_path and container_mount_path",
			input: map[string]interface{}{
				"import_mode":          "full",
				"host_path":            "/etc/ssl/certs/proxy-ca.pem",
				"container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1SpectroVsphereClusterImportEntitySpec{
				ClusterConfig: &models.V1ImportClusterConfig{
					ImportMode: "",
					Proxy: &models.V1ClusterProxySpec{
						CaHostPath:           "/etc/ssl/certs/proxy-ca.pem",
						CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create schema for brownfield cluster resource
			schemaMap := resourceClusterBrownfield().Schema

			// Create ResourceData from input
			d := schema.TestResourceDataRaw(t, schemaMap, tt.input)

			// Call the function under test
			result := toBrownfieldClusterSpecVsphere(d)

			// Assert the result
			assert.NotNil(t, result)
			assert.NotNil(t, result.ClusterConfig)
			assert.Equal(t, tt.expected.ClusterConfig.ImportMode, result.ClusterConfig.ImportMode)

			// Assert proxy configuration
			if tt.expected.ClusterConfig.Proxy == nil {
				assert.Nil(t, result.ClusterConfig.Proxy)
			} else {
				assert.NotNil(t, result.ClusterConfig.Proxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.HTTPProxy, result.ClusterConfig.Proxy.HTTPProxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.NoProxy, result.ClusterConfig.Proxy.NoProxy)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.CaHostPath, result.ClusterConfig.Proxy.CaHostPath)
				assert.Equal(t, tt.expected.ClusterConfig.Proxy.CaContainerMountPath, result.ClusterConfig.Proxy.CaContainerMountPath)
			}
		})
	}
}
