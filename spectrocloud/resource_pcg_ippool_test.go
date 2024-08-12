package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-api-go/models"
	"github.com/stretchr/testify/assert"
	"testing"
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
