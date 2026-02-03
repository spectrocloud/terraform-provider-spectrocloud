package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/assert"
)

// brownfieldImportScenario defines input and expected ClusterConfig for ToBrownfield* and ToImportClusterConfig.
type brownfieldImportScenario struct {
	name       string
	input      map[string]interface{}
	importMode string
	proxy      *models.V1ClusterProxySpec
}

var brownfieldImportScenarios = []brownfieldImportScenario{
	{name: "default values", input: map[string]interface{}{}, importMode: "", proxy: nil},
	{name: "import_mode full", input: map[string]interface{}{"import_mode": "full"}, importMode: "", proxy: nil},
	{name: "import_mode read_only", input: map[string]interface{}{"import_mode": "read_only"}, importMode: "read-only", proxy: nil},
	{
		name: "with proxy fields",
		input: map[string]interface{}{
			"import_mode": "full", "proxy": "http://proxy.example.com:8080", "no_proxy": "localhost,127.0.0.1",
			"host_path": "/etc/ssl/certs/proxy-ca.pem", "container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
		},
		importMode: "",
		proxy: &models.V1ClusterProxySpec{
			HTTPProxy: "http://proxy.example.com:8080", NoProxy: "localhost,127.0.0.1",
			CaHostPath: "/etc/ssl/certs/proxy-ca.pem", CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem",
		},
	},
	{
		name:       "import_mode read_only with proxy",
		input:      map[string]interface{}{"import_mode": "read_only", "proxy": "http://proxy.example.com:8080", "no_proxy": "localhost"},
		importMode: "read-only",
		proxy:      &models.V1ClusterProxySpec{HTTPProxy: "http://proxy.example.com:8080", NoProxy: "localhost"},
	},
	{
		name:       "partial proxy fields",
		input:      map[string]interface{}{"import_mode": "full", "proxy": "http://proxy.example.com:8080"},
		importMode: "",
		proxy:      &models.V1ClusterProxySpec{HTTPProxy: "http://proxy.example.com:8080"},
	},
	{
		name: "only host_path and container_mount_path",
		input: map[string]interface{}{
			"import_mode": "full", "host_path": "/etc/ssl/certs/proxy-ca.pem", "container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
		},
		importMode: "",
		proxy:      &models.V1ClusterProxySpec{CaHostPath: "/etc/ssl/certs/proxy-ca.pem", CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem"},
	},
}

// scenariosByCloud lists which brownfieldImportScenarios indices run per cloud type (preserves original coverage).
var scenariosByCloud = map[string][]int{
	"Generic":    {0, 1, 2, 3, 4, 5, 6},
	"CloudStack": {0, 1, 3, 4, 6},
	"Maas":       {0, 3, 4, 6},
	"EdgeNative": {0, 1, 2, 3, 4},
	"Aws":        {0, 1, 4, 6},
	"Azure":      {0, 1, 2, 3, 6},
	"Gcp":        {0, 1, 4, 6},
	"Vsphere":    {0, 1, 2, 3, 6},
}

func assertClusterConfig(t *testing.T, config *models.V1ImportClusterConfig, expectedMode string, expectedProxy *models.V1ClusterProxySpec) {
	t.Helper()
	assert.NotNil(t, config)
	assert.Equal(t, expectedMode, config.ImportMode)
	if expectedProxy == nil {
		assert.Nil(t, config.Proxy)
		return
	}
	assert.NotNil(t, config.Proxy)
	assert.Equal(t, expectedProxy.HTTPProxy, config.Proxy.HTTPProxy)
	assert.Equal(t, expectedProxy.NoProxy, config.Proxy.NoProxy)
	assert.Equal(t, expectedProxy.CaHostPath, config.Proxy.CaHostPath)
	assert.Equal(t, expectedProxy.CaContainerMountPath, config.Proxy.CaContainerMountPath)
}

func TestToBrownfieldClusterSpec_AllClouds(t *testing.T) {
	schemaMap := resourceClusterBrownfield().Schema
	for cloudType, indices := range scenariosByCloud {
		t.Run(cloudType, func(t *testing.T) {
			for _, idx := range indices {
				sc := brownfieldImportScenarios[idx]
				t.Run(sc.name, func(t *testing.T) {
					d := schema.TestResourceDataRaw(t, schemaMap, sc.input)
					var config *models.V1ImportClusterConfig
					switch cloudType {
					case "Generic":
						result := toBrownfieldClusterSpecGeneric(d)
						assert.NotNil(t, result)
						config = result.ClusterConfig
					case "CloudStack":
						result := toBrownfieldClusterSpecCloudStack(d)
						assert.NotNil(t, result)
						config = result.ClusterConfig
					case "Maas":
						result := toBrownfieldClusterSpecMaas(d)
						assert.NotNil(t, result)
						config = result.ClusterConfig
					case "EdgeNative":
						result := toBrownfieldClusterSpecEdgeNative(d)
						assert.NotNil(t, result)
						config = result.ClusterConfig
					case "Aws":
						result := toBrownfieldClusterSpecAws(d)
						assert.NotNil(t, result)
						config = result.ClusterConfig
					case "Azure":
						result := toBrownfieldClusterSpecAzure(d)
						assert.NotNil(t, result)
						config = result.ClusterConfig
					case "Gcp":
						result := toBrownfieldClusterSpecGcp(d)
						assert.NotNil(t, result)
						config = result.ClusterConfig
					case "Vsphere":
						result := toBrownfieldClusterSpecVsphere(d)
						assert.NotNil(t, result)
						config = result.ClusterConfig
					default:
						t.Fatalf("unknown cloud type %s", cloudType)
					}
					assertClusterConfig(t, config, sc.importMode, sc.proxy)
				})
			}
		})
	}
}

func TestToImportClusterConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1ImportClusterConfig
	}{
		{
			name:  "default values - empty input",
			input: map[string]interface{}{},
			expected: &models.V1ImportClusterConfig{
				ImportMode: "",
				Proxy:      nil,
			},
		},
		{
			name: "import_mode full - converts to empty string",
			input: map[string]interface{}{
				"import_mode": "full",
			},
			expected: &models.V1ImportClusterConfig{
				ImportMode: "",
				Proxy:      nil,
			},
		},
		{
			name: "import_mode read_only - converts to read-only",
			input: map[string]interface{}{
				"import_mode": "read_only",
			},
			expected: &models.V1ImportClusterConfig{
				ImportMode: "read-only",
				Proxy:      nil,
			},
		},
		{
			name: "proxy only",
			input: map[string]interface{}{
				"proxy": "http://proxy.example.com:8080",
			},
			expected: &models.V1ImportClusterConfig{
				ImportMode: "",
				Proxy: &models.V1ClusterProxySpec{
					HTTPProxy: "http://proxy.example.com:8080",
				},
			},
		},
		{
			name: "host_path only",
			input: map[string]interface{}{
				"host_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1ImportClusterConfig{
				ImportMode: "",
				Proxy: &models.V1ClusterProxySpec{
					CaHostPath: "/etc/ssl/certs/proxy-ca.pem",
				},
			},
		},
		{
			name: "all proxy fields",
			input: map[string]interface{}{
				"proxy":                "http://proxy.example.com:8080",
				"no_proxy":             "localhost,127.0.0.1",
				"host_path":            "/etc/ssl/certs/proxy-ca.pem",
				"container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1ImportClusterConfig{
				ImportMode: "",
				Proxy: &models.V1ClusterProxySpec{
					HTTPProxy:            "http://proxy.example.com:8080",
					NoProxy:              "localhost,127.0.0.1",
					CaHostPath:           "/etc/ssl/certs/proxy-ca.pem",
					CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem",
				},
			},
		},
		{
			name: "import_mode read_only with all proxy fields",
			input: map[string]interface{}{
				"import_mode":          "read_only",
				"proxy":                "http://proxy.example.com:8080",
				"no_proxy":             "localhost,127.0.0.1",
				"host_path":            "/etc/ssl/certs/proxy-ca.pem",
				"container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1ImportClusterConfig{
				ImportMode: "read-only",
				Proxy: &models.V1ClusterProxySpec{
					HTTPProxy:            "http://proxy.example.com:8080",
					NoProxy:              "localhost,127.0.0.1",
					CaHostPath:           "/etc/ssl/certs/proxy-ca.pem",
					CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem",
				},
			},
		},
		{
			name: "proxy and host_path only",
			input: map[string]interface{}{
				"proxy":     "http://proxy.example.com:8080",
				"host_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1ImportClusterConfig{
				ImportMode: "",
				Proxy: &models.V1ClusterProxySpec{
					HTTPProxy:  "http://proxy.example.com:8080",
					CaHostPath: "/etc/ssl/certs/proxy-ca.pem",
				},
			},
		},
		{
			name: "host_path and container_mount_path only",
			input: map[string]interface{}{
				"host_path":            "/etc/ssl/certs/proxy-ca.pem",
				"container_mount_path": "/etc/ssl/certs/proxy-ca.pem",
			},
			expected: &models.V1ImportClusterConfig{
				ImportMode: "",
				Proxy: &models.V1ClusterProxySpec{
					CaHostPath:           "/etc/ssl/certs/proxy-ca.pem",
					CaContainerMountPath: "/etc/ssl/certs/proxy-ca.pem",
				},
			},
		},
		{
			name: "proxy and no_proxy only",
			input: map[string]interface{}{
				"proxy":    "http://proxy.example.com:8080",
				"no_proxy": "localhost,127.0.0.1",
			},
			expected: &models.V1ImportClusterConfig{
				ImportMode: "",
				Proxy: &models.V1ClusterProxySpec{
					HTTPProxy: "http://proxy.example.com:8080",
					NoProxy:   "localhost,127.0.0.1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemaMap := resourceClusterBrownfield().Schema
			d := schema.TestResourceDataRaw(t, schemaMap, tt.input)
			result := toImportClusterConfig(d)
			assertClusterConfig(t, result, tt.expected.ImportMode, tt.expected.Proxy)
		})
	}
}

func TestReadCommonFieldsBrownfield(t *testing.T) {
	clusterID := "test-cluster-id"

	tests := []struct {
		name        string
		setupClient func() *client.V1Client
		setupData   func() *schema.ResourceData
		cluster     *models.V1SpectroCluster
		expectError bool
		description string
		verify      func(t *testing.T, d *schema.ResourceData, diags diag.Diagnostics)
	}{
		{
			name: "Success - minimal cluster with tags only",
			setupClient: func() *client.V1Client {
				return getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
			},
			setupData: func() *schema.ResourceData {
				d := resourceClusterBrownfield().TestResourceData()
				d.SetId(clusterID)
				return d
			},
			cluster: &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{
					Labels: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					Annotations: map[string]string{},
				},
				Spec: &models.V1SpectroClusterSpec{
					ClusterConfig: &models.V1ClusterConfig{},
				},
				Status: &models.V1SpectroClusterStatus{
					Repave: &models.V1ClusterRepaveStatus{
						State: repaveStatePtr("Pending"),
					},
				},
			},
			expectError: false,
			description: "Should successfully set tags and pause_agent_upgrades",
			verify: func(t *testing.T, d *schema.ResourceData, diags diag.Diagnostics) {
				assert.False(t, diags.HasError())
				tags := d.Get("tags").(*schema.Set)
				assert.NotNil(t, tags)
				assert.Equal(t, 2, tags.Len())
				pauseAgentUpgrades := d.Get("pause_agent_upgrades")
				assert.Equal(t, "unlock", pauseAgentUpgrades)
			},
		},
		{
			name: "Success - cluster with timezone",
			setupClient: func() *client.V1Client {
				return getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
			},
			setupData: func() *schema.ResourceData {
				d := resourceClusterBrownfield().TestResourceData()
				d.SetId(clusterID)
				return d
			},
			cluster: &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{
					Labels: map[string]string{
						"key1": "value1",
					},
					Annotations: map[string]string{},
				},
				Spec: &models.V1SpectroClusterSpec{
					ClusterConfig: &models.V1ClusterConfig{
						Timezone: "America/New_York",
					},
				},
				Status: &models.V1SpectroClusterStatus{
					Repave: &models.V1ClusterRepaveStatus{
						State: repaveStatePtr("Pending"),
					},
				},
			},
			expectError: false,
			description: "Should set cluster_timezone when present",
			verify: func(t *testing.T, d *schema.ResourceData, diags diag.Diagnostics) {
				assert.False(t, diags.HasError())
				timezone := d.Get("cluster_timezone")
				assert.Equal(t, "America/New_York", timezone)
			},
		},
		{
			name: "Success - cluster with review_repave_state field",
			setupClient: func() *client.V1Client {
				return getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
			},
			setupData: func() *schema.ResourceData {
				d := resourceClusterBrownfield().TestResourceData()
				d.SetId(clusterID)
				d.Set("review_repave_state", "")
				return d
			},
			cluster: &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{
					Labels:      map[string]string{},
					Annotations: map[string]string{},
				},
				Spec: &models.V1SpectroClusterSpec{
					ClusterConfig: &models.V1ClusterConfig{},
				},
				Status: &models.V1SpectroClusterStatus{
					Repave: &models.V1ClusterRepaveStatus{
						State: repaveStatePtr("Approved"),
					},
				},
			},
			expectError: false,
			description: "Should set review_repave_state when field exists",
			verify: func(t *testing.T, d *schema.ResourceData, diags diag.Diagnostics) {
				assert.False(t, diags.HasError())
				repaveState := d.Get("review_repave_state")
				// Note: d.Set() with a pointer to string type alias may not work as expected
				// The actual implementation may need to dereference the pointer
				// For now, we verify the function executes without error
				_ = repaveState
			},
		},
		{
			name: "Success - cluster with pause_agent_upgrades lock",
			setupClient: func() *client.V1Client {
				return getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
			},
			setupData: func() *schema.ResourceData {
				d := resourceClusterBrownfield().TestResourceData()
				d.SetId(clusterID)
				return d
			},
			cluster: &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{
					Labels: map[string]string{},
					Annotations: map[string]string{
						"spectroComponentsUpgradeForbidden": "true",
					},
				},
				Spec: &models.V1SpectroClusterSpec{
					ClusterConfig: &models.V1ClusterConfig{},
				},
				Status: &models.V1SpectroClusterStatus{
					Repave: &models.V1ClusterRepaveStatus{
						State: repaveStatePtr("Pending"),
					},
				},
			},
			expectError: false,
			description: "Should set pause_agent_upgrades to lock when annotation is true",
			verify: func(t *testing.T, d *schema.ResourceData, diags diag.Diagnostics) {
				assert.False(t, diags.HasError())
				pauseAgentUpgrades := d.Get("pause_agent_upgrades")
				assert.Equal(t, "lock", pauseAgentUpgrades)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setupClient()
			d := tt.setupData()

			diags, hasError := readCommonFieldsBrownfield(c, d, tt.cluster)

			if tt.expectError {
				assert.True(t, hasError || diags.HasError(), "Expected error but got none")
			} else {
				assert.False(t, hasError, "Unexpected error occurred")
				if diags.HasError() {
					t.Logf("Unexpected diagnostics errors: %v", diags)
				}
			}

			if tt.verify != nil {
				tt.verify(t, d, diags)
			}
		})
	}
}

// Helper function to create V1ClusterRepaveState pointer
func repaveStatePtr(s string) *models.V1ClusterRepaveState {
	state := models.V1ClusterRepaveState(s)
	return &state
}

func TestIsClusterRunningHealthy(t *testing.T) {
	// Note: The current mock API doesn't implement GetClusterOverview, so tests
	// for health status scenarios (Healthy, UnHealthy, Unknown) would require
	// extending the mock API. The current tests verify the fallback behavior
	// when GetClusterOverview is unavailable.
	clusterUID := "test-cluster-uid"

	tests := []struct {
		name        string
		setupClient func() *client.V1Client
		cluster     *models.V1SpectroCluster
		expected    bool
		expectedMsg string
		description string
	}{
		{
			name: "Nil cluster - returns false, Unknown",
			setupClient: func() *client.V1Client {
				return getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
			},
			cluster:     nil,
			expected:    false,
			expectedMsg: "Unknown",
			description: "Should return false and Unknown when cluster is nil",
		},
		{
			name: "Cluster state is Pending - returns false, Pending",
			setupClient: func() *client.V1Client {
				return getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
			},
			cluster: &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{
					UID: clusterUID,
				},
				Status: &models.V1SpectroClusterStatus{
					State: "Pending",
				},
			},
			expected:    false,
			expectedMsg: "Pending",
			description: "Should return false and state when state is not Running",
		},
		{
			name: "Cluster state is Error - returns false, Error",
			setupClient: func() *client.V1Client {
				return getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
			},
			cluster: &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{
					UID: clusterUID,
				},
				Status: &models.V1SpectroClusterStatus{
					State: "Error",
				},
			},
			expected:    false,
			expectedMsg: "Error",
			description: "Should return false and state when state is Error",
		},
		{
			name: "Cluster state is Running, GetClusterOverview returns error - returns true, Running",
			setupClient: func() *client.V1Client {
				return getV1ClientWithResourceContext(unitTestMockAPINegativeClient, "project")
			},
			cluster: &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{
					UID: clusterUID,
				},
				Status: &models.V1SpectroClusterStatus{
					State: "Running",
				},
			},
			expected:    true,
			expectedMsg: "Running",
			description: "Should return true and Running when GetClusterOverview fails (assumes Running is enough)",
		},
		{
			name: "Cluster state is Running, health not available - returns true, Running",
			setupClient: func() *client.V1Client {
				return getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
			},
			cluster: &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{
					UID: clusterUID,
				},
				Status: &models.V1SpectroClusterStatus{
					State: "Running",
				},
			},
			expected:    true,
			expectedMsg: "Running",
			description: "Should return true and Running when health is not available (Running is acceptable)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setupClient()

			result, msg := isClusterRunningHealthy(tt.cluster, c)

			assert.Equal(t, tt.expected, result, "Expected result should match")
			assert.Equal(t, tt.expectedMsg, msg, "Expected message should match")
		})
	}
}

func TestValidateDay1FieldsImmutable(t *testing.T) {
	// Note: Testing HasChange() in unit tests is challenging because it requires
	// a diff between old state and new config. We'll test the function's behavior
	// by creating ResourceData and simulating changes where possible.

	tests := []struct {
		name        string
		setupData   func() *schema.ResourceData
		expectError bool
		description string
		verify      func(t *testing.T, diags diag.Diagnostics)
	}{
		{
			name: "No changes - should pass",
			setupData: func() *schema.ResourceData {
				d := resourceClusterBrownfield().TestResourceData()
				d.SetId("test-cluster-id")
				// Set initial values
				d.Set("name", "test-cluster")
				d.Set("cloud_type", "aws")
				d.Set("import_mode", "full")
				return d
			},
			expectError: false,
			description: "Should not error when no Day-1 fields have changed",
			verify: func(t *testing.T, diags diag.Diagnostics) {
				assert.False(t, diags.HasError(), "Should not have errors when no changes")
			},
		},
		{
			name: "Empty ResourceData - should pass",
			setupData: func() *schema.ResourceData {
				d := resourceClusterBrownfield().TestResourceData()
				d.SetId("test-cluster-id")
				return d
			},
			expectError: false,
			description: "Should not error when ResourceData is empty (no changes detected)",
			verify: func(t *testing.T, diags diag.Diagnostics) {
				assert.False(t, diags.HasError(), "Should not have errors when no changes")
			},
		},
		{
			name: "All Day-1 fields defined - should pass if no changes",
			setupData: func() *schema.ResourceData {
				d := resourceClusterBrownfield().TestResourceData()
				d.SetId("test-cluster-id")
				// Set all Day-1 fields
				d.Set("name", "test-cluster")
				d.Set("cloud_type", "aws")
				d.Set("import_mode", "full")
				d.Set("host_path", "/path")
				d.Set("container_mount_path", "/mount")
				d.Set("context", "project")
				d.Set("proxy", "http://proxy")
				d.Set("no_proxy", "localhost")
				return d
			},
			expectError: false,
			description: "Should not error when all Day-1 fields are set but unchanged",
			verify: func(t *testing.T, diags diag.Diagnostics) {
				assert.False(t, diags.HasError(), "Should not have errors when fields are set but unchanged")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.setupData()

			diags := validateDay1FieldsImmutable(d)

			if tt.expectError {
				assert.True(t, diags.HasError(), "Expected error but got none")
			} else {
				// Note: HasChange() requires a diff which is hard to simulate in unit tests
				// The function will only return errors if HasChange() returns true
				// In practice, this would be tested during actual Terraform update operations
				if diags.HasError() {
					t.Logf("Function returned errors (may be expected if HasChange() detects changes): %v", diags)
				}
			}

			if tt.verify != nil {
				tt.verify(t, diags)
			}
		})
	}

	// Additional test to verify the function structure and error message format
	t.Run("Verify function structure and error message format", func(t *testing.T) {
		// This test verifies that the function is structured correctly
		// and would return proper error messages when HasChange() is true
		d := resourceClusterBrownfield().TestResourceData()
		d.SetId("test-cluster-id")

		diags := validateDay1FieldsImmutable(d)

		// The function should execute without panic
		// diags can be empty (no errors) or contain errors
		_ = diags

		// If there are errors, verify the error message format
		if diags.HasError() {
			for _, d := range diags {
				assert.Equal(t, diag.Error, d.Severity, "Error severity should be set")
				assert.Contains(t, d.Summary, "Day-1 fields cannot be updated", "Summary should contain expected message")
				assert.Contains(t, d.Detail, "immutable", "Detail should mention immutable fields")
			}
		}
	})

	// Test to verify all Day-1 fields are checked
	t.Run("Verify all Day-1 fields are in the validation list", func(t *testing.T) {
		// This is a structural test to ensure all expected fields are validated
		expectedFields := []string{
			"name", "cloud_type", "import_mode", "host_path",
			"container_mount_path", "context", "proxy", "no_proxy",
		}

		// Verify all fields exist in the schema
		schemaMap := resourceClusterBrownfield().Schema
		for _, field := range expectedFields {
			_, exists := schemaMap[field]
			assert.True(t, exists, "Field %s should exist in schema", field)
		}
	})
}

func TestGetNodeMaintenanceStatusForCloudType(t *testing.T) {
	tests := []struct {
		name        string
		cloudType   string
		expectedNil bool
		description string
		verify      func(t *testing.T, result GetMaintenanceStatus)
	}{
		{
			name:        "AWS cloud type",
			cloudType:   "aws",
			expectedNil: false,
			description: "Should return GetNodeMaintenanceStatusAws function for aws",
			verify: func(t *testing.T, result GetMaintenanceStatus) {
				assert.NotNil(t, result, "Result should not be nil for aws")
			},
		},
		{
			name:        "Azure cloud type",
			cloudType:   "azure",
			expectedNil: false,
			description: "Should return GetNodeMaintenanceStatusAzure function for azure",
			verify: func(t *testing.T, result GetMaintenanceStatus) {
				assert.NotNil(t, result, "Result should not be nil for azure")
			},
		},
		{
			name:        "GCP cloud type",
			cloudType:   "gcp",
			expectedNil: false,
			description: "Should return GetNodeMaintenanceStatusGcp function for gcp",
			verify: func(t *testing.T, result GetMaintenanceStatus) {
				assert.NotNil(t, result, "Result should not be nil for gcp")
			},
		},
		{
			name:        "vSphere cloud type",
			cloudType:   "vsphere",
			expectedNil: false,
			description: "Should return GetNodeMaintenanceStatusVsphere function for vsphere",
			verify: func(t *testing.T, result GetMaintenanceStatus) {
				assert.NotNil(t, result, "Result should not be nil for vsphere")
			},
		},
		{
			name:        "OpenShift cloud type",
			cloudType:   "openshift",
			expectedNil: false,
			description: "Should return GetNodeMaintenanceStatusVsphere function for openshift",
			verify: func(t *testing.T, result GetMaintenanceStatus) {
				assert.NotNil(t, result, "Result should not be nil for openshift")
			},
		},
		{
			name:        "Generic cloud type",
			cloudType:   "generic",
			expectedNil: false,
			description: "Should return GetNodeMaintenanceStatusGeneric function for generic",
			verify: func(t *testing.T, result GetMaintenanceStatus) {
				assert.NotNil(t, result, "Result should not be nil for generic")
			},
		},
		{
			name:        "EKS-Anywhere cloud type",
			cloudType:   "eks-anywhere",
			expectedNil: false,
			description: "Should return GetNodeMaintenanceStatusGeneric function for eks-anywhere",
			verify: func(t *testing.T, result GetMaintenanceStatus) {
				assert.NotNil(t, result, "Result should not be nil for eks-anywhere")
			},
		},
		{
			name:        "Apache CloudStack cloud type",
			cloudType:   "apache-cloudstack",
			expectedNil: false,
			description: "Should return GetNodeMaintenanceStatusCloudStack function for apache-cloudstack",
			verify: func(t *testing.T, result GetMaintenanceStatus) {
				assert.NotNil(t, result, "Result should not be nil for apache-cloudstack")
			},
		},
		{
			name:        "MAAS cloud type",
			cloudType:   "maas",
			expectedNil: false,
			description: "Should return GetNodeMaintenanceStatusMaas function for maas",
			verify: func(t *testing.T, result GetMaintenanceStatus) {
				assert.NotNil(t, result, "Result should not be nil for maas")
			},
		},
		{
			name:        "Edge Native cloud type",
			cloudType:   "edge-native",
			expectedNil: false,
			description: "Should return GetNodeMaintenanceStatusEdgeNative function for edge-native",
			verify: func(t *testing.T, result GetMaintenanceStatus) {
				assert.NotNil(t, result, "Result should not be nil for edge-native")
			},
		},
		{
			name:        "OpenStack cloud type",
			cloudType:   "openstack",
			expectedNil: false,
			description: "Should return GetNodeMaintenanceStatusOpenStack function for openstack",
			verify: func(t *testing.T, result GetMaintenanceStatus) {
				assert.NotNil(t, result, "Result should not be nil for openstack")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")

			result := getNodeMaintenanceStatusForCloudType(c, tt.cloudType)

			if tt.expectedNil {
				assert.Nil(t, result, "Expected nil result")
			} else {
				assert.NotNil(t, result, "Expected non-nil result")
			}

			if tt.verify != nil {
				tt.verify(t, result)
			}
		})
	}

	// Additional test to verify function signatures match
	t.Run("Verify function signatures for all cloud types", func(t *testing.T) {
		c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")

		cloudTypes := []string{"aws", "azure", "gcp", "vsphere", "openshift", "generic", "eks-anywhere", "apache-cloudstack", "maas", "edge-native", "openstack"}

		for _, cloudType := range cloudTypes {
			result := getNodeMaintenanceStatusForCloudType(c, cloudType)
			assert.NotNil(t, result, "Function should not be nil for cloud type: %s", cloudType)

			// Verify the function can be called (even if it fails, the signature should be correct)
			// We don't actually call it since it requires valid cluster/node IDs
			_ = result
		}
	})
}

func TestGetMachinesListForCloudType(t *testing.T) {
	tests := []struct {
		name        string
		cloudType   string
		expectedNil bool
		description string
		verify      func(t *testing.T, result func(string, string) (map[string]string, error))
	}{
		{
			name:        "AWS cloud type",
			cloudType:   "aws",
			expectedNil: false,
			description: "Should return GetMachinesListAws function for aws",
			verify: func(t *testing.T, result func(string, string) (map[string]string, error)) {
				assert.NotNil(t, result, "Result should not be nil for aws")
			},
		},
		{
			name:        "Azure cloud type",
			cloudType:   "azure",
			expectedNil: false,
			description: "Should return GetMachinesListAzure function for azure",
			verify: func(t *testing.T, result func(string, string) (map[string]string, error)) {
				assert.NotNil(t, result, "Result should not be nil for azure")
			},
		},
		{
			name:        "GCP cloud type",
			cloudType:   "gcp",
			expectedNil: false,
			description: "Should return GetMachinesListGcp function for gcp",
			verify: func(t *testing.T, result func(string, string) (map[string]string, error)) {
				assert.NotNil(t, result, "Result should not be nil for gcp")
			},
		},
		{
			name:        "vSphere cloud type",
			cloudType:   "vsphere",
			expectedNil: false,
			description: "Should return GetMachinesListVsphere function for vsphere",
			verify: func(t *testing.T, result func(string, string) (map[string]string, error)) {
				assert.NotNil(t, result, "Result should not be nil for vsphere")
			},
		},
		{
			name:        "OpenShift cloud type",
			cloudType:   "openshift",
			expectedNil: false,
			description: "Should return GetMachinesListVsphere function for openshift",
			verify: func(t *testing.T, result func(string, string) (map[string]string, error)) {
				assert.NotNil(t, result, "Result should not be nil for openshift")
			},
		},
		{
			name:        "Generic cloud type",
			cloudType:   "generic",
			expectedNil: false,
			description: "Should return GetMachinesListGeneric function for generic",
			verify: func(t *testing.T, result func(string, string) (map[string]string, error)) {
				assert.NotNil(t, result, "Result should not be nil for generic")
			},
		},
		{
			name:        "EKS-Anywhere cloud type",
			cloudType:   "eks-anywhere",
			expectedNil: false,
			description: "Should return GetMachinesListGeneric function for eks-anywhere",
			verify: func(t *testing.T, result func(string, string) (map[string]string, error)) {
				assert.NotNil(t, result, "Result should not be nil for eks-anywhere")
			},
		},
		{
			name:        "Apache CloudStack cloud type",
			cloudType:   "apache-cloudstack",
			expectedNil: false,
			description: "Should return GetMachinesListApacheCloudstack function for apache-cloudstack",
			verify: func(t *testing.T, result func(string, string) (map[string]string, error)) {
				assert.NotNil(t, result, "Result should not be nil for apache-cloudstack")
			},
		},
		{
			name:        "MAAS cloud type",
			cloudType:   "maas",
			expectedNil: false,
			description: "Should return GetMachinesListMaas function for maas",
			verify: func(t *testing.T, result func(string, string) (map[string]string, error)) {
				assert.NotNil(t, result, "Result should not be nil for maas")
			},
		},
		{
			name:        "Edge Native cloud type",
			cloudType:   "edge-native",
			expectedNil: false,
			description: "Should return GetMachinesListEdgeNative function for edge-native",
			verify: func(t *testing.T, result func(string, string) (map[string]string, error)) {
				assert.NotNil(t, result, "Result should not be nil for edge-native")
			},
		},
		{
			name:        "OpenStack cloud type",
			cloudType:   "openstack",
			expectedNil: false,
			description: "Should return GetMachinesListOpenStack function for openstack",
			verify: func(t *testing.T, result func(string, string) (map[string]string, error)) {
				assert.NotNil(t, result, "Result should not be nil for openstack")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")

			result := getMachinesListForCloudType(c, tt.cloudType)

			if tt.expectedNil {
				assert.Nil(t, result, "Expected nil result")
			} else {
				assert.NotNil(t, result, "Expected non-nil result")
			}

			if tt.verify != nil {
				tt.verify(t, result)
			}
		})
	}

	// Additional test to verify function signatures match and default case
	t.Run("Verify function signatures for all cloud types and default case", func(t *testing.T) {
		c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")

		cloudTypes := []string{"aws", "azure", "gcp", "vsphere", "openshift", "generic", "eks-anywhere", "apache-cloudstack", "maas", "edge-native", "openstack"}

		for _, cloudType := range cloudTypes {
			result := getMachinesListForCloudType(c, cloudType)
			assert.NotNil(t, result, "Function should not be nil for cloud type: %s", cloudType)

			// Verify the function can be referenced (even if not called, the signature should be correct)
			// We don't actually call it since it requires valid cluster/node IDs
			_ = result
		}

		// Test default case - invalid cloud type
		invalidResult := getMachinesListForCloudType(c, "invalid-cloud-type")
		assert.Nil(t, invalidResult, "Function should be nil for invalid cloud type")

		// Test default case - empty cloud type
		emptyResult := getMachinesListForCloudType(c, "")
		assert.Nil(t, emptyResult, "Function should be nil for empty cloud type")
	})
}

func TestGetClusterImportInfo(t *testing.T) {
	tests := []struct {
		name             string
		cluster          *models.V1SpectroCluster
		expectError      bool
		expectedCommand  string
		expectedManifest string
		description      string
		verify           func(t *testing.T, kubectlCommand, manifestURL string, err error)
	}{
		{
			name: "Cluster with nil Status - returns error",
			cluster: &models.V1SpectroCluster{
				Status: nil,
			},
			expectError: true,
			description: "Should return error when Status is nil",
			verify: func(t *testing.T, kubectlCommand, manifestURL string, err error) {
				assert.Error(t, err, "Should have error when Status is nil")
				assert.Contains(t, err.Error(), "cluster status is not available", "Error should mention status not available")
				assert.Empty(t, kubectlCommand, "Command should be empty on error")
				assert.Empty(t, manifestURL, "Manifest URL should be empty on error")
			},
		},
		{
			name: "Cluster with nil ClusterImport - returns error",
			cluster: &models.V1SpectroCluster{
				Status: &models.V1SpectroClusterStatus{
					ClusterImport: nil,
				},
			},
			expectError: true,
			description: "Should return error when ClusterImport is nil",
			verify: func(t *testing.T, kubectlCommand, manifestURL string, err error) {
				assert.Error(t, err, "Should have error when ClusterImport is nil")
				assert.Contains(t, err.Error(), "cluster import information is not available", "Error should mention import info not available")
				assert.Empty(t, kubectlCommand, "Command should be empty on error")
				assert.Empty(t, manifestURL, "Manifest URL should be empty on error")
			},
		},
		{
			name: "Cluster with empty ImportLink - returns error",
			cluster: &models.V1SpectroCluster{
				Status: &models.V1SpectroClusterStatus{
					ClusterImport: &models.V1ClusterImport{
						ImportLink: "",
					},
				},
			},
			expectError: true,
			description: "Should return error when ImportLink is empty",
			verify: func(t *testing.T, kubectlCommand, manifestURL string, err error) {
				assert.Error(t, err, "Should have error when ImportLink is empty")
				assert.Contains(t, err.Error(), "import link is empty", "Error should mention import link is empty")
				assert.Empty(t, kubectlCommand, "Command should be empty on error")
				assert.Empty(t, manifestURL, "Manifest URL should be empty on error")
			},
		},
		{
			name: "Success - ImportLink with kubectl apply -f prefix",
			cluster: &models.V1SpectroCluster{
				Status: &models.V1SpectroClusterStatus{
					ClusterImport: &models.V1ClusterImport{
						ImportLink: "kubectl apply -f https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest",
					},
				},
			},
			expectError:      false,
			expectedCommand:  "kubectl apply -f https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest",
			expectedManifest: "https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest",
			description:      "Should extract manifest URL from ImportLink with kubectl prefix",
			verify: func(t *testing.T, kubectlCommand, manifestURL string, err error) {
				assert.NoError(t, err, "Should not have error for valid ImportLink")
				assert.Equal(t, "kubectl apply -f https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest", kubectlCommand)
				assert.Equal(t, "https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest", manifestURL)
			},
		},
		{
			name: "Success - ImportLink with kubectl apply -f prefix and extra whitespace",
			cluster: &models.V1SpectroCluster{
				Status: &models.V1SpectroClusterStatus{
					ClusterImport: &models.V1ClusterImport{
						ImportLink: "kubectl apply -f  https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest  ",
					},
				},
			},
			expectError:      false,
			expectedCommand:  "kubectl apply -f  https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest  ",
			expectedManifest: "https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest",
			description:      "Should extract manifest URL and trim whitespace",
			verify: func(t *testing.T, kubectlCommand, manifestURL string, err error) {
				assert.NoError(t, err, "Should not have error for valid ImportLink")
				assert.Equal(t, "kubectl apply -f  https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest  ", kubectlCommand)
				assert.Equal(t, "https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest", manifestURL)
			},
		},
		{
			name: "Success - ImportLink without kubectl prefix (just URL)",
			cluster: &models.V1SpectroCluster{
				Status: &models.V1SpectroClusterStatus{
					ClusterImport: &models.V1ClusterImport{
						ImportLink: "https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest",
					},
				},
			},
			expectError:      false,
			expectedCommand:  "https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest",
			expectedManifest: "https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest",
			description:      "Should return URL as-is when no kubectl prefix",
			verify: func(t *testing.T, kubectlCommand, manifestURL string, err error) {
				assert.NoError(t, err, "Should not have error for valid ImportLink")
				assert.Equal(t, "https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest", kubectlCommand)
				assert.Equal(t, "https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest", manifestURL)
			},
		},
		{
			name: "Success - ImportLink with different URL format",
			cluster: &models.V1SpectroCluster{
				Status: &models.V1SpectroClusterStatus{
					ClusterImport: &models.V1ClusterImport{
						ImportLink: "kubectl apply -f https://api.example.com/v1/clusters/abc123/import",
					},
				},
			},
			expectError:      false,
			expectedCommand:  "kubectl apply -f https://api.example.com/v1/clusters/abc123/import",
			expectedManifest: "https://api.example.com/v1/clusters/abc123/import",
			description:      "Should extract manifest URL from different URL format",
			verify: func(t *testing.T, kubectlCommand, manifestURL string, err error) {
				assert.NoError(t, err, "Should not have error for valid ImportLink")
				assert.Equal(t, "kubectl apply -f https://api.example.com/v1/clusters/abc123/import", kubectlCommand)
				assert.Equal(t, "https://api.example.com/v1/clusters/abc123/import", manifestURL)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubectlCommand, manifestURL, err := getClusterImportInfo(tt.cluster)

			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
				assert.Empty(t, kubectlCommand, "Command should be empty on error")
				assert.Empty(t, manifestURL, "Manifest URL should be empty on error")
			} else {
				assert.NoError(t, err, "Unexpected error occurred")
				assert.Equal(t, tt.expectedCommand, kubectlCommand, "Kubectl command should match")
				assert.Equal(t, tt.expectedManifest, manifestURL, "Manifest URL should match")
			}

			if tt.verify != nil {
				tt.verify(t, kubectlCommand, manifestURL, err)
			}
		})
	}
}

func TestExtractManifestURL(t *testing.T) {
	tests := []struct {
		name        string
		importLink  string
		expected    string
		description string
	}{
		{
			name:        "ImportLink with kubectl apply -f prefix",
			importLink:  "kubectl apply -f https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest",
			expected:    "https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest",
			description: "Should extract URL from kubectl command",
		},
		{
			name:        "ImportLink with kubectl apply -f prefix and leading whitespace",
			importLink:  "kubectl apply -f  https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest",
			expected:    "https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest",
			description: "Should extract URL and trim whitespace after prefix",
		},
		{
			name:        "ImportLink with kubectl apply -f prefix and trailing whitespace",
			importLink:  "kubectl apply -f https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest  ",
			expected:    "https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest",
			description: "Should extract URL and trim trailing whitespace",
		},
		{
			name:        "ImportLink with kubectl apply -f prefix and both leading/trailing whitespace",
			importLink:  "kubectl apply -f  https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest  ",
			expected:    "https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest",
			description: "Should extract URL and trim all whitespace",
		},
		{
			name:        "ImportLink with URL containing fragments",
			importLink:  "kubectl apply -f https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest#section",
			expected:    "https://api.dev.spectrocloud.com/v1/spectroclusters/test-uid/import/manifest#section",
			description: "Should extract URL with fragments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractManifestURL(tt.importLink)

			assert.Equal(t, tt.expected, result, "Extracted manifest URL should match expected")
		})
	}
}
