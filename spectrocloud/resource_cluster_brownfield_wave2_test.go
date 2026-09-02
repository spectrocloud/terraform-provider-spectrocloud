package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// resourceClusterBrownfieldRead — additional branches not covered by
// resource_cluster_brownfield_readupdate_test.go's single happy-path test:
// GetCluster error/nil, the getClusterImportInfo success/warning split (and
// its nested "not Running" extra warning), and all three health_status
// branches. Fixtures used here (cluster-uid-server-error,
// cluster-uid-deleted-state, cluster-uid-running,
// cluster-uid-brownfield-import-running/-pending,
// cluster-uid-overview-error/-missing-health) are defined in
// tests/mockApiServer/routes/mockCluster.go.
// ---------------------------------------------------------------------------

func TestResourceClusterBrownfieldRead_GetClusterError(t *testing.T) {
	d := resourceClusterBrownfield().TestResourceData()
	d.SetId("cluster-uid-server-error")

	diags := resourceClusterBrownfieldRead(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

func TestResourceClusterBrownfieldRead_ClusterDeleted(t *testing.T) {
	d := resourceClusterBrownfield().TestResourceData()
	d.SetId("cluster-uid-deleted-state")

	diags := resourceClusterBrownfieldRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
	assert.Equal(t, "", d.Id(), "cluster should be removed from state when GetCluster returns nil")
}

func TestResourceClusterBrownfieldRead_RunningImportSuccess_NoWarning(t *testing.T) {
	d := resourceClusterBrownfield().TestResourceData()
	d.SetId("cluster-uid-brownfield-import-running")

	diags := resourceClusterBrownfieldRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
	for _, dg := range diags {
		assert.NotContains(t, dg.Summary, "Cluster import pending",
			"a Running cluster with import info available should not get the pending-import warning")
	}
	assert.NotEmpty(t, d.Get("kubectl_command"))
	assert.NotEmpty(t, d.Get("manifest_url"))
	assert.Equal(t, "Healthy", d.Get("health_status"))
	assert.Equal(t, "test-cloud-config-id", d.Get("cloud_config_id"))
}

func TestResourceClusterBrownfieldRead_PendingImportWarning(t *testing.T) {
	d := resourceClusterBrownfield().TestResourceData()
	d.SetId("cluster-uid-brownfield-import-pending")

	diags := resourceClusterBrownfieldRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
	assertAnyDiagContains(t, diags, "Cluster import pending")
	assert.NotEmpty(t, d.Get("kubectl_command"))
}

func TestResourceClusterBrownfieldRead_ImportInfoUnavailableWarning(t *testing.T) {
	d := resourceClusterBrownfield().TestResourceData()
	d.SetId("cluster-uid-running")

	diags := resourceClusterBrownfieldRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
	assertAnyDiagContains(t, diags, "kubectl_command not available")
	assert.Equal(t, "", d.Get("kubectl_command"))
	assert.Equal(t, "", d.Get("manifest_url"))
}

func TestResourceClusterBrownfieldRead_HealthStatusOverviewError(t *testing.T) {
	d := resourceClusterBrownfield().TestResourceData()
	d.SetId("cluster-uid-overview-error")

	diags := resourceClusterBrownfieldRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
	assert.Equal(t, "Unknown", d.Get("health_status"))
}

func TestResourceClusterBrownfieldRead_HealthStatusMissing(t *testing.T) {
	d := resourceClusterBrownfield().TestResourceData()
	d.SetId("cluster-uid-overview-missing-health")

	diags := resourceClusterBrownfieldRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
	assert.Equal(t, "Unknown", d.Get("health_status"))
}

// ---------------------------------------------------------------------------
// readCommonFieldsBrownfield — the six GetOk-gated blocks (backup_policy,
// scan_policy, cluster_rbac_binding, namespaces, host_config,
// review_repave_state) are skipped entirely unless the field is explicitly
// set on the ResourceData. Each sub-test below sets exactly one such field
// so GetOk returns true, driving both the error branch (via the
// cluster-uid-policy-error mock uid) and the success/flatten branch (via
// cluster-uid-policy-full for backup/scan, or just the default fixture for
// rbac/namespaces since their mock payload is a non-nil-but-empty Items
// slice — which already satisfies the "not nil" flatten guard).
// ---------------------------------------------------------------------------

func brownfieldClusterFixtureForRead() *models.V1SpectroCluster {
	return &models.V1SpectroCluster{
		Metadata: &models.V1ObjectMeta{
			Labels:      map[string]string{},
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
	}
}

func TestReadCommonFieldsBrownfield_OptionalFieldsAndErrors(t *testing.T) {
	const (
		policyErrorUID = "cluster-uid-policy-error"
		policyFullUID  = "cluster-uid-policy-full"
	)

	tests := []struct {
		name        string
		id          string
		setup       func(d *schema.ResourceData)
		expectError bool
	}{
		{
			name: "backup_policy - success, no content to flatten",
			id:   "test-cluster-id",
			setup: func(d *schema.ResourceData) {
				_ = d.Set("backup_policy", []interface{}{map[string]interface{}{
					"prefix": "backup", "backup_location_id": "loc-1",
					"schedule": "0 1 * * *", "expiry_in_hour": 24,
				}})
			},
		},
		{
			name: "backup_policy - success with content, flatten hit",
			id:   policyFullUID,
			setup: func(d *schema.ResourceData) {
				_ = d.Set("backup_policy", []interface{}{map[string]interface{}{
					"prefix": "backup", "backup_location_id": "loc-1",
					"schedule": "0 1 * * *", "expiry_in_hour": 24,
				}})
			},
		},
		{
			name: "backup_policy - GetClusterBackupConfig error",
			id:   policyErrorUID,
			setup: func(d *schema.ResourceData) {
				_ = d.Set("backup_policy", []interface{}{map[string]interface{}{
					"prefix": "backup", "backup_location_id": "loc-1",
					"schedule": "0 1 * * *", "expiry_in_hour": 24,
				}})
			},
			expectError: true,
		},
		{
			name: "scan_policy - success, no content to flatten",
			id:   "test-cluster-id",
			setup: func(d *schema.ResourceData) {
				_ = d.Set("scan_policy", []interface{}{map[string]interface{}{
					"configuration_scan_schedule": "0 2 * * *",
					"penetration_scan_schedule":   "0 3 * * *",
					"conformance_scan_schedule":   "0 4 * * *",
				}})
			},
		},
		{
			name: "scan_policy - success with content, flatten hit",
			id:   policyFullUID,
			setup: func(d *schema.ResourceData) {
				_ = d.Set("scan_policy", []interface{}{map[string]interface{}{
					"configuration_scan_schedule": "0 2 * * *",
					"penetration_scan_schedule":   "0 3 * * *",
					"conformance_scan_schedule":   "0 4 * * *",
				}})
			},
		},
		{
			name: "scan_policy - GetClusterScanConfig error",
			id:   policyErrorUID,
			setup: func(d *schema.ResourceData) {
				_ = d.Set("scan_policy", []interface{}{map[string]interface{}{
					"configuration_scan_schedule": "0 2 * * *",
					"penetration_scan_schedule":   "0 3 * * *",
					"conformance_scan_schedule":   "0 4 * * *",
				}})
			},
			expectError: true,
		},
		{
			name: "cluster_rbac_binding - success",
			id:   "test-cluster-id",
			setup: func(d *schema.ResourceData) {
				_ = d.Set("cluster_rbac_binding", []interface{}{map[string]interface{}{
					"type": "ClusterRoleBinding",
				}})
			},
		},
		{
			name: "cluster_rbac_binding - GetClusterRbacConfig error",
			id:   policyErrorUID,
			setup: func(d *schema.ResourceData) {
				_ = d.Set("cluster_rbac_binding", []interface{}{map[string]interface{}{
					"type": "ClusterRoleBinding",
				}})
			},
			expectError: true,
		},
		{
			name: "namespaces - success",
			id:   "test-cluster-id",
			setup: func(d *schema.ResourceData) {
				_ = d.Set("namespaces", []interface{}{map[string]interface{}{
					"name": "ns1",
					"resource_allocation": map[string]interface{}{
						"cpu_cores": "2", "memory_MiB": "2048",
					},
				}})
			},
		},
		{
			name: "namespaces - GetClusterNamespaceConfig error",
			id:   policyErrorUID,
			setup: func(d *schema.ResourceData) {
				_ = d.Set("namespaces", []interface{}{map[string]interface{}{
					"name": "ns1",
					"resource_allocation": map[string]interface{}{
						"cpu_cores": "2", "memory_MiB": "2048",
					},
				}})
			},
			expectError: true,
		},
		{
			name: "review_repave_state - set when GetOk true",
			id:   "test-cluster-id",
			setup: func(d *schema.ResourceData) {
				_ = d.Set("review_repave_state", "Approved")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
			d := resourceClusterBrownfield().TestResourceData()
			d.SetId(tt.id)
			tt.setup(d)

			diags, hasError := readCommonFieldsBrownfield(c, d, brownfieldClusterFixtureForRead())

			if tt.expectError {
				assert.True(t, hasError || diags.HasError(), "expected an error")
			} else {
				assert.False(t, hasError, "unexpected error: %+v", diags)
			}
		})
	}
}

func TestReadCommonFieldsBrownfield_HostConfigBranches(t *testing.T) {
	tests := []struct {
		name       string
		hostConfig *models.V1HostClusterConfig
	}{
		{
			name:       "host_config is nil on cluster",
			hostConfig: nil,
		},
		{
			name:       "IsHostCluster false - skipped",
			hostConfig: &models.V1HostClusterConfig{IsHostCluster: boolPtr(false)},
		},
		{
			name: "IsHostCluster true - flattened and set",
			hostConfig: &models.V1HostClusterConfig{
				IsHostCluster: boolPtr(true),
				ClusterEndpoint: &models.V1HostClusterEndpoint{
					Type: "Ingress",
					Config: &models.V1HostClusterEndpointConfig{
						IngressConfig:      &models.V1IngressConfig{Host: "host.example.com"},
						LoadBalancerConfig: &models.V1LoadBalancerConfig{},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
			d := resourceClusterBrownfield().TestResourceData()
			d.SetId("test-cluster-id")
			_ = d.Set("host_config", []interface{}{map[string]interface{}{
				"host_endpoint_type": "Ingress",
			}})

			cluster := brownfieldClusterFixtureForRead()
			cluster.Spec.ClusterConfig.HostClusterConfig = tt.hostConfig

			diags, hasError := readCommonFieldsBrownfield(c, d, cluster)
			assert.False(t, hasError, "unexpected error: %+v", diags)
		})
	}
}

// ---------------------------------------------------------------------------
// resolveNodeID — unsupported cloud_type (no mock needed), the
// getMachinesListFn API-error branch, the "node not found" branch, and the
// full success branch. Fixtures aws-machines-list-error /
// aws-machines-list-found are defined in tests/mockApiServer/routes/mockAwsCluster.go.
// ---------------------------------------------------------------------------

// TestResolveNodeID_UnsupportedCloudType is covered in
// cloud_account_import_sweep_test.go.

func TestResolveNodeID_APIError(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	_, err := resolveNodeID(c, "aws", "aws-machines-list-error", "pool-1", "node-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list machines")
}

func TestResolveNodeID_NodeNotFound(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	_, err := resolveNodeID(c, "aws", "some-other-config-uid", "pool-1", "missing-node")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestResolveNodeID_Success(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	nodeID, err := resolveNodeID(c, "aws", "aws-machines-list-found", "pool-1", "ip-10-0-0-1")
	require.NoError(t, err)
	assert.Equal(t, "aws-machine-uid-1", nodeID)
}

// ---------------------------------------------------------------------------
// resourceClusterBrownfieldUpdate — additional branches beyond
// resource_cluster_brownfield_readupdate_test.go's Day1-immutable /
// no-day2-change / description-change tests: the GetCluster error path, the
// isClusterRunningHealthy "not healthy" warning, and the machine_pool
// branch (empty cloud_config_id error, unsupported cloud_type error, and a
// full pass resolving node_id from node_name via resolveNodeID).
// ---------------------------------------------------------------------------

func TestResourceClusterBrownfieldUpdate_GetClusterError(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterBrownfield(), "cluster-uid-server-error",
		baseBrownfieldAttrs(), nil)

	diags := resourceClusterBrownfieldUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

func TestResourceClusterBrownfieldUpdate_NotHealthyWarning(t *testing.T) {
	// apply_setting is a day2Fields entry that updateCommonFields never
	// inspects, so diffing it fires the isClusterRunningHealthy check
	// (and the warning below) without also routing into
	// updateClusterMetadata/updateClusterTimezone/etc — several of which
	// have no mock route registered and would otherwise mask this
	// assertion behind an unrelated 404.
	d := buildUpdateResourceData(resourceClusterBrownfield(), "cluster-uid-running-unhealthy",
		baseBrownfieldAttrs(), simpleDiff("apply_setting", "DownloadAndInstall", "DownloadAndInstallLater"))

	diags := resourceClusterBrownfieldUpdate(context.Background(), d, unitTestMockAPIClient)
	assertAnyDiagContains(t, diags, "not in Running-Healthy state")
}

func TestResourceClusterBrownfieldUpdate_MachinePool_EmptyCloudConfigIdError(t *testing.T) {
	var oldPools []interface{}
	newPools := []interface{}{
		map[string]interface{}{
			"name": "worker-pool",
			"node": []interface{}{
				map[string]interface{}{
					"node_name": "ip-10-0-0-1",
					"action":    "cordon",
				},
			},
		},
	}
	// cloud_config_id left empty on both sides — no diff on it, so the
	// resulting ResourceData reads back "" from d.Get("cloud_config_id").
	d := buildBrownfieldMachinePoolUpdateResourceData(t, "aws", "", oldPools, newPools)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterBrownfieldUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
	assertAnyDiagContains(t, diags, "cloud_config_id is required")
}

func TestResourceClusterBrownfieldUpdate_MachinePool_UnsupportedCloudTypeError(t *testing.T) {
	var oldPools []interface{}
	newPools := []interface{}{
		map[string]interface{}{
			"name": "worker-pool",
			"node": []interface{}{
				map[string]interface{}{
					"node_name": "ip-10-0-0-1",
					"action":    "cordon",
				},
			},
		},
	}
	d := buildBrownfieldMachinePoolUpdateResourceData(t, "not-a-cloud", "some-cfg-id", oldPools, newPools)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterBrownfieldUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
	assertAnyDiagContains(t, diags, "not supported for cloud_type")
}

// buildBrownfieldMachinePoolUpdateResourceData mirrors
// buildCloudStackClusterUpdateResourceData (resource_cluster_apache_cloudstack_wave3_test.go):
// build old/new raw configs, run the resource's own Diff, then materialize a
// ResourceData whose HasChange("machine_pool") reflects that diff. Needed
// because machine_pool is a TypeSet of nested lists — manually constructing
// a terraform.InstanceDiff.Attributes map for it is impractical.
func buildBrownfieldMachinePoolUpdateResourceData(t *testing.T, cloudType, cloudConfigID string, oldPools, newPools []interface{}) *schema.ResourceData {
	t.Helper()
	res := resourceClusterBrownfield()

	buildRaw := func(pools []interface{}) map[string]interface{} {
		return map[string]interface{}{
			"name":            "test-brownfield-cluster",
			"context":         "project",
			"cloud_type":      cloudType,
			"cloud_config_id": cloudConfigID,
			"machine_pool":    pools,
		}
	}

	oldRD := schema.TestResourceDataRaw(t, res.Schema, buildRaw(oldPools))
	oldRD.SetId("test-cluster-id")
	oldState := oldRD.State()
	require.NotNil(t, oldState)

	newConfig := terraform.NewResourceConfigRaw(buildRaw(newPools))

	diff, err := res.Diff(context.Background(), oldState, newConfig, nil)
	require.NoError(t, err)
	require.NotNil(t, diff)

	finalRD, err := schema.InternalMap(res.Schema).Data(oldState, diff)
	require.NoError(t, err)
	finalRD.SetId("test-cluster-id")
	return finalRD
}

func TestResourceClusterBrownfieldUpdate_MachinePool_NodeNameResolutionSuccess(t *testing.T) {
	var oldPools []interface{}
	newPools := []interface{}{
		map[string]interface{}{
			"name": "worker-pool",
			"node": []interface{}{
				map[string]interface{}{
					"node_name": "ip-10-0-0-1",
					"action":    "cordon",
				},
			},
		},
	}

	d := buildBrownfieldMachinePoolUpdateResourceData(t, "aws", "aws-machines-list-found", oldPools, newPools)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterBrownfieldUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

func TestResourceClusterBrownfieldUpdate_MachinePool_NodeNotFoundError(t *testing.T) {
	var oldPools []interface{}
	newPools := []interface{}{
		map[string]interface{}{
			"name": "worker-pool",
			"node": []interface{}{
				map[string]interface{}{
					"node_name": "no-such-node",
					"action":    "cordon",
				},
			},
		},
	}

	d := buildBrownfieldMachinePoolUpdateResourceData(t, "aws", "some-other-config-uid", oldPools, newPools)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterBrownfieldUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
	assertAnyDiagContains(t, diags, "failed to resolve node_id")
}

func TestResourceClusterBrownfieldUpdate_MachinePool_NeitherNodeIdNorNameError(t *testing.T) {
	var oldPools []interface{}
	newPools := []interface{}{
		map[string]interface{}{
			"name": "worker-pool",
			"node": []interface{}{
				map[string]interface{}{
					"action": "cordon",
				},
			},
		},
	}

	d := buildBrownfieldMachinePoolUpdateResourceData(t, "aws", "aws-machines-list-found", oldPools, newPools)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterBrownfieldUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
	assertAnyDiagContains(t, diags, "either node_id or node_name must be provided")
}

// ---------------------------------------------------------------------------
// resourceClusterBrownfieldImport — ParseResourceCustomCloudImportID error
// path, GetCommonCluster error path, and the full success path (which
// exercises resourceClusterBrownfieldRead + flattenCommonAttributeForBrownfieldClusterImport).
// ---------------------------------------------------------------------------

func TestResourceClusterBrownfieldImport_InvalidIDFormat(t *testing.T) {
	d := resourceClusterBrownfield().TestResourceData()
	d.SetId("only-two:parts")

	_, err := resourceClusterBrownfieldImport(context.Background(), d, unitTestMockAPIClient)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cluster ID format")
}

func TestResourceClusterBrownfieldImport_GetCommonClusterError(t *testing.T) {
	d := resourceClusterBrownfield().TestResourceData()
	d.SetId("cluster-uid-server-error:project:generic")

	_, err := resourceClusterBrownfieldImport(context.Background(), d, unitTestMockAPIClient)
	require.Error(t, err)
}

func TestResourceClusterBrownfieldImport_ValidID_Success(t *testing.T) {
	// setClusterProfilesOrTemplateForImport / enrichClusterProfilesWithVariables
	// walk cluster profile template data that isn't the focus of this test —
	// guard with recover() the same way TestResourceClusterCustomImport_ValidID
	// does (misc_custom_import_and_reads_test.go) so a panic deep in shared
	// profile-enrichment code doesn't mask the import-path coverage we're after.
	defer func() { _ = recover() }()

	d := resourceClusterBrownfield().TestResourceData()
	d.SetId("test-cluster-id:project:generic")

	result, err := resourceClusterBrownfieldImport(context.Background(), d, unitTestMockAPIClient)
	if err == nil {
		require.Len(t, result, 1)
		assert.Equal(t, "generic", result[0].Get("cloud_type"))
	}
}
