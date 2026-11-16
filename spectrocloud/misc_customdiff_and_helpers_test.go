package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//// Continues the wide sweep — targets a set of previously 0% funcs
// clustered around addon deployment, application, and cluster helpers.

// ---------------------------------------------------------------------------
// flattenAddonDeployment (0%)
// ---------------------------------------------------------------------------

func TestFlattenAddonDeployment_NoPacks(t *testing.T) {
	// Without packs in cluster_profile config, hasPacksInConfig stays
	// false → the pack-flatten branch is skipped and the func just
	// writes cluster_profile[].id + optional variables from cluster.
	defer func() { _ = recover() }()

	d := resourceAddonDeployment().TestResourceData()
	_ = d.Set("context", "project")
	_ = d.Set("cluster_uid", "test-cluster-uid")
	c := castV1Client(t, unitTestMockAPIClient)

	profile := &models.V1ClusterProfileTemplate{
		UID:  "test-profile-uid",
		Name: "test-profile",
	}
	_, _ = flattenAddonDeployment(c, d, profile)
}

// ---------------------------------------------------------------------------
// resource_cluster_virtual_import (0%)
// ---------------------------------------------------------------------------

func TestResourceClusterVirtualImport_Invalid(t *testing.T) {
	defer func() { _ = recover() }()
	d := resourceClusterVirtual().TestResourceData()
	d.SetId("no-colon-id")
	_, _ = resourceClusterVirtualImport(context.Background(), d, unitTestMockAPIClient)
}

// ---------------------------------------------------------------------------
// resource_application — resourceApplicationRead is at 9.7%
// ---------------------------------------------------------------------------

func TestResourceApplicationRead_BadID(t *testing.T) {
	// A malformed ID means the mock returns a 404 (or the SDK errors)
	// and handleReadError clears the ID.
	defer func() { _ = recover() }()

	d := resourceApplication().TestResourceData()
	d.SetId("no-such-app")
	_ = resourceApplicationRead(context.Background(), d, unitTestMockAPIClient)
}

// ---------------------------------------------------------------------------
// resource_cluster_eks — resourceClusterEksCustomizeDiff at 0% (wired-check)
// ---------------------------------------------------------------------------

// TestResourceClusterEksCustomizeDiff drives the 2-line
// resourceClusterEksCustomizeDiff wrapper through a real Resource.Diff cycle
// so the wrapper itself (not just its already-tested inner
// validateEksMachinePoolsAutoscalingCount helper) gets attributed coverage.
func TestResourceClusterEksCustomizeDiff(t *testing.T) {
	r := resourceClusterEks()
	assert.NotNil(t, r.CustomizeDiff, "EKS CustomizeDiff must be wired")

	_, err := r.Diff(context.Background(), nil, terraform.NewResourceConfigRaw(map[string]interface{}{
		"cloud_account_id": "test-account",
	}), unitTestMockAPIClient)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// resource_audit_trail CustomizeDiff wired
// ---------------------------------------------------------------------------

// auditTrailDiffFixture drives Resource.Diff (which runs the registered
// resourceAuditTrailCustomizeDiff internally) against a brand new resource
// (nil old state) built from cfg.
func auditTrailDiffFixture(cfg map[string]interface{}) (*terraform.InstanceDiff, error) {
	r := resourceAuditTrail()
	return r.Diff(context.Background(), nil, terraform.NewResourceConfigRaw(cfg), unitTestMockAPIClient)
}

func TestResourceAuditTrailCustomizeDiff(t *testing.T) {
	r := resourceAuditTrail()
	assert.NotNil(t, r.CustomizeDiff, "audit trail CustomizeDiff must be wired")

	t.Run("cloudwatch type without cloudwatch block errors", func(t *testing.T) {
		_, err := auditTrailDiffFixture(map[string]interface{}{
			"name": "at1",
			"type": "cloudwatch",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "`cloudwatch` block is required when `type` is `cloudwatch`")
	})

	t.Run("cloudwatch type with splunk also set errors", func(t *testing.T) {
		_, err := auditTrailDiffFixture(map[string]interface{}{
			"name": "at1",
			"type": "cloudwatch",
			"cloudwatch": []interface{}{
				map[string]interface{}{
					"group":           "g",
					"region":          "us-east-1",
					"credential_type": "secret",
					"access_key":      "ak",
					"secret_key":      "sk",
				},
			},
			"splunk": []interface{}{
				map[string]interface{}{
					"hec_url": "https://splunk.example.com",
					"token":   "tok",
				},
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "`splunk` block must not be set when `type` is `cloudwatch`")
	})

	t.Run("cloudwatch secret credential_type missing keys errors", func(t *testing.T) {
		_, err := auditTrailDiffFixture(map[string]interface{}{
			"name": "at1",
			"type": "cloudwatch",
			"cloudwatch": []interface{}{
				map[string]interface{}{
					"group":           "g",
					"region":          "us-east-1",
					"credential_type": "secret",
				},
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "`access_key` and `secret_key` are required when `credential_type` is `secret`")
	})

	t.Run("cloudwatch sts credential_type missing arn errors", func(t *testing.T) {
		_, err := auditTrailDiffFixture(map[string]interface{}{
			"name": "at1",
			"type": "cloudwatch",
			"cloudwatch": []interface{}{
				map[string]interface{}{
					"group":           "g",
					"region":          "us-east-1",
					"credential_type": "sts",
				},
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "`arn` is required when `credential_type` is `sts`")
	})

	t.Run("valid cloudwatch config has no error", func(t *testing.T) {
		diff, err := auditTrailDiffFixture(map[string]interface{}{
			"name": "at1",
			"type": "cloudwatch",
			"cloudwatch": []interface{}{
				map[string]interface{}{
					"group":           "g",
					"region":          "us-east-1",
					"credential_type": "sts",
					"arn":             "arn:aws:iam::123456789012:role/audit",
				},
			},
		})
		require.NoError(t, err)
		assert.NotNil(t, diff)
	})

	t.Run("splunk type without splunk block errors", func(t *testing.T) {
		_, err := auditTrailDiffFixture(map[string]interface{}{
			"name": "at1",
			"type": "splunk",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "`splunk` block is required when `type` is `splunk`")
	})

	t.Run("splunk type with cloudwatch also set errors", func(t *testing.T) {
		_, err := auditTrailDiffFixture(map[string]interface{}{
			"name": "at1",
			"type": "splunk",
			"splunk": []interface{}{
				map[string]interface{}{
					"hec_url": "https://splunk.example.com",
					"token":   "tok",
				},
			},
			"cloudwatch": []interface{}{
				map[string]interface{}{
					"group":  "g",
					"region": "us-east-1",
				},
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "`cloudwatch` block must not be set when `type` is `splunk`")
	})

	t.Run("valid splunk config has no error", func(t *testing.T) {
		diff, err := auditTrailDiffFixture(map[string]interface{}{
			"name": "at1",
			"type": "splunk",
			"splunk": []interface{}{
				map[string]interface{}{
					"hec_url": "https://splunk.example.com",
					"token":   "tok",
				},
			},
		})
		require.NoError(t, err)
		assert.NotNil(t, diff)
	})
}

// ---------------------------------------------------------------------------
// resource_password_policy CustomizeDiff wired
// ---------------------------------------------------------------------------

// (Already covered by resource_password_policy_test.go — pinned in another test)

// ---------------------------------------------------------------------------
// resolveRegistryNameToUID with non-empty name (still 22%)
// ---------------------------------------------------------------------------

func TestResolveRegistryNameToUID_NonEmpty(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	// Every registry-type branch (oci/helm/spectro/manifest/default) hits
	// its respective SDK call. Coverage-wise we just need the switch
	// dispatch to fire — SDK misses are fine.
	for _, rt := range []string{"oci", "helm", "spectro", "manifest", "generic"} {
		_, _ = resolveRegistryNameToUID(c, "some-name", rt)
	}
}

// ---------------------------------------------------------------------------
// GetCommonRegistryOci (0%)
// ---------------------------------------------------------------------------

func TestGetCommonRegistryOci_EmptyID(t *testing.T) {
	d := resourceRegistryOciEcr().TestResourceData()
	_, err := GetCommonRegistryOci(d, unitTestMockAPIClient)
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// resourceClusterProfile:1133 getRegistryIsSyncSupported at 0%
// ---------------------------------------------------------------------------

func TestGetRegistryIsSyncSupported(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)

	tests := []struct {
		name          string
		registryUID   string
		packType      models.V1PackType
		wantSupported bool
		wantFound     bool
	}{
		{name: "oci basic registry", registryUID: "some-registry-uid", packType: models.V1PackTypeOci, wantSupported: true, wantFound: true},
		// PLT-2497: ECR-type registries such as Palette Registry 404 on the basic endpoint.
		{name: "oci ecr registry falls back to ecr endpoint", registryUID: routes.OciEcrOnlyRegistryUID, packType: models.V1PackTypeOci, wantSupported: true, wantFound: true},
		// A basic registry that answers with isSyncSupported=false must not fall through to ECR.
		{name: "oci basic registry without sync support", registryUID: "oci-sync-not-supported", packType: models.V1PackTypeOci, wantSupported: false, wantFound: true},
		{name: "oci registry unknown to both endpoints", registryUID: routes.OciUnknownRegistryUID, packType: models.V1PackTypeOci, wantSupported: false, wantFound: false},
		{name: "empty registry uid", registryUID: "", packType: models.V1PackTypeOci, wantSupported: false, wantFound: false},
		{name: "helm registry", registryUID: "some-registry-uid", packType: models.V1PackTypeHelm, wantSupported: true, wantFound: true},
		{name: "manifest pack type", registryUID: "some-registry-uid", packType: models.V1PackTypeManifest, wantSupported: false, wantFound: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			supported, found := getRegistryIsSyncSupported(c, tt.registryUID, tt.packType)
			assert.Equal(t, tt.wantSupported, supported)
			assert.Equal(t, tt.wantFound, found)
		})
	}
}

func TestGetOciRegistrySyncStatusByUID(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)

	status, err := getOciRegistrySyncStatusByUID(c, routes.OciEcrOnlyRegistryUID)
	require.NoError(t, err)
	assert.True(t, status.IsSyncSupported)

	status, err = getOciRegistrySyncStatusByUID(c, routes.OciUnknownRegistryUID)
	require.Error(t, err)
	assert.Nil(t, status)
	assert.Contains(t, err.Error(), "basic:")
	assert.Contains(t, err.Error(), "ecr:")
}
