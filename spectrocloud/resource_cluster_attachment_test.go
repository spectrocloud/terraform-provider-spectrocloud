package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// toAddonDeployment / updateAddonDeployment (0%)
//
// clusterUid "cluster-uid-addon-ready" is a dispatch key in
// tests/mockApiServer/routes/mockCluster.go: GET on that UID returns a
// cluster whose node Conditions are all Ready and whose Status.Packs
// contains a ProfileUID of "test-addon-profile-uid" — matching that UID in
// cluster_profile.id lets waitForAddonDeploymentUpdate's StateChangeConf
// resolve to "Ready" on the very first poll (no real wait).
// ---------------------------------------------------------------------------

func TestToAddonDeployment(t *testing.T) {
	d := resourceAddonDeployment().TestResourceData()
	_ = d.Set("context", "project")
	_ = d.Set("cluster_uid", "cluster-uid-addon-ready")
	_ = d.Set("apply_setting", "DownloadAndInstall")
	_ = d.Set("cluster_profile", []interface{}{
		map[string]interface{}{
			"id": "test-addon-profile-uid",
		},
	})
	c := castV1Client(t, unitTestMockAPIClient)

	addonDeployment, err := toAddonDeployment(c, d)
	assert.NoError(t, err)
	assert.NotNil(t, addonDeployment)
	assert.Len(t, addonDeployment.Profiles, 1)
	assert.Equal(t, "test-addon-profile-uid", addonDeployment.Profiles[0].UID)
	assert.NotNil(t, addonDeployment.SpcApplySettings)
	assert.Equal(t, "DownloadAndInstall", addonDeployment.SpcApplySettings.ActionType)
}

func TestToAddonDeployment_NoProfile(t *testing.T) {
	// cluster_profile left empty → toAddonDeplProfiles returns zero
	// profiles, not an error (validateSingleClusterProfile is the caller's
	// job, not toAddonDeployment's).
	d := resourceAddonDeployment().TestResourceData()
	_ = d.Set("context", "project")
	_ = d.Set("cluster_uid", "cluster-uid-addon-ready")
	c := castV1Client(t, unitTestMockAPIClient)

	addonDeployment, err := toAddonDeployment(c, d)
	assert.NoError(t, err)
	assert.NotNil(t, addonDeployment)
	assert.Len(t, addonDeployment.Profiles, 0)
}

func TestToAddonDeployment_BadContext(t *testing.T) {
	d := resourceAddonDeployment().TestResourceData()
	_ = d.Set("context", "not-a-real-context")
	_ = d.Set("cluster_uid", "cluster-uid-addon-ready")
	c := castV1Client(t, unitTestMockAPIClient)

	_, err := toAddonDeployment(c, d)
	assert.Error(t, err)
}

func TestUpdateAddonDeployment(t *testing.T) {
	d := resourceAddonDeployment().TestResourceData()
	_ = d.Set("context", "project")
	_ = d.Set("cluster_uid", "cluster-uid-addon-ready")
	_ = d.Set("apply_setting", "DownloadAndInstall")
	_ = d.Set("cluster_profile", []interface{}{
		map[string]interface{}{
			"id": "test-addon-profile-uid",
		},
	})
	c := castV1Client(t, unitTestMockAPIClient)

	cluster, err := c.GetCluster("cluster-uid-addon-ready")
	assert.NoError(t, err)
	assert.NotNil(t, cluster)

	var diags diag.Diagnostics
	result := updateAddonDeployment(context.Background(), d, unitTestMockAPIClient, c, cluster, "cluster-uid-addon-ready", diags)
	assert.False(t, result.HasError(), "diags: %+v", result)
	// The main UpdateAddonDeployment/wait/SetId path all succeeded (no
	// error). The trailing resourceAddonDeploymentRead call then clears the
	// ID again because the mock cluster fixture's ClusterProfileTemplates
	// don't happen to have a name/version match against
	// GetClusterProfile's static fixture — that's a real (and harmless)
	// branch in readAddonDeployment, not a test bug.
}

func TestUpdateAddonDeployment_NoProfiles(t *testing.T) {
	// toAddonDeployment succeeds with zero profiles → the explicit
	// "zero profiles found" error branch is exercised.
	d := resourceAddonDeployment().TestResourceData()
	_ = d.Set("context", "project")
	_ = d.Set("cluster_uid", "cluster-uid-addon-ready")
	c := castV1Client(t, unitTestMockAPIClient)

	cluster, err := c.GetCluster("cluster-uid-addon-ready")
	assert.NoError(t, err)

	var diags diag.Diagnostics
	result := updateAddonDeployment(context.Background(), d, unitTestMockAPIClient, c, cluster, "cluster-uid-addon-ready", diags)
	assert.True(t, result.HasError())
}

func TestUpdateAddonDeployment_GetClusterProfileError(t *testing.T) {
	d := resourceAddonDeployment().TestResourceData()
	_ = d.Set("context", "project")
	_ = d.Set("cluster_uid", "cluster-uid-addon-ready")
	_ = d.Set("cluster_profile", []interface{}{
		map[string]interface{}{
			"id": "test-addon-profile-uid",
		},
	})
	cNeg := castV1Client(t, unitTestMockAPINegativeClient)

	cluster, _ := cNeg.GetCluster("cluster-uid-addon-ready")

	var diags diag.Diagnostics
	result := updateAddonDeployment(context.Background(), d, unitTestMockAPINegativeClient, cNeg, cluster, "cluster-uid-addon-ready", diags)
	assert.True(t, result.HasError())
}

func TestGetAddonDeploymentIdANDReverse(t *testing.T) {
	clusterId := "5eea74ed19"
	clusterProfileId := "0d445deb3ca"
	addonDeploymentId := clusterId + "_" + clusterProfileId

	testAddonDeploymentId := getAddonDeploymentId(clusterId, &models.V1ClusterProfile{Metadata: &models.V1ObjectMeta{UID: clusterProfileId}})
	if testAddonDeploymentId != addonDeploymentId {
		t.Errorf("got %s, wanted %s", testAddonDeploymentId, addonDeploymentId)
	}

	testClusterId := getClusterUID(testAddonDeploymentId)
	if testClusterId != clusterId {
		t.Errorf("got %s, wanted %s", testClusterId, clusterId)
	}

	testClusterProfileId, _ := getClusterProfileUID(testAddonDeploymentId)
	if testClusterProfileId != clusterProfileId {
		t.Errorf("got %s, wanted %s", testClusterProfileId, clusterProfileId)
	}
}

func TestIsProfileAttached(t *testing.T) {
	tests := []struct {
		name     string
		cluster  *models.V1SpectroCluster
		uid      string
		expected bool
	}{
		{
			name: "Profile Attached",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{UID: "profile-123"},
						{UID: "profile-456"},
					},
				},
			},
			uid:      "profile-123",
			expected: true,
		},
		{
			name: "Profile Not Attached",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{UID: "profile-123"},
						{UID: "profile-456"},
					},
				},
			},
			uid:      "profile-789",
			expected: false,
		},
		{
			name: "Empty Profile List",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{},
				},
			},
			uid:      "profile-123",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := isProfileAttached(tt.cluster, tt.uid)
			assert.Equal(t, tt.expected, output)
		})
	}
}
