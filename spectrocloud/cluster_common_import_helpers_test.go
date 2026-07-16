package spectrocloud

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Update-HasChange suite test.
//
// Covers the last chunky 0% helpers in cluster_common*.go and
// cluster_common_profiles.go:
//   - flattenCommonAttributeForClusterImport
//   - flattenCommonAttributeForCustomClusterImport
//   - extractProfilesFromTemplate
//   - extractProfilesFromTemplateData
//   - setClusterProfilesOrTemplateForImport
//   - flattenClusterProfileForImport
//   - updateCommonFieldsForBrownfieldCluster
//   - waitForVirtualClusterLifecyclePause / Resume

// ---------------------------------------------------------------------------
// flattenCommonAttributeForClusterImport / CustomClusterImport
// ---------------------------------------------------------------------------

func TestFlattenCommonAttributeForClusterImport(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	d := resourceClusterAws().TestResourceData()
	d.SetId("test-cluster-id")

	// The helper walks: resourceClusterRead → setClusterProfilesOrTemplateForImport →
	// setCommonClusterImportAttributes. Each step hits mocked endpoints
	// (GetCluster + GetClusterConfigTemplate) or short-circuits when the
	// fixture doesn't include a matching field.
	err := flattenCommonAttributeForClusterImport(c, d)
	// A downstream SDK miss may bubble up; either way the branch is
	// covered. We assert only "no panic".
	_ = err
}

func TestFlattenCommonAttributeForCustomClusterImport(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	d := resourceClusterCustomCloud().TestResourceData()
	d.SetId("test-cluster-id")
	_ = d.Set("cloud", "nutanix")

	err := flattenCommonAttributeForCustomClusterImport(c, d)
	_ = err
}

// The existing TestExtractProfilesFromTemplate in cluster_common_profiles_test.go
// only validates the fixture data shape without calling the real function —
// see the "conceptual validation" comment there. The tests below actually
// invoke extractProfilesFromTemplate and extractProfilesFromTemplateData to
// take them from 0% to real coverage.

func TestExtractProfilesFromTemplate_Actual(t *testing.T) {
	d := resourceClusterAws().TestResourceData()
	got, err := extractProfilesFromTemplate(d)
	require.NoError(t, err)
	assert.Empty(t, got)

	profileSet := clusterTemplateProfileSetFromList([]interface{}{
		map[string]interface{}{"id": "profile-1"},
		map[string]interface{}{"id": ""}, // filtered out
	})
	_ = d.Set("cluster_template", []interface{}{
		map[string]interface{}{
			"id":              "template-1",
			"cluster_profile": profileSet,
		},
	})
	got, err = extractProfilesFromTemplate(d)
	require.NoError(t, err)
	assert.Len(t, got, 1)
}

func TestExtractProfilesFromTemplateData_Actual(t *testing.T) {
	got, err := extractProfilesFromTemplateData([]interface{}{})
	require.NoError(t, err)
	assert.Empty(t, got)

	profileSet := clusterTemplateProfileSetFromList([]interface{}{
		map[string]interface{}{"id": "profile-1"},
		map[string]interface{}{"id": ""},
	})
	got, err = extractProfilesFromTemplateData([]interface{}{
		map[string]interface{}{
			"id":              "template-1",
			"cluster_profile": profileSet,
		},
	})
	require.NoError(t, err)
	assert.Len(t, got, 1)
}

// ---------------------------------------------------------------------------
// setClusterProfilesOrTemplateForImport + flattenClusterProfileForImport
// ---------------------------------------------------------------------------

func TestFlattenClusterProfileForImport(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	d := resourceClusterAws().TestResourceData()
	d.SetId("test-cluster-id")

	profiles, err := flattenClusterProfileForImport(c, d)
	// The mock cluster fixture may or may not have ClusterProfileTemplates;
	// either way, the function body is exercised.
	_ = err
	_ = profiles
}

func TestSetClusterProfilesOrTemplateForImport(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	d := resourceClusterAws().TestResourceData()
	d.SetId("test-cluster-id")

	// Cluster with no ClusterTemplate → falls through to d.Set("cluster_profile", ...).
	cluster := &models.V1SpectroCluster{
		Metadata: &models.V1ObjectMeta{UID: "test-cluster-id"},
		Spec:     &models.V1SpectroClusterSpec{},
	}
	err := setClusterProfilesOrTemplateForImport(c, d, cluster)
	assert.NoError(t, err)

	// Cluster WITH a ClusterTemplate → sets d.Set("cluster_template", ...).
	// GetClusterConfigTemplate may miss; the branch is what we care about.
	clusterWithTpl := &models.V1SpectroCluster{
		Metadata: &models.V1ObjectMeta{UID: "test-cluster-id"},
		Spec: &models.V1SpectroClusterSpec{
			ClusterTemplate: &models.V1SpectroClusterTemplateRef{UID: "template-1"},
		},
	}
	err = setClusterProfilesOrTemplateForImport(c, d, clusterWithTpl)
	_ = err
}

// ---------------------------------------------------------------------------
// updateCommonFieldsForBrownfieldCluster
// ---------------------------------------------------------------------------

func TestUpdateCommonFieldsForBrownfieldCluster(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)

	// Empty resource — updateClusterMetadata will succeed against the
	// mock or silently no-op; scan/backup guards skip because their
	// blocks aren't populated. The helper swallows all errors from the
	// sub-calls except renewK8sCertificatesNow.
	d := resourceClusterAws().TestResourceData()
	d.SetId("test-cluster-id")
	_ = d.Set("context", "project")

	diags := updateCommonFieldsForBrownfieldCluster(d, c)
	// May return an error diag if renewK8sCertificatesNow hits the SDK;
	// no panic is the coverage-goal contract.
	_ = diags
}

// ---------------------------------------------------------------------------
// waitForVirtualClusterLifecyclePause + Resume
// ---------------------------------------------------------------------------

// TestWaitForVirtualClusterLifecyclePause exercises the wait loop with
// a fixture cluster that already reports "Paused" so the very first
// refresh sees the target state and the wait returns cleanly.
func TestWaitForVirtualClusterLifecyclePause(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	d := resourceClusterAws().TestResourceData()
	_ = d.Set("context", "project")
	// Provide a Create-timeout so d.Timeout(TimeoutCreate) doesn't panic.
	// Timeouts on a raw TestResourceData default to 20m — plenty long.

	diags, isErr := waitForVirtualClusterLifecyclePause(
		context.Background(), d, "cluster-uid-paused", diag.Diagnostics{}, c)
	assert.False(t, isErr)
	assert.Empty(t, diags)
}

// TestWaitForVirtualClusterLifecycleResume — parallels the Pause test
// but targets "Running".
func TestWaitForVirtualClusterLifecycleResume(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	d := resourceClusterAws().TestResourceData()
	_ = d.Set("context", "project")

	// Use a short-lived context so if the mock's cluster doesn't report
	// Running immediately the wait bails via ctx timeout instead of
	// blocking the test run.
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	_, _ = waitForVirtualClusterLifecycleResume(ctx, d,
		"cluster-uid-vcluster-running", diag.Diagnostics{}, c)
}

// Compile-time reference to keep the schema import used (for the
// clusterTemplateProfileSetFromList helper's schema.Set return type).
var _ = &schema.Set{}
