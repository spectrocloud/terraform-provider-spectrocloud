package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/assert"
)

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
		name        string
		cluster     *models.V1SpectroCluster
		uid         string
		expected    bool
		description string
	}{
		{
			name: "Profile Attached - First in list",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{UID: "profile-123"},
						{UID: "profile-456"},
					},
				},
			},
			uid:         "profile-123",
			expected:    true,
			description: "Should return true when profile is first in the list",
		},
		{
			name: "Profile Attached - Last in list",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{UID: "profile-123"},
						{UID: "profile-456"},
						{UID: "profile-789"},
					},
				},
			},
			uid:         "profile-789",
			expected:    true,
			description: "Should return true when profile is last in the list",
		},
		{
			name: "Profile Attached - Middle of list",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{UID: "profile-123"},
						{UID: "profile-456"},
						{UID: "profile-789"},
					},
				},
			},
			uid:         "profile-456",
			expected:    true,
			description: "Should return true when profile is in the middle of the list",
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
			uid:         "profile-789",
			expected:    false,
			description: "Should return false when profile is not in the list",
		},
		{
			name: "Empty Profile List",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{},
				},
			},
			uid:         "profile-123",
			expected:    false,
			description: "Should return false when profile list is empty",
		},
		{
			name: "Single Profile Attached",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{UID: "profile-123"},
					},
				},
			},
			uid:         "profile-123",
			expected:    true,
			description: "Should return true when single profile matches",
		},
		{
			name: "Single Profile Not Attached",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{UID: "profile-123"},
					},
				},
			},
			uid:         "profile-456",
			expected:    false,
			description: "Should return false when single profile does not match",
		},
		{
			name: "Many Profiles - Profile Attached",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{UID: "profile-1"},
						{UID: "profile-2"},
						{UID: "profile-3"},
						{UID: "profile-4"},
						{UID: "profile-5"},
						{UID: "profile-6"},
						{UID: "profile-7"},
						{UID: "profile-8"},
						{UID: "profile-9"},
					},
				},
			},
			uid:         "profile-5",
			expected:    true,
			description: "Should return true when profile exists in large list",
		},
		{
			name: "Many Profiles - Profile Not Attached",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{UID: "profile-1"},
						{UID: "profile-2"},
						{UID: "profile-3"},
						{UID: "profile-4"},
						{UID: "profile-5"},
						{UID: "profile-6"},
						{UID: "profile-7"},
						{UID: "profile-8"},
						{UID: "profile-9"},
					},
				},
			},
			uid:         "profile-10",
			expected:    false,
			description: "Should return false when profile does not exist in large list",
		},
		{
			name: "Empty UID - Not Found",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{UID: "profile-123"},
						{UID: "profile-456"},
					},
				},
			},
			uid:         "",
			expected:    false,
			description: "Should return false when searching for empty UID",
		},
		{
			name: "Empty UID in List - Found",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{UID: ""},
						{UID: "profile-456"},
					},
				},
			},
			uid:         "",
			expected:    true,
			description: "Should return true when empty UID exists in list",
		},
		{
			name: "Case Sensitive Match",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{UID: "Profile-123"},
						{UID: "profile-456"},
					},
				},
			},
			uid:         "profile-123",
			expected:    false,
			description: "Should be case sensitive - Profile-123 != profile-123",
		},
		{
			name: "Case Sensitive Match - Exact",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{UID: "Profile-123"},
						{UID: "profile-456"},
					},
				},
			},
			uid:         "Profile-123",
			expected:    true,
			description: "Should return true for exact case match",
		},
		{
			name: "UID with Special Characters",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{UID: "profile-123-abc"},
						{UID: "profile-456-xyz"},
					},
				},
			},
			uid:         "profile-123-abc",
			expected:    true,
			description: "Should handle UIDs with special characters",
		},
		{
			name: "Long UID",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{UID: "very-long-profile-uid-12345678901234567890"},
						{UID: "profile-456"},
					},
				},
			},
			uid:         "very-long-profile-uid-12345678901234567890",
			expected:    true,
			description: "Should handle long UIDs",
		},
		{
			name: "Numeric UID",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{UID: "12345"},
						{UID: "67890"},
					},
				},
			},
			uid:         "12345",
			expected:    true,
			description: "Should handle numeric UIDs",
		},
		{
			name: "Real-world Hex Format UID",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{UID: "5eea74ed19"},
						{UID: "0d445deb3ca"},
					},
				},
			},
			uid:         "5eea74ed19",
			expected:    true,
			description: "Should handle real-world hex format UIDs",
		},
		{
			name: "Duplicate UIDs in List - First Match",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{UID: "profile-123"},
						{UID: "profile-123"},
						{UID: "profile-456"},
					},
				},
			},
			uid:         "profile-123",
			expected:    true,
			description: "Should return true when duplicate UIDs exist (matches first occurrence)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := isProfileAttached(tt.cluster, tt.uid)
			assert.Equal(t, tt.expected, output, tt.description)
		})
	}
}

func TestBuildAddonDeploymentId(t *testing.T) {
	tests := []struct {
		name        string
		clusterUid  string
		profileUIDs []string
		expected    string
		description string
	}{
		{
			name:        "Single profile UID",
			clusterUid:  "cluster-123",
			profileUIDs: []string{"profile-456"},
			expected:    "cluster-123_profile-456",
			description: "Should create ID with single profile UID",
		},
		{
			name:        "Multiple profile UIDs - already sorted",
			clusterUid:  "cluster-123",
			profileUIDs: []string{"profile-111", "profile-222", "profile-333"},
			expected:    "cluster-123_profile-111_profile-222_profile-333",
			description: "Should create ID with multiple sorted profile UIDs",
		},
		{
			name:        "Multiple profile UIDs - unsorted input",
			clusterUid:  "cluster-123",
			profileUIDs: []string{"profile-333", "profile-111", "profile-222"},
			expected:    "cluster-123_profile-111_profile-222_profile-333",
			description: "Should sort profile UIDs before joining",
		},
		{
			name:        "Multiple profile UIDs - reverse order",
			clusterUid:  "cluster-123",
			profileUIDs: []string{"profile-999", "profile-888", "profile-777"},
			expected:    "cluster-123_profile-777_profile-888_profile-999",
			description: "Should sort profile UIDs in ascending order",
		},
		{
			name:        "Empty profile UIDs slice",
			clusterUid:  "cluster-123",
			profileUIDs: []string{},
			expected:    "cluster-123",
			description: "Should return only cluster UID when no profiles",
		},
		{
			name:        "Nil profile UIDs slice",
			clusterUid:  "cluster-123",
			profileUIDs: nil,
			expected:    "cluster-123",
			description: "Should handle nil slice gracefully",
		},
		{
			name:        "Single character UIDs",
			clusterUid:  "c1",
			profileUIDs: []string{"p3", "p1", "p2"},
			expected:    "c1_p1_p2_p3",
			description: "Should work with short UIDs",
		},
		{
			name:        "Long UIDs",
			clusterUid:  "very-long-cluster-uid-12345678901234567890",
			profileUIDs: []string{"very-long-profile-uid-98765432109876543210", "another-long-profile-uid-11111111111111111111"},
			expected:    "very-long-cluster-uid-12345678901234567890_another-long-profile-uid-11111111111111111111_very-long-profile-uid-98765432109876543210",
			description: "Should handle long UIDs correctly",
		},
		{
			name:        "Duplicate profile UIDs",
			clusterUid:  "cluster-123",
			profileUIDs: []string{"profile-111", "profile-111", "profile-222"},
			expected:    "cluster-123_profile-111_profile-111_profile-222",
			description: "Should handle duplicate UIDs and sort them",
		},
		{
			name:        "Special characters in UIDs",
			clusterUid:  "cluster-123-abc",
			profileUIDs: []string{"profile-xyz-999", "profile-abc-111"},
			expected:    "cluster-123-abc_profile-abc-111_profile-xyz-999",
			description: "Should handle UIDs with hyphens and special characters",
		},
		{
			name:        "Many profile UIDs",
			clusterUid:  "cluster-123",
			profileUIDs: []string{"p9", "p1", "p5", "p3", "p7", "p2", "p6", "p4", "p8"},
			expected:    "cluster-123_p1_p2_p3_p4_p5_p6_p7_p8_p9",
			description: "Should sort many profile UIDs correctly",
		},
		{
			name:        "Same UIDs different order produces same result",
			clusterUid:  "cluster-123",
			profileUIDs: []string{"profile-333", "profile-111", "profile-222"},
			expected:    "cluster-123_profile-111_profile-222_profile-333",
			description: "Should produce consistent ID regardless of input order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildAddonDeploymentId(tt.clusterUid, tt.profileUIDs)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestBuildAddonDeploymentIdConsistency(t *testing.T) {
	// Test that the same inputs in different orders produce the same result
	clusterUid := "cluster-123"
	profileUIDs1 := []string{"profile-333", "profile-111", "profile-222"}
	profileUIDs2 := []string{"profile-111", "profile-222", "profile-333"}
	profileUIDs3 := []string{"profile-222", "profile-333", "profile-111"}

	result1 := buildAddonDeploymentId(clusterUid, profileUIDs1)
	result2 := buildAddonDeploymentId(clusterUid, profileUIDs2)
	result3 := buildAddonDeploymentId(clusterUid, profileUIDs3)

	expected := "cluster-123_profile-111_profile-222_profile-333"

	assert.Equal(t, expected, result1, "Result should be consistent regardless of input order")
	assert.Equal(t, expected, result2, "Result should be consistent regardless of input order")
	assert.Equal(t, expected, result3, "Result should be consistent regardless of input order")
	assert.Equal(t, result1, result2, "All results should be identical")
	assert.Equal(t, result2, result3, "All results should be identical")
}

func TestGetAddonDeploymentId(t *testing.T) {
	tests := []struct {
		name           string
		clusterUid     string
		clusterProfile *models.V1ClusterProfile
		expected       string
		description    string
	}{
		{
			name:       "Basic cluster and profile UIDs",
			clusterUid: "cluster-123",
			clusterProfile: &models.V1ClusterProfile{
				Metadata: &models.V1ObjectMeta{
					UID: "profile-456",
				},
			},
			expected:    "cluster-123_profile-456",
			description: "Should create ID with cluster UID and profile UID",
		},
		{
			name:       "Short UIDs",
			clusterUid: "c1",
			clusterProfile: &models.V1ClusterProfile{
				Metadata: &models.V1ObjectMeta{
					UID: "p1",
				},
			},
			expected:    "c1_p1",
			description: "Should work with short UIDs",
		},
		{
			name:       "Long UIDs",
			clusterUid: "very-long-cluster-uid-12345678901234567890",
			clusterProfile: &models.V1ClusterProfile{
				Metadata: &models.V1ObjectMeta{
					UID: "very-long-profile-uid-98765432109876543210",
				},
			},
			expected:    "very-long-cluster-uid-12345678901234567890_very-long-profile-uid-98765432109876543210",
			description: "Should handle long UIDs correctly",
		},
		{
			name:       "UIDs with hyphens",
			clusterUid: "cluster-123-abc",
			clusterProfile: &models.V1ClusterProfile{
				Metadata: &models.V1ObjectMeta{
					UID: "profile-xyz-999",
				},
			},
			expected:    "cluster-123-abc_profile-xyz-999",
			description: "Should handle UIDs with hyphens and special characters",
		},
		{
			name:       "UIDs with underscores in cluster UID",
			clusterUid: "cluster_123",
			clusterProfile: &models.V1ClusterProfile{
				Metadata: &models.V1ObjectMeta{
					UID: "profile-456",
				},
			},
			expected:    "cluster_123_profile-456",
			description: "Should handle underscores in cluster UID",
		},
		{
			name:       "UIDs with underscores in profile UID",
			clusterUid: "cluster-123",
			clusterProfile: &models.V1ClusterProfile{
				Metadata: &models.V1ObjectMeta{
					UID: "profile_456",
				},
			},
			expected:    "cluster-123_profile_456",
			description: "Should handle underscores in profile UID",
		},
		{
			name:       "Empty cluster UID",
			clusterUid: "",
			clusterProfile: &models.V1ClusterProfile{
				Metadata: &models.V1ObjectMeta{
					UID: "profile-456",
				},
			},
			expected:    "_profile-456",
			description: "Should handle empty cluster UID",
		},
		{
			name:       "Empty profile UID",
			clusterUid: "cluster-123",
			clusterProfile: &models.V1ClusterProfile{
				Metadata: &models.V1ObjectMeta{
					UID: "",
				},
			},
			expected:    "cluster-123_",
			description: "Should handle empty profile UID",
		},
		{
			name:       "Numeric UIDs",
			clusterUid: "12345",
			clusterProfile: &models.V1ClusterProfile{
				Metadata: &models.V1ObjectMeta{
					UID: "67890",
				},
			},
			expected:    "12345_67890",
			description: "Should handle numeric UIDs",
		},
		{
			name:       "Alphanumeric UIDs",
			clusterUid: "cluster123abc",
			clusterProfile: &models.V1ClusterProfile{
				Metadata: &models.V1ObjectMeta{
					UID: "profile456xyz",
				},
			},
			expected:    "cluster123abc_profile456xyz",
			description: "Should handle alphanumeric UIDs",
		},
		{
			name:       "UIDs with mixed case",
			clusterUid: "Cluster-123",
			clusterProfile: &models.V1ClusterProfile{
				Metadata: &models.V1ObjectMeta{
					UID: "Profile-456",
				},
			},
			expected:    "Cluster-123_Profile-456",
			description: "Should preserve case in UIDs",
		},
		{
			name:       "Real-world format UIDs",
			clusterUid: "5eea74ed19",
			clusterProfile: &models.V1ClusterProfile{
				Metadata: &models.V1ObjectMeta{
					UID: "0d445deb3ca",
				},
			},
			expected:    "5eea74ed19_0d445deb3ca",
			description: "Should handle real-world hex format UIDs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getAddonDeploymentId(tt.clusterUid, tt.clusterProfile)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestParseAddonDeploymentId(t *testing.T) {
	tests := []struct {
		name             string
		id               string
		expectedUID      string
		expectedProfiles []string
		expectedError    bool
		errorMessage     string
		description      string
	}{
		{
			name:             "Single profile UID",
			id:               "cluster-123_profile-456",
			expectedUID:      "cluster-123",
			expectedProfiles: []string{"profile-456"},
			expectedError:    false,
			description:      "Should parse ID with single profile UID",
		},
		{
			name:             "Multiple profile UIDs",
			id:               "cluster-123_profile-111_profile-222_profile-333",
			expectedUID:      "cluster-123",
			expectedProfiles: []string{"profile-111", "profile-222", "profile-333"},
			expectedError:    false,
			description:      "Should parse ID with multiple profile UIDs",
		},
		{
			name:             "Two profile UIDs",
			id:               "cluster-123_profile-111_profile-222",
			expectedUID:      "cluster-123",
			expectedProfiles: []string{"profile-111", "profile-222"},
			expectedError:    false,
			description:      "Should parse ID with two profile UIDs",
		},
		{
			name:             "Short UIDs",
			id:               "c1_p1",
			expectedUID:      "c1",
			expectedProfiles: []string{"p1"},
			expectedError:    false,
			description:      "Should work with short UIDs",
		},
		{
			name:             "Long UIDs",
			id:               "very-long-cluster-uid-12345678901234567890_very-long-profile-uid-98765432109876543210",
			expectedUID:      "very-long-cluster-uid-12345678901234567890",
			expectedProfiles: []string{"very-long-profile-uid-98765432109876543210"},
			expectedError:    false,
			description:      "Should handle long UIDs correctly",
		},
		{
			name:             "UIDs with hyphens",
			id:               "cluster-123-abc_profile-xyz-999",
			expectedUID:      "cluster-123-abc",
			expectedProfiles: []string{"profile-xyz-999"},
			expectedError:    false,
			description:      "Should handle UIDs with hyphens",
		},
		{
			name:             "UIDs with underscores in cluster UID",
			id:               "cluster_123_profile-456",
			expectedUID:      "cluster",
			expectedProfiles: []string{"123", "profile-456"},
			expectedError:    false,
			description:      "Should split on all underscores - cluster UID with underscore gets split",
		},
		{
			name:             "UIDs with underscores in profile UIDs",
			id:               "cluster-123_profile_456_profile_789",
			expectedUID:      "cluster-123",
			expectedProfiles: []string{"profile", "456", "profile", "789"},
			expectedError:    false,
			description:      "Should split on all underscores - profile UIDs with underscores get split",
		},
		{
			name:             "Numeric UIDs",
			id:               "12345_67890",
			expectedUID:      "12345",
			expectedProfiles: []string{"67890"},
			expectedError:    false,
			description:      "Should handle numeric UIDs",
		},
		{
			name:             "Alphanumeric UIDs",
			id:               "cluster123abc_profile456xyz",
			expectedUID:      "cluster123abc",
			expectedProfiles: []string{"profile456xyz"},
			expectedError:    false,
			description:      "Should handle alphanumeric UIDs",
		},
		{
			name:             "Real-world format UIDs",
			id:               "5eea74ed19_0d445deb3ca",
			expectedUID:      "5eea74ed19",
			expectedProfiles: []string{"0d445deb3ca"},
			expectedError:    false,
			description:      "Should handle real-world hex format UIDs",
		},
		{
			name:             "Only underscore",
			id:               "_",
			expectedUID:      "",
			expectedProfiles: []string{""},
			expectedError:    false,
			description:      "Should handle single underscore (edge case)",
		},
		{
			name:             "Cluster UID with trailing underscore",
			id:               "cluster-123_",
			expectedUID:      "cluster-123",
			expectedProfiles: []string{""},
			expectedError:    false,
			description:      "Should handle trailing underscore with empty profile UID",
		},
		{
			name:             "Empty cluster UID with profile",
			id:               "_profile-456",
			expectedUID:      "",
			expectedProfiles: []string{"profile-456"},
			expectedError:    false,
			description:      "Should handle empty cluster UID",
		},
		{
			name:             "Multiple consecutive underscores",
			id:               "cluster-123__profile-456",
			expectedUID:      "cluster-123",
			expectedProfiles: []string{"", "profile-456"},
			expectedError:    false,
			description:      "Should handle multiple consecutive underscores",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterUID, profileUIDs, err := parseAddonDeploymentId(tt.id)

			if tt.expectedError {
				assert.Error(t, err, tt.description)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage, "Error message should match")
				}
				assert.Empty(t, clusterUID, "Cluster UID should be empty on error")
				assert.Nil(t, profileUIDs, "Profile UIDs should be nil on error")
			} else {
				assert.NoError(t, err, tt.description)
				assert.Equal(t, tt.expectedUID, clusterUID, "Cluster UID should match")
				assert.Equal(t, tt.expectedProfiles, profileUIDs, "Profile UIDs should match")
			}
		})
	}
}

func TestToAddonDeployment(t *testing.T) {
	tests := []struct {
		name             string
		setupData        func() *schema.ResourceData
		setupClient      func() *client.V1Client
		expectError      bool
		expectedProfiles int
		expectedSettings *models.V1SpcApplySettings
		description      string
	}{
		{
			name: "Success with profiles and apply_setting",
			setupData: func() *schema.ResourceData {
				d := resourceAddonDeployment().TestResourceData()
				d.SetId("cluster-123_profile-456")
				d.Set("cluster_uid", "cluster-123")
				d.Set("context", "project")
				d.Set("apply_setting", "DownloadAndInstall")
				d.Set("cluster_profile", []interface{}{
					map[string]interface{}{
						"id": "profile-456",
					},
				})
				return d
			},
			setupClient: func() *client.V1Client {
				if unitTestMockAPIClient != nil {
					if c, ok := unitTestMockAPIClient.(*client.V1Client); ok {
						return c
					}
				}
				return nil
			},
			expectError:      false,
			expectedProfiles: 1,
			expectedSettings: &models.V1SpcApplySettings{
				ActionType: "DownloadAndInstall",
			},
			description: "Should successfully create addon deployment with profiles and settings",
		},
		{
			name: "Success with DownloadAndInstallLater setting",
			setupData: func() *schema.ResourceData {
				d := resourceAddonDeployment().TestResourceData()
				d.SetId("cluster-123_profile-456")
				d.Set("cluster_uid", "cluster-123")
				d.Set("context", "project")
				d.Set("apply_setting", "DownloadAndInstallLater")
				d.Set("cluster_profile", []interface{}{
					map[string]interface{}{
						"id": "profile-456",
					},
				})
				return d
			},
			setupClient: func() *client.V1Client {
				if unitTestMockAPIClient != nil {
					if c, ok := unitTestMockAPIClient.(*client.V1Client); ok {
						return c
					}
				}
				return nil
			},
			expectError:      false,
			expectedProfiles: 1,
			expectedSettings: &models.V1SpcApplySettings{
				ActionType: "DownloadAndInstallLater",
			},
			description: "Should handle DownloadAndInstallLater apply setting",
		},
		{
			name: "Success with empty apply_setting",
			setupData: func() *schema.ResourceData {
				d := resourceAddonDeployment().TestResourceData()
				d.SetId("cluster-123_profile-456")
				d.Set("cluster_uid", "cluster-123")
				d.Set("context", "project")
				d.Set("apply_setting", "")
				d.Set("cluster_profile", []interface{}{
					map[string]interface{}{
						"id": "profile-456",
					},
				})
				return d
			},
			setupClient: func() *client.V1Client {
				if unitTestMockAPIClient != nil {
					if c, ok := unitTestMockAPIClient.(*client.V1Client); ok {
						return c
					}
				}
				return nil
			},
			expectError:      false,
			expectedProfiles: 1,
			expectedSettings: nil,
			description:      "Should handle empty apply_setting (returns nil)",
		},
		{
			name: "Success with no apply_setting field",
			setupData: func() *schema.ResourceData {
				d := resourceAddonDeployment().TestResourceData()
				d.SetId("cluster-123_profile-456")
				d.Set("cluster_uid", "cluster-123")
				d.Set("context", "project")
				// Don't set apply_setting
				d.Set("cluster_profile", []interface{}{
					map[string]interface{}{
						"id": "profile-456",
					},
				})
				return d
			},
			setupClient: func() *client.V1Client {
				if unitTestMockAPIClient != nil {
					if c, ok := unitTestMockAPIClient.(*client.V1Client); ok {
						return c
					}
				}
				return nil
			},
			expectError:      false,
			expectedProfiles: 1,
			expectedSettings: nil,
			description:      "Should handle missing apply_setting field (returns nil)",
		},
		{
			name: "Success with multiple profiles",
			setupData: func() *schema.ResourceData {
				d := resourceAddonDeployment().TestResourceData()
				d.SetId("cluster-123_profile-456_profile-789")
				d.Set("cluster_uid", "cluster-123")
				d.Set("context", "project")
				d.Set("apply_setting", "DownloadAndInstall")
				d.Set("cluster_profile", []interface{}{
					map[string]interface{}{
						"id": "profile-456",
					},
					map[string]interface{}{
						"id": "profile-789",
					},
				})
				return d
			},
			setupClient: func() *client.V1Client {
				if unitTestMockAPIClient != nil {
					if c, ok := unitTestMockAPIClient.(*client.V1Client); ok {
						return c
					}
				}
				return nil
			},
			expectError:      false,
			expectedProfiles: 2,
			expectedSettings: &models.V1SpcApplySettings{
				ActionType: "DownloadAndInstall",
			},
			description: "Should handle multiple profiles",
		},
		{
			name: "Success with profiles and variables",
			setupData: func() *schema.ResourceData {
				d := resourceAddonDeployment().TestResourceData()
				d.SetId("cluster-123_profile-456")
				d.Set("cluster_uid", "cluster-123")
				d.Set("context", "project")
				d.Set("apply_setting", "DownloadAndInstall")
				d.Set("cluster_profile", []interface{}{
					map[string]interface{}{
						"id": "profile-456",
						"variables": map[string]interface{}{
							"var1": "value1",
							"var2": "value2",
						},
					},
				})
				return d
			},
			setupClient: func() *client.V1Client {
				if unitTestMockAPIClient != nil {
					if c, ok := unitTestMockAPIClient.(*client.V1Client); ok {
						return c
					}
				}
				return nil
			},
			expectError:      false,
			expectedProfiles: 1,
			expectedSettings: &models.V1SpcApplySettings{
				ActionType: "DownloadAndInstall",
			},
			description: "Should handle profiles with variables",
		},
		{
			name: "Error with invalid context",
			setupData: func() *schema.ResourceData {
				d := resourceAddonDeployment().TestResourceData()
				d.SetId("cluster-123_profile-456")
				d.Set("cluster_uid", "cluster-123")
				d.Set("context", "invalid-context")
				d.Set("apply_setting", "DownloadAndInstall")
				d.Set("cluster_profile", []interface{}{
					map[string]interface{}{
						"id": "profile-456",
					},
				})
				return d
			},
			setupClient: func() *client.V1Client {
				if unitTestMockAPIClient != nil {
					if c, ok := unitTestMockAPIClient.(*client.V1Client); ok {
						return c
					}
				}
				return nil
			},
			expectError:      true,
			expectedProfiles: 0,
			expectedSettings: nil,
			description:      "Should return error for invalid context",
		},
		{
			name: "Error with nil client when cluster retrieval needed",
			setupData: func() *schema.ResourceData {
				d := resourceAddonDeployment().TestResourceData()
				d.SetId("cluster-123_profile-456")
				d.Set("cluster_uid", "cluster-123")
				d.Set("context", "project")
				d.Set("apply_setting", "DownloadAndInstall")
				d.Set("cluster_profile", []interface{}{
					map[string]interface{}{
						"id": "profile-456",
					},
				})
				return d
			},
			setupClient: func() *client.V1Client {
				return nil
			},
			expectError:      true,
			expectedProfiles: 0,
			expectedSettings: nil,
			description:      "Should return error when client is nil and cluster retrieval is needed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.setupData()
			c := tt.setupClient()

			result, err := toAddonDeployment(c, d)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, result, "Result should be nil on error")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, result, "Result should not be nil on success")
				if result != nil {
					assert.Equal(t, tt.expectedProfiles, len(result.Profiles), "Profile count should match")
					if tt.expectedSettings != nil {
						assert.NotNil(t, result.SpcApplySettings, "Settings should not be nil")
						assert.Equal(t, tt.expectedSettings.ActionType, result.SpcApplySettings.ActionType, "ActionType should match")
					} else {
						assert.Nil(t, result.SpcApplySettings, "Settings should be nil when not set")
					}
				}
			}
		})
	}

	// Test to verify the function combines profiles and settings correctly
	t.Run("Verify profiles and settings are combined correctly", func(t *testing.T) {
		d := resourceAddonDeployment().TestResourceData()
		d.SetId("cluster-123_profile-456")
		d.Set("cluster_uid", "cluster-123")
		d.Set("context", "project")
		d.Set("apply_setting", "DownloadAndInstall")
		d.Set("cluster_profile", []interface{}{
			map[string]interface{}{
				"id": "profile-456",
			},
		})

		c := func() *client.V1Client {
			if unitTestMockAPIClient != nil {
				if client, ok := unitTestMockAPIClient.(*client.V1Client); ok {
					return client
				}
			}
			return nil
		}()

		if c != nil {
			result, err := toAddonDeployment(c, d)
			assert.NoError(t, err, "Should not error with valid input")
			if err == nil && result != nil {
				// Verify both profiles and settings are present
				assert.NotNil(t, result.Profiles, "Profiles should not be nil")
				assert.NotNil(t, result.SpcApplySettings, "Settings should not be nil")
				assert.Equal(t, "DownloadAndInstall", result.SpcApplySettings.ActionType, "ActionType should be set correctly")
				assert.Equal(t, 1, len(result.Profiles), "Should have one profile")
				if len(result.Profiles) > 0 {
					assert.Equal(t, "profile-456", result.Profiles[0].UID, "Profile UID should match")
				}
			}
		}
	})
}

func TestGetClusterProfileUIDs(t *testing.T) {
	tests := []struct {
		name              string
		addonDeploymentId string
		expectedProfiles  []string
		expectedError     bool
		errorMessage      string
		description       string
	}{
		{
			name:              "Single profile UID",
			addonDeploymentId: "cluster-123_profile-456",
			expectedProfiles:  []string{"profile-456"},
			expectedError:     false,
			description:       "Should extract single profile UID",
		},
		{
			name:              "Multiple profile UIDs",
			addonDeploymentId: "cluster-123_profile-111_profile-222_profile-333",
			expectedProfiles:  []string{"profile-111", "profile-222", "profile-333"},
			expectedError:     false,
			description:       "Should extract multiple profile UIDs",
		},
		{
			name:              "Two profile UIDs",
			addonDeploymentId: "cluster-123_profile-111_profile-222",
			expectedProfiles:  []string{"profile-111", "profile-222"},
			expectedError:     false,
			description:       "Should extract two profile UIDs",
		},
		{
			name:              "Short UIDs",
			addonDeploymentId: "c1_p1",
			expectedProfiles:  []string{"p1"},
			expectedError:     false,
			description:       "Should work with short UIDs",
		},
		{
			name:              "Long UIDs",
			addonDeploymentId: "very-long-cluster-uid-12345678901234567890_very-long-profile-uid-98765432109876543210",
			expectedProfiles:  []string{"very-long-profile-uid-98765432109876543210"},
			expectedError:     false,
			description:       "Should handle long UIDs correctly",
		},
		{
			name:              "UIDs with hyphens",
			addonDeploymentId: "cluster-123-abc_profile-xyz-999",
			expectedProfiles:  []string{"profile-xyz-999"},
			expectedError:     false,
			description:       "Should handle UIDs with hyphens",
		},
		{
			name:              "UIDs with underscores in cluster UID",
			addonDeploymentId: "cluster_123_profile-456",
			expectedProfiles:  []string{"123", "profile-456"},
			expectedError:     false,
			description:       "Should split on all underscores - cluster UID with underscore gets split",
		},
		{
			name:              "UIDs with underscores in profile UIDs",
			addonDeploymentId: "cluster-123_profile_456_profile_789",
			expectedProfiles:  []string{"profile", "456", "profile", "789"},
			expectedError:     false,
			description:       "Should split on all underscores - profile UIDs with underscores get split",
		},
		{
			name:              "Numeric UIDs",
			addonDeploymentId: "12345_67890",
			expectedProfiles:  []string{"67890"},
			expectedError:     false,
			description:       "Should handle numeric UIDs",
		},
		{
			name:              "Alphanumeric UIDs",
			addonDeploymentId: "cluster123abc_profile456xyz",
			expectedProfiles:  []string{"profile456xyz"},
			expectedError:     false,
			description:       "Should handle alphanumeric UIDs",
		},
		{
			name:              "Real-world format UIDs",
			addonDeploymentId: "5eea74ed19_0d445deb3ca",
			expectedProfiles:  []string{"0d445deb3ca"},
			expectedError:     false,
			description:       "Should handle real-world hex format UIDs",
		},
		{
			name:              "Many profile UIDs",
			addonDeploymentId: "cluster-123_p1_p2_p3_p4_p5_p6_p7_p8_p9",
			expectedProfiles:  []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9"},
			expectedError:     false,
			description:       "Should handle many profile UIDs",
		},
		{
			name:              "Only underscore",
			addonDeploymentId: "_",
			expectedProfiles:  []string{""},
			expectedError:     false,
			description:       "Should handle single underscore (edge case)",
		},
		{
			name:              "Cluster UID with trailing underscore",
			addonDeploymentId: "cluster-123_",
			expectedProfiles:  []string{""},
			expectedError:     false,
			description:       "Should handle trailing underscore with empty profile UID",
		},
		{
			name:              "Empty cluster UID with profile",
			addonDeploymentId: "_profile-456",
			expectedProfiles:  []string{"profile-456"},
			expectedError:     false,
			description:       "Should handle empty cluster UID",
		},
		{
			name:              "Multiple consecutive underscores",
			addonDeploymentId: "cluster-123__profile-456",
			expectedProfiles:  []string{"", "profile-456"},
			expectedError:     false,
			description:       "Should handle multiple consecutive underscores",
		},
		{
			name:              "Empty string",
			addonDeploymentId: "",
			expectedProfiles:  nil,
			expectedError:     true,
			errorMessage:      "invalid addon deployment ID format: ",
			description:       "Should return error for empty string",
		},
		{
			name:              "Only cluster UID (no underscore)",
			addonDeploymentId: "cluster-123",
			expectedProfiles:  nil,
			expectedError:     true,
			errorMessage:      "invalid addon deployment ID format: cluster-123",
			description:       "Should return error when no profile UID present",
		},
		{
			name:              "Single character cluster UID",
			addonDeploymentId: "c_p1",
			expectedProfiles:  []string{"p1"},
			expectedError:     false,
			description:       "Should handle single character cluster UID",
		},
		{
			name:              "Single character profile UID",
			addonDeploymentId: "cluster-123_p",
			expectedProfiles:  []string{"p"},
			expectedError:     false,
			description:       "Should handle single character profile UID",
		},
		{
			name:              "Profile UIDs with special characters",
			addonDeploymentId: "cluster-123_profile-abc-123_profile-xyz-789",
			expectedProfiles:  []string{"profile-abc-123", "profile-xyz-789"},
			expectedError:     false,
			description:       "Should handle profile UIDs with special characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileUIDs, err := getClusterProfileUIDs(tt.addonDeploymentId)

			if tt.expectedError {
				assert.Error(t, err, tt.description)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage, "Error message should match")
				}
				assert.Nil(t, profileUIDs, "Profile UIDs should be nil on error")
			} else {
				assert.NoError(t, err, tt.description)
				assert.Equal(t, tt.expectedProfiles, profileUIDs, "Profile UIDs should match")
			}
		})
	}
}

// Test to verify getClusterProfileUIDs is consistent with parseAddonDeploymentId
func TestGetClusterProfileUIDsConsistency(t *testing.T) {
	testCases := []string{
		"cluster-123_profile-456",
		"cluster-123_profile-111_profile-222_profile-333",
		"c1_p1",
		"5eea74ed19_0d445deb3ca",
		"cluster-123_p1_p2_p3",
	}

	for _, testID := range testCases {
		t.Run(testID, func(t *testing.T) {
			// Get profile UIDs using getClusterProfileUIDs
			profileUIDs1, err1 := getClusterProfileUIDs(testID)

			// Get profile UIDs using parseAddonDeploymentId
			_, profileUIDs2, err2 := parseAddonDeploymentId(testID)

			// Both should have the same error state
			if err1 != nil {
				assert.Error(t, err2, "parseAddonDeploymentId should also error")
			} else {
				assert.NoError(t, err2, "parseAddonDeploymentId should not error")
				// Both should return the same profile UIDs
				assert.Equal(t, profileUIDs2, profileUIDs1, "Profile UIDs should match between both functions")
			}
		})
	}
}
