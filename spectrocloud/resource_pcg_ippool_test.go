package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

func TestToIpPool(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1IPPoolInputEntity
	}{
		{
			name: "IP pool with range type",
			input: map[string]interface{}{
				"name":                       "test-pool",
				"gateway":                    "192.168.1.1",
				"prefix":                     24,
				"network_type":               "range",
				"ip_start_range":             "192.168.1.10",
				"ip_end_range":               "192.168.1.100",
				"nameserver_addresses":       []interface{}{"8.8.8.8", "8.8.4.4"},
				"nameserver_search_suffix":   []interface{}{"example.com", "sub.example.com"},
				"restrict_to_single_cluster": true,
			},
			expected: &models.V1IPPoolInputEntity{
				Metadata: &models.V1ObjectMeta{
					Name: "test-pool",
					UID:  "1234", // Example UID, should be set accordingly
				},
				Spec: &models.V1IPPoolInputEntitySpec{
					Pool: &models.V1Pool{
						Gateway: "192.168.1.1",
						Nameserver: &models.V1Nameserver{
							Addresses: []string{"8.8.4.4", "8.8.8.8"},
							Search:    []string{"example.com", "sub.example.com"},
						},
						Prefix: 24,
						Start:  "192.168.1.10",
						End:    "192.168.1.100",
					},
					RestrictToSingleCluster: true,
				},
			},
		},
		{
			name: "IP pool with subnet type",
			input: map[string]interface{}{
				"name":                       "test-pool",
				"gateway":                    "192.168.2.1",
				"prefix":                     24,
				"network_type":               "subnet",
				"subnet_cidr":                "192.168.2.0/24",
				"nameserver_addresses":       []interface{}{"1.1.1.1", "1.0.0.1"},
				"nameserver_search_suffix":   []interface{}{"example.org"},
				"restrict_to_single_cluster": false,
			},
			expected: &models.V1IPPoolInputEntity{
				Metadata: &models.V1ObjectMeta{
					Name: "test-pool",
					UID:  "1234", // Example UID, should be set accordingly
				},
				Spec: &models.V1IPPoolInputEntitySpec{
					Pool: &models.V1Pool{
						Gateway: "192.168.2.1",
						Nameserver: &models.V1Nameserver{
							Addresses: []string{"1.1.1.1", "1.0.0.1"},
							Search:    []string{"example.org"},
						},
						Prefix: 24,
						Subnet: "192.168.2.0/24",
					},
					RestrictToSingleCluster: false,
				},
			},
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up schema.ResourceData
			d := schema.TestResourceDataRaw(t, resourcePrivateCloudGatewayIpPool().Schema, tc.input)
			d.SetId("1234") // Set the UID as needed

			// Call the function
			result := toIpPool(d)

			// Compare the results
			assert.Equal(t, tc.expected, result)
		})
	}
}

func prepareResourcePrivateCloudGatewayIpPool() *schema.ResourceData {
	d := resourcePrivateCloudGatewayIpPool().TestResourceData()
	d.SetId("test-pcg-id")
	_ = d.Set("name", "test-ippool")
	_ = d.Set("private_cloud_gateway_id", "test-pcg-id")
	_ = d.Set("network_type", "subnet")
	_ = d.Set("ip_start_range", "121.0.0.1")
	_ = d.Set("ip_end_range", "121.0.0.100")
	_ = d.Set("subnet_cidr", "test-subnet-cidr")
	_ = d.Set("prefix", 0)
	_ = d.Set("gateway", "test-gateway")
	_ = d.Set("nameserver_addresses", []string{"test.test.cm"})
	_ = d.Set("nameserver_search_suffix", []string{"test-suffix"})
	_ = d.Set("restrict_to_single_cluster", false)
	return d
}

func TestResourceIpPoolCRUD(t *testing.T) {
	testResourceCRUD(t, prepareResourcePrivateCloudGatewayIpPool, unitTestMockAPIClient,
		resourceIpPoolCreate, resourceIpPoolRead, resourceIpPoolUpdate, resourceIpPoolDelete)
}
func TestResourceIpPoolReadRange(t *testing.T) {
	d := prepareResourcePrivateCloudGatewayIpPool()
	ctx := context.Background()
	diags := resourceIpPoolRead(ctx, d, unitTestMockAPIClient)
	_ = d.Set("network_type", "range")
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-pcg-id", d.Id())
}

func TestValidateNetworkType(t *testing.T) {
	tests := []struct {
		name           string
		input          interface{}
		expectedDiags  diag.Diagnostics
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:          "Valid network type - range",
			input:         "range",
			expectedDiags: diag.Diagnostics{},
			expectedError: false,
		},
		{
			name:          "Valid network type - subnet",
			input:         "subnet",
			expectedDiags: diag.Diagnostics{},
			expectedError: false,
		},
		{
			name:           "Invalid network type - random",
			input:          "random",
			expectedError:  true,
			expectedErrMsg: "network type 'random' is invalid. valid network types are 'range' and 'subnet'",
		},
		{
			name:           "Invalid network type - empty string",
			input:          "",
			expectedError:  true,
			expectedErrMsg: "network type '' is invalid. valid network types are 'range' and 'subnet'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := validateNetworkType(tt.input, cty.Path{})

			if tt.expectedError {
				assert.NotEmpty(t, diags)
				assert.Equal(t, diag.Error, diags[0].Severity)
				assert.Equal(t, tt.expectedErrMsg, diags[0].Summary)
			} else {
				assert.Empty(t, diags)
			}
		})
	}
}
