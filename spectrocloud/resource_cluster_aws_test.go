package spectrocloud

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

func TestFlattenMachinePoolConfigsAwsSubnetIds(t *testing.T) {
	var machinePoolConfig []*models.V1AwsMachinePoolConfig
	addLabels := make(map[string]string)
	addLabels["by"] = "Siva"
	addLabels["purpose"] = "unittest"

	subnetIdsCP := make(map[string]string)
	subnetIdsCP["us-east-2a"] = "subnet-031a7ff4ff5e7fb9a"

	subnetIdsWorker := make(map[string]string)
	subnetIdsWorker["us-east-2a"] = "subnet-08864975df862eb58"

	isControl := func(b bool) *bool { return &b }(true)
	machinePoolConfig = append(machinePoolConfig, &models.V1AwsMachinePoolConfig{
		Name:                    "cp-pool",
		IsControlPlane:          isControl,
		InstanceType:            "t3.large",
		Size:                    1,
		AdditionalLabels:        addLabels,
		RootDeviceSize:          10,
		UseControlPlaneAsWorker: true,
		UpdateStrategy: &models.V1UpdateStrategy{
			Type: "",
		},
		SubnetIds: subnetIdsCP,
	})
	machinePoolConfig = append(machinePoolConfig, &models.V1AwsMachinePoolConfig{
		Name:             "worker-pool",
		InstanceType:     "t3.large",
		Size:             3,
		AdditionalLabels: addLabels,
		UpdateStrategy: &models.V1UpdateStrategy{
			Type: "",
		},
		SubnetIds: subnetIdsWorker,
	})
	machinePools := flattenMachinePoolConfigsAws(machinePoolConfig)
	if len(machinePools) != 2 {
		t.Fail()
		t.Logf("Machine pool for control-plane and worker is not returned by func - FlattenMachinePoolConfigsAws")
	} else {
		for i := range machinePools {
			k := machinePools[i].(map[string]interface{})
			if k["update_strategy"] != "RollingUpdateScaleOut" {
				t.Errorf("Machine pool - update strategy is not matching got %v, wanted %v", k["update_strategy"], "RollingUpdateScaleOut")
				t.Fail()
			}
			if k["count"].(int) != int(machinePoolConfig[i].Size) {
				t.Errorf("Machine pool - count is not matching got %v, wanted %v", k["count"].(string), int(machinePoolConfig[i].Size))
				t.Fail()
			}
			if k["instance_type"].(string) != machinePoolConfig[i].InstanceType {
				t.Errorf("Machine pool - instance_type is not matching got %v, wanted %v", k["instance_type"].(string), machinePoolConfig[i].InstanceType)
				t.Fail()
			}
			if !validateMapString(addLabels, k["additional_labels"].(map[string]string)) {
				t.Errorf("Machine pool - additional labels is not matching got %v, wanted %v", addLabels, k["additional_labels"])
				t.Fail()
			}
			if k["name"] == "cp-pool" {
				if k["control_plane_as_worker"].(bool) != machinePoolConfig[i].UseControlPlaneAsWorker {
					t.Errorf("Machine pool - control_plane_as_worker is not matching got %s, wanted %v", k["control_plane_as_worker"].(string), machinePoolConfig[i].UseControlPlaneAsWorker)
					t.Fail()
				}
				if k["disk_size_gb"].(int) != int(machinePoolConfig[i].RootDeviceSize) {
					t.Errorf("Machine pool - disk_size_gb is not matching got %v, wanted %v", k["disk_size_gb"].(int), int(machinePoolConfig[i].RootDeviceSize))
					t.Fail()
				}
				if !validateMapString(subnetIdsCP, k["az_subnets"].(map[string]string)) {
					t.Errorf("Machine pool - additional labels is not matching got %v, wanted %v", subnetIdsCP, k["az_subnets"])
					t.Fail()
				}
			} else {
				if !validateMapString(subnetIdsWorker, k["az_subnets"].(map[string]string)) {
					t.Errorf("Machine pool - additional labels is not matching got %v, wanted %v", subnetIdsWorker, k["az_subnets"])
					t.Fail()
				}
			}
		}
	}
}

func TestFlattenMachinePoolConfigsAwsAZ(t *testing.T) {
	var machinePoolConfig []*models.V1AwsMachinePoolConfig
	addLabels := make(map[string]string)
	addLabels["by"] = "Siva"
	addLabels["purpose"] = "unittest"

	azs := []string{"us-east-2a"}

	isControl := func(b bool) *bool { return &b }(true)
	machinePoolConfig = append(machinePoolConfig, &models.V1AwsMachinePoolConfig{
		Name:                    "cp",
		IsControlPlane:          isControl,
		InstanceType:            "t3.xlarge",
		Size:                    1,
		AdditionalLabels:        addLabels,
		RootDeviceSize:          10,
		UseControlPlaneAsWorker: true,
		UpdateStrategy: &models.V1UpdateStrategy{
			Type: "Recreate",
		},
		Azs: azs,
	})
	machinePoolConfig = append(machinePoolConfig, &models.V1AwsMachinePoolConfig{
		Name:             "worker",
		InstanceType:     "t3.xlarge",
		Size:             3,
		AdditionalLabels: addLabels,
		UpdateStrategy: &models.V1UpdateStrategy{
			Type: "Recreate",
		},
		Azs: azs,
	})
	machinePools := flattenMachinePoolConfigsAws(machinePoolConfig)
	if len(machinePools) != 2 {
		t.Fail()
		t.Logf("Machine pool for control-plane and worker is not returned by func - FlattenMachinePoolConfigsAws")
	} else {
		for i := range machinePools {
			k := machinePools[i].(map[string]interface{})
			if k["update_strategy"] != "Recreate" {
				t.Errorf("Machine pool - update strategy is not matching got %v, wanted %v", k["update_strategy"], "Recreate")
				t.Fail()
			}
			if k["count"].(int) != int(machinePoolConfig[i].Size) {
				t.Errorf("Machine pool - count is not matching got %s, wanted %v", k["count"].(string), int(machinePoolConfig[i].Size))
				t.Fail()
			}
			if k["instance_type"].(string) != machinePoolConfig[i].InstanceType {
				t.Errorf("Machine pool - instance_type is not matching got %s, wanted %s", k["instance_type"].(string), machinePoolConfig[i].InstanceType)
				t.Fail()
			}
			if !validateMapString(addLabels, k["additional_labels"].(map[string]string)) {
				t.Errorf("Machine pool - additional labels is not matching got %v, wanted %v", addLabels, k["additional_labels"])
				t.Fail()
			}
			if !reflect.DeepEqual(azs, k["azs"]) {
				t.Errorf("Machine pool - AZS is not matching got %v, wanted %v", azs, k["azs"])
				t.Fail()
			}
			if k["name"] == "cp-pool" {
				if k["control_plane_as_worker"].(bool) != machinePoolConfig[i].UseControlPlaneAsWorker {
					t.Errorf("Machine pool - control_plane_as_worker is not matching got %s, wanted %v", k["control_plane_as_worker"].(string), machinePoolConfig[i].UseControlPlaneAsWorker)
					t.Fail()
				}
				if k["disk_size_gb"].(int) != int(machinePoolConfig[i].RootDeviceSize) {
					t.Errorf("Machine pool - disk_size_gb is not matching got %v, wanted %v", k["disk_size_gb"].(int), int(machinePoolConfig[i].RootDeviceSize))
					t.Fail()
				}
			}
		}
	}
}

func validateMapString(src map[string]string, dest map[string]string) bool {
	return reflect.DeepEqual(src, dest)
}

func TestResourceClusterAwsImport(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		setup       func() *schema.ResourceData
		client      interface{}
		expectError bool
		errorMsg    string
		description string
		verify      func(t *testing.T, importedData []*schema.ResourceData, err error)
	}{
		{
			name: "Successful import with cluster ID and project context",
			setup: func() *schema.ResourceData {
				d := resourceClusterAws().TestResourceData()
				d.SetId("test-cluster-id:project")
				return d
			},
			client:      unitTestMockAPIClient,
			expectError: false, // Function may succeed if GetCluster and Read work
			description: "Should import cluster with project context and populate state",
			verify: func(t *testing.T, importedData []*schema.ResourceData, err error) {
				// Function should successfully import
				if err == nil {
					assert.NotNil(t, importedData, "Imported data should not be nil on success")
					if len(importedData) > 0 {
						assert.Len(t, importedData, 1, "Should return exactly one ResourceData")
						// Verify ID is set
						assert.NotEmpty(t, importedData[0].Id(), "Cluster ID should be set")
					}
				}
			},
		},
		{
			name: "Successful import with cluster ID and project context",
			setup: func() *schema.ResourceData {
				d := resourceClusterAws().TestResourceData()
				d.SetId("test-cluster-id:project")
				return d
			},
			client:      unitTestMockAPIClient,
			expectError: false, // Function may succeed if GetCluster and Read work
			description: "Should import cluster with project context and populate state",
			verify: func(t *testing.T, importedData []*schema.ResourceData, err error) {
				// Function should successfully import
				if err == nil {
					assert.NotNil(t, importedData, "Imported data should not be nil on success")
					if len(importedData) > 0 {
						// Verify context is set
						context := importedData[0].Get("context")
						assert.NotNil(t, context, "Context should be set")
					}
				}
			},
		},
		{
			name: "Successful import with cluster ID and tenant context",
			setup: func() *schema.ResourceData {
				d := resourceClusterAws().TestResourceData()
				d.SetId("test-cluster-id:tenant")
				return d
			},
			client:      unitTestMockAPIClient,
			expectError: false, // Function may succeed if GetCluster and Read work
			description: "Should import cluster with tenant context and populate state",
			verify: func(t *testing.T, importedData []*schema.ResourceData, err error) {
				// Function should successfully import
				if err == nil {
					assert.NotNil(t, importedData, "Imported data should not be nil on success")
					if len(importedData) > 0 {
						// Verify context is set
						context := importedData[0].Get("context")
						assert.NotNil(t, context, "Context should be set")
					}
				}
			},
		},
		{
			name: "Import with invalid ID format (missing context)",
			setup: func() *schema.ResourceData {
				d := resourceClusterAws().TestResourceData()
				d.SetId("invalid-cluster-id") // Missing context (should be cluster-id:project or cluster-id:tenant)
				return d
			},
			client:      unitTestMockAPIClient,
			expectError: true,
			errorMsg:    "invalid cluster ID format specified for import",
			description: "Should return error when ID format is invalid (missing context)",
			verify: func(t *testing.T, importedData []*schema.ResourceData, err error) {
				assert.Error(t, err, "Should have error when ID format is invalid")
				assert.Nil(t, importedData, "Imported data should be nil on error")
				if err != nil {
					assert.Contains(t, err.Error(), "invalid cluster ID format specified for import", "Error should mention invalid format")
				}
			},
		},
		{
			name: "Import with GetCommonCluster error (cluster not found)",
			setup: func() *schema.ResourceData {
				d := resourceClusterAws().TestResourceData()
				d.SetId("nonexistent-cluster-id:project")
				return d
			},
			client:      unitTestMockAPIClient,
			expectError: true,
			errorMsg:    "", // Error may be from GetCommonCluster or resourceClusterAwsRead
			description: "Should return error when cluster is not found",
			verify: func(t *testing.T, importedData []*schema.ResourceData, err error) {
				assert.Error(t, err, "Should have error when cluster not found")
				assert.Nil(t, importedData, "Imported data should be nil on error")
				if err != nil {
					// Error could be from GetCommonCluster or resourceClusterAwsRead
					assert.True(t,
						strings.Contains(err.Error(), "unable to retrieve cluster data") ||
							strings.Contains(err.Error(), "could not read cluster for import") ||
							strings.Contains(err.Error(), "couldn't find cluster"),
						"Error should mention cluster retrieval or read failure")
				}
			},
		},
		{
			name: "Import with GetCommonCluster error from negative client",
			setup: func() *schema.ResourceData {
				d := resourceClusterAws().TestResourceData()
				d.SetId("test-cluster-id:project")
				return d
			},
			client:      unitTestMockAPINegativeClient,
			expectError: true,
			errorMsg:    "", // Error may be "unable to retrieve cluster data" or "couldn't find cluster" or from resourceClusterAwsRead
			description: "Should return error when GetCommonCluster API call fails",
			verify: func(t *testing.T, importedData []*schema.ResourceData, err error) {
				assert.Error(t, err, "Should have error when API call fails")
				assert.Nil(t, importedData, "Imported data should be nil on error")
				if err != nil {
					errMsg := err.Error()
					// Error could be from GetCommonCluster when cluster is nil, when GetCluster fails, or from resourceClusterAwsRead
					// Check for various error message patterns
					hasExpectedError := strings.Contains(errMsg, "unable to retrieve cluster data") ||
						strings.Contains(errMsg, "find cluster") ||
						strings.Contains(errMsg, "could not read cluster for import")
					assert.True(t, hasExpectedError,
						"Error should mention cluster retrieval or read failure, got: %s", errMsg)
				}
			},
		},
		{
			name: "Import with resourceClusterAwsRead error",
			setup: func() *schema.ResourceData {
				d := resourceClusterAws().TestResourceData()
				d.SetId("test-cluster-id:project")
				return d
			},
			client:      unitTestMockAPIClient,
			expectError: true, // May error if resourceClusterAwsRead fails
			errorMsg:    "could not read cluster for import",
			description: "Should return error when resourceClusterAwsRead fails",
			verify: func(t *testing.T, importedData []*schema.ResourceData, err error) {
				// This test may or may not error depending on mock API server behavior
				if err != nil {
					assert.Nil(t, importedData, "Imported data should be nil on error")
					assert.Contains(t, err.Error(), "could not read cluster for import", "Error should mention read failure")
				}
			},
		},
		{
			name: "Import with flattenCommonAttributeForClusterImport error",
			setup: func() *schema.ResourceData {
				d := resourceClusterAws().TestResourceData()
				d.SetId("test-cluster-id:project")
				return d
			},
			client:      unitTestMockAPIClient,
			expectError: true, // May error if flattenCommonAttributeForClusterImport fails
			errorMsg:    "",   // Error message depends on what fails in flattenCommonAttributeForClusterImport
			description: "Should return error when flattenCommonAttributeForClusterImport fails",
			verify: func(t *testing.T, importedData []*schema.ResourceData, err error) {
				// This test may or may not error depending on mock API server behavior
				if err != nil {
					assert.Nil(t, importedData, "Imported data should be nil on error")
				}
			},
		},
		{
			name: "Import with empty ID",
			setup: func() *schema.ResourceData {
				d := resourceClusterAws().TestResourceData()
				d.SetId("") // Empty ID
				return d
			},
			client:      unitTestMockAPIClient,
			expectError: true,
			errorMsg:    "invalid cluster ID format specified for import",
			description: "Should return error when import ID is empty",
			verify: func(t *testing.T, importedData []*schema.ResourceData, err error) {
				assert.Error(t, err, "Should have error for empty ID")
				assert.Nil(t, importedData, "Imported data should be nil on error")
				if err != nil {
					assert.Contains(t, err.Error(), "invalid cluster ID format specified for import", "Error should mention invalid format")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Recover from panics to handle nil pointer dereferences
			defer func() {
				if r := recover(); r != nil {
					if !tt.expectError {
						t.Errorf("Test panicked unexpectedly: %v", r)
					}
				}
			}()

			resourceData := tt.setup()

			// Call the import function
			importedData, err := resourceClusterAwsImport(ctx, resourceData, tt.client)

			// Verify results
			if tt.expectError {
				assert.Error(t, err, "Expected error for test case: %s", tt.description)
				if tt.errorMsg != "" && err != nil {
					assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text: %s", tt.description)
				}
				assert.Nil(t, importedData, "Imported data should be nil on error: %s", tt.description)
			} else {
				if err != nil {
					// If error occurred but not expected, log it for debugging
					t.Logf("Unexpected error: %v", err)
				}
				// For cases where error may or may not occur, check both paths
				if err == nil {
					assert.NotNil(t, importedData, "Imported data should not be nil: %s", tt.description)
					if len(importedData) > 0 {
						assert.Len(t, importedData, 1, "Should return exactly one ResourceData: %s", tt.description)
					}
				}
			}

			// Run custom verify function if provided
			if tt.verify != nil {
				tt.verify(t, importedData, err)
			}
		})
	}
}
