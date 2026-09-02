package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file consolidates the malformed-input and empty-ID guard tests for
// every resource_*_import.go entry point. Historically these files sat at
// 0% coverage because their parse-and-lookup branches were never exercised.
//
// Convention:
//   - Table-driven where the resources share a pattern (cluster imports
//     all funnel through ParseResourceID).
//   - Individual t.Run blocks for the resources with idiosyncratic error
//     messages (team, filter, ssh_key, etc.) so the assertion strings
//     stay close to the code they pin.
//
// Happy-path coverage for imports lives in resource_import_workflows_test.go
// (already present) and in each resource's dedicated CRUD test — this file
// only pins the error branches, which is where the code diverges most and
// where regressions are easiest to hide.

// ---------------------------------------------------------------------------
// Cluster imports — all use ParseResourceID (uid:context format)
// ---------------------------------------------------------------------------

// clusterImportFunc mirrors the signature of every resourceCluster*Import.
type clusterImportFunc func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error)

// clusterImportCase describes a single cluster-import target. The resource
// pointer lets us build a schema-correct TestResourceData without pulling
// the whole provider() map — we only need the function pointer for the
// import call.
type clusterImportCase struct {
	name      string
	newRes    func() *schema.Resource
	importFn  clusterImportFunc
	invalidID string
	badMsg    string
	validID   string
}

func TestClusterImports_InvalidFormat(t *testing.T) {
	// Every cluster import goes through GetCommonCluster → ParseResourceID,
	// which rejects any ID that doesn't split into exactly two parts with
	// the second being "project" or "tenant". A single malformed-ID input
	// is enough to prove the guard fires — no per-cloud variation.
	cases := []clusterImportCase{
		{"aks", resourceClusterAks, resourceClusterAksImport, "no-colon", "invalid cluster ID format", ""},
		{"aws", resourceClusterAws, resourceClusterAwsImport, "no-colon", "invalid cluster ID format", ""},
		{"azure", resourceClusterAzure, resourceClusterAzureImport, "no-colon", "invalid cluster ID format", ""},
		{"edge_vsphere", resourceClusterEdgeVsphere, resourceClusterEdgeVsphereImport, "no-colon", "invalid cluster ID format", ""},
		{"eks", resourceClusterEks, resourceClusterEksImport, "no-colon", "invalid cluster ID format", ""},
		{"gcp", resourceClusterGcp, resourceClusterGcpImport, "no-colon", "invalid cluster ID format", ""},
		{"gke", resourceClusterGke, resourceClusterGkeImport, "no-colon", "invalid cluster ID format", ""},
		{"maas", resourceClusterMaas, resourceClusterMaasImport, "no-colon", "invalid cluster ID format", ""},
		// virtual is intentionally omitted — resourceClusterVirtualImport
		// does NOT go through ParseResourceID; it dereferences the whole
		// spec chain (Spec.ClusterConfig.HostClusterConfig...) which our
		// mock leaves nil. Covered separately below via the empty-ID guard.
		{"vsphere", resourceClusterVsphere, resourceClusterVsphereImport, "no-colon", "invalid cluster ID format", ""},
		{"edge_native", resourceClusterEdgeNative, resourceClusterEdgeNativeImport, "no-colon", "invalid cluster ID format", ""},
		// cluster_custom_cloud takes an extra string arg — tested separately.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := tc.newRes().TestResourceData()
			d.SetId(tc.invalidID)
			_, err := tc.importFn(context.Background(), d, unitTestMockAPIClient)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.badMsg)
		})
	}
}

func TestClusterVirtualImport_EmptyID(t *testing.T) {
	// The virtual cluster importer doesn't go through the shared
	// ParseResourceID — it just checks d.Id() is non-empty. Cover the
	// guard, which is the only reachable error branch without a fully
	// nested SpectroCluster payload from the mock.
	d := resourceClusterVirtual().TestResourceData()
	d.SetId("")
	_, err := resourceClusterVirtualImport(context.Background(), d, unitTestMockAPIClient)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestClusterImports_WrongScope(t *testing.T) {
	// Second half of ParseResourceID's guard: two-part ID with a scope
	// that isn't "project" or "tenant". Same error message but a distinct
	// input shape — worth pinning both branches so a future refactor
	// that accepts new scopes doesn't silently break older ones.
	d := resourceClusterAws().TestResourceData()
	d.SetId("uid-abc:staging") // valid split but "staging" is not accepted
	_, err := resourceClusterAwsImport(context.Background(), d, unitTestMockAPIClient)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cluster ID format")
}

// TestClusterCustomCloudImport_InvalidFormat covers the odd sibling —
// resourceClusterCustomImport calls its own ParseResourceCustomCloudImportID.
// A malformed ID exercises that guard.
func TestClusterCustomCloudImport_InvalidFormat(t *testing.T) {
	d := resourceClusterCustomCloud().TestResourceData()
	d.SetId("bad-id")
	_, err := resourceClusterCustomImport(context.Background(), d, unitTestMockAPIClient)
	require.Error(t, err)
}

// (Direct ParseResourceID tests already live in
// resource_cluster_edge_native_import_test.go — no need to duplicate.)

// ---------------------------------------------------------------------------
// Non-cluster resource imports — one t.Run per resource, error-branch only
// ---------------------------------------------------------------------------

func TestResourceTeamImport(t *testing.T) {
	t.Run("uid lookup succeeds", func(t *testing.T) {
		// Mock returns a fixed team for any GET /v1/teams/{uid}, so any
		// non-empty ID exercises the happy path.
		d := resourceTeam().TestResourceData()
		d.SetId("team-123")
		got, err := resourceTeamImport(context.Background(), d, unitTestMockAPIClient)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "team-name", got[0].Get("name"))
	})
}

func TestResourceFilterImport(t *testing.T) {
	t.Run("empty id errors", func(t *testing.T) {
		d := resourceFilter().TestResourceData()
		d.SetId("")
		_, err := resourceFilterImport(context.Background(), d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("uid lookup succeeds", func(t *testing.T) {
		d := resourceFilter().TestResourceData()
		d.SetId("filter-123")
		got, err := resourceFilterImport(context.Background(), d, unitTestMockAPIClient)
		require.NoError(t, err)
		require.Len(t, got, 1)
	})
}

func TestResourceSSHKeyImport(t *testing.T) {
	t.Run("empty id errors", func(t *testing.T) {
		d := resourceSSHKey().TestResourceData()
		d.SetId("")
		_, err := resourceSSHKeyImport(context.Background(), d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("invalid context errors", func(t *testing.T) {
		d := resourceSSHKey().TestResourceData()
		d.SetId("some-name:staging") // not project|tenant
		_, err := resourceSSHKeyImport(context.Background(), d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid context")
	})

	t.Run("bare id succeeds", func(t *testing.T) {
		// Single-part ID → defaults to project context. The mock's
		// SSHKeyRoutes GET returns a valid key for any UID.
		d := resourceSSHKey().TestResourceData()
		d.SetId("some-ssh-uid")
		got, err := resourceSSHKeyImport(context.Background(), d, unitTestMockAPIClient)
		require.NoError(t, err)
		require.Len(t, got, 1)
	})
}

func TestResourceRegistryHelmImport(t *testing.T) {
	t.Run("empty id errors", func(t *testing.T) {
		d := resourceRegistryHelm().TestResourceData()
		d.SetId("")
		_, err := resourceRegistryHelmImport(context.Background(), d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})
}

func TestResourceRegistryOciImport(t *testing.T) {
	t.Run("empty id errors", func(t *testing.T) {
		d := resourceRegistryOciEcr().TestResourceData()
		d.SetId("")
		_, err := resourceRegistryOciImport(context.Background(), d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})
}

func TestResourcePCGDNSMapImport(t *testing.T) {
	t.Run("empty id errors", func(t *testing.T) {
		d := resourcePrivateCloudGatewayDNSMap().TestResourceData()
		d.SetId("")
		_, err := resourcePrivateCloudGatewayDNSMapImport(context.Background(), d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("malformed id errors", func(t *testing.T) {
		d := resourcePrivateCloudGatewayDNSMap().TestResourceData()
		d.SetId("only-one-part") // needs pcg:dnsmap format
		_, err := resourcePrivateCloudGatewayDNSMapImport(context.Background(), d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid import ID format")
	})
}

func TestResourcePCGIpPoolImport(t *testing.T) {
	t.Run("empty id errors", func(t *testing.T) {
		d := resourcePrivateCloudGatewayIpPool().TestResourceData()
		d.SetId("")
		_, err := resourcePrivateCloudGatewayIpPoolImport(context.Background(), d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("malformed id errors", func(t *testing.T) {
		d := resourcePrivateCloudGatewayIpPool().TestResourceData()
		d.SetId("only-one-part")
		_, err := resourcePrivateCloudGatewayIpPoolImport(context.Background(), d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid import ID format")
	})
}

// (Appliance import coverage lives in resource_appliance_test.go — already
// exhaustive with a table-driven test hitting empty-ID, invalid-ID, and
// success branches.)

func TestResourceProjectImport(t *testing.T) {
	// Extends the single "empty ID" case in resource_import_workflows_test.go
	// with the "project not found" branch — worth pinning because the
	// project resource is used everywhere and a regression that treats
	// nil-not-found as success would corrupt state on import.
	t.Run("empty id errors", func(t *testing.T) {
		d := resourceProject().TestResourceData()
		d.SetId("")
		_, err := resourceProjectImport(context.Background(), d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})
}

func TestResourceClusterGroupImport(t *testing.T) {
	// resource_import_workflows_test.go already covers invalid-format and
	// uid-success. Round it out with the wrong-scope branch — the current
	// two-branch guard ("2 parts AND scope must be project/tenant")
	// should reject "uid:staging" the same as "uid" alone.
	t.Run("wrong scope errors", func(t *testing.T) {
		d := resourceClusterGroup().TestResourceData()
		d.SetId("some-cg:staging")
		_, err := resourceClusterGroupImport(context.Background(), d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid import ID format")
	})
}

// ---------------------------------------------------------------------------
// Cloud account imports — resourceAccountAwsImport / resourceAccountGcpImport
// share GetCommonAccount, which validates that d.Id() and the cloud kind
// resolve. Empty ID → error.
// ---------------------------------------------------------------------------

func TestResourceAccountAwsImport_EmptyID(t *testing.T) {
	d := resourceCloudAccountAws().TestResourceData()
	d.SetId("")
	_, err := resourceAccountAwsImport(context.Background(), d, unitTestMockAPIClient)
	require.Error(t, err)
}

func TestResourceAccountGcpImport_EmptyID(t *testing.T) {
	d := resourceCloudAccountGcp().TestResourceData()
	d.SetId("")
	_, err := resourceAccountGcpImport(context.Background(), d, unitTestMockAPIClient)
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Miscellaneous — user, alert already covered; document what's not here.
// ---------------------------------------------------------------------------

// Left OUT (intentionally):
//   - resourceUserImport — covered by TestResourceUserWorkspaceRoleMapping*
//     et al. flat schema tests + resource_user_test.go covers the happy path.
//   - resourceAlertImport — covered in resource_import_workflows_test.go.
//   - resourceApplicationImport — covered by TestGetCommonApplication in
//     resource_import_workflows_test.go.
//   - resourceApplicationProfileImport / resourceBackupStorageLocationImport /
//     resourceClusterProfileImport — same file, GetCommon* variants.
//   - resourceRegistrationTokenImport — covered in the token CRUD tests
//     added earlier.
//   - resourcePasswordPolicyImport / resourceDeveloperSettingImport /
//     resourceResourceLimitsImport / resourceAuditTrailImport — covered
//     in their respective *_test.go files.
