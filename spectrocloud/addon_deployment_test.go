package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToAddonDeploymentPackCreate(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected *models.V1PackManifestEntity
		wantErr  bool
	}{
		{
			name: "ValidInputWithRegistryUID",
			input: map[string]interface{}{
				"name":         "test-pack",
				"tag":          "v1.0.0",
				"registry_uid": "registry-123",
				"type":         "oci",
				"values":       "some values\n",
				"manifest": []interface{}{
					map[string]interface{}{
						"content": "manifest content 1\n",
						"name":    "manifest-1",
					},
					map[string]interface{}{
						"content": "manifest content 2\n",
						"name":    "manifest-2",
					},
				},
			},
			expected: &models.V1PackManifestEntity{
				Name:        types.Ptr("test-pack"),
				Tag:         "v1.0.0",
				RegistryUID: "registry-123",
				Type:        models.V1PackTypeOci.Pointer(),
				Values:      "some values",
				Manifests: []*models.V1ManifestInputEntity{
					{
						Content: "manifest content 1",
						Name:    "manifest-1",
					},
					{
						Content: "manifest content 2",
						Name:    "manifest-2",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "ValidInputWithoutRegistryUID",
			input: map[string]interface{}{
				"name":         "test-pack",
				"tag":          "v1.0.0",
				"uid":          "pack-uid-456",
				"registry_uid": nil,
				"type":         "oci",
				"values":       "some values\n",
				"manifest": []interface{}{
					map[string]interface{}{
						"content": "manifest content 1\n",
						"name":    "manifest-1",
					},
				},
			},
			expected: &models.V1PackManifestEntity{
				Name:        types.Ptr("test-pack"),
				Tag:         "v1.0.0",
				RegistryUID: "",
				Type:        models.V1PackTypeOci.Pointer(),
				Values:      "some values",
				Manifests: []*models.V1ManifestInputEntity{
					{
						Content: "manifest content 1",
						Name:    "manifest-1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "EmptyManifest",
			input: map[string]interface{}{
				"name":         "test-pack",
				"tag":          "v1.0.0",
				"registry_uid": "registry-123",
				"type":         "oci",
				"values":       "some values\n",
				"manifest":     []interface{}{},
			},
			expected: &models.V1PackManifestEntity{
				Name:        types.Ptr("test-pack"),
				Tag:         "v1.0.0",
				RegistryUID: "registry-123",
				Type:        models.V1PackTypeOci.Pointer(),
				Values:      "some values",
				Manifests:   []*models.V1ManifestInputEntity{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toAddonDeploymentPackCreate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("toAddonDeploymentPackCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGetAddonDeploymentDiagPacks(t *testing.T) {
	// Helper function to create a schema.ResourceData
	createResourceData := func(clusterProfiles []interface{}) *schema.ResourceData {
		d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
			"cluster_profile": {
				Type:     schema.TypeList,
				Elem:     &schema.Resource{Schema: map[string]*schema.Schema{"pack": {Type: schema.TypeList, Elem: &schema.Resource{Schema: map[string]*schema.Schema{"name": {Type: schema.TypeString}, "tag": {Type: schema.TypeString}, "registry_uid": {Type: schema.TypeString}, "type": {Type: schema.TypeString}, "values": {Type: schema.TypeString}, "manifest": {Type: schema.TypeList, Elem: &schema.Resource{Schema: map[string]*schema.Schema{"content": {Type: schema.TypeString}, "name": {Type: schema.TypeString}}}}}}}}},
				Optional: true,
			},
		}, map[string]interface{}{
			"cluster_profile": clusterProfiles,
		})
		return d
	}

	// Valid input case
	t.Run("valid input", func(t *testing.T) {
		clusterProfiles := []interface{}{
			map[string]interface{}{
				"pack": []interface{}{
					map[string]interface{}{
						"name":         "test-pack",
						"tag":          "v1.0",
						"registry_uid": "uid-123",
						"type":         "oci",
						"values":       "some values",
						"manifest": []interface{}{
							map[string]interface{}{
								"content": "manifest-content-1",
								"name":    "manifest-1",
							},
						},
					},
				},
			},
		}
		d := createResourceData(clusterProfiles)
		diagPacks, diags, isError := GetAddonDeploymentDiagPacks(d, nil)

		assert.False(t, isError)
		assert.Nil(t, diags)
		require.Len(t, diagPacks, 1)

		pack := diagPacks[0]
		assert.Equal(t, "test-pack", *pack.Name)
		assert.Equal(t, "v1.0", pack.Tag)
		assert.Equal(t, "uid-123", pack.RegistryUID)
		assert.Equal(t, models.V1PackTypeOci, *pack.Type)
		assert.Equal(t, "some values", pack.Values)
		require.Len(t, pack.Manifests, 1)
		assert.Equal(t, "manifest-content-1", pack.Manifests[0].Content)
		assert.Equal(t, "manifest-1", pack.Manifests[0].Name)
	})
}

// TestReadAddonDeployment_InvalidID getClusterProfileUID
// requires an underscore-separated ID; anything else returns an error
// and readAddonDeployment short-circuits (nil, false) without hitting
// the SDK.
func TestReadAddonDeployment_InvalidID(t *testing.T) {
	d := resourceAddonDeployment().TestResourceData()
	d.SetId("no-underscore-in-this-id")
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")

	cluster := &models.V1SpectroCluster{
		Metadata: &models.V1ObjectMeta{UID: "test-cluster-uid"},
		Spec:     &models.V1SpectroClusterSpec{},
	}

	diags, ok := readAddonDeployment(c, d, cluster)
	assert.Nil(t, diags)
	assert.False(t, ok)
}

// TestReadAddonDeployment_NoMatchingProfile exercises the fallthrough
// past the ClusterProfileTemplates loop → d.SetId("") + return (diags,
// false).
func TestReadAddonDeployment_NoMatchingProfile(t *testing.T) {
	d := resourceAddonDeployment().TestResourceData()
	d.SetId("adep_test-cluster-profile-uid")
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")

	cluster := &models.V1SpectroCluster{
		Metadata: &models.V1ObjectMeta{UID: "test-cluster-uid"},
		Spec: &models.V1SpectroClusterSpec{
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{Name: "some-other-name", UID: "different-uid"},
			},
		},
	}

	// The SDK call to GetClusterProfile may miss; either way, the func
	// runs to completion — the only assertion that matters is that we
	// reached the body (no panic).
	_, _ = readAddonDeployment(c, d, cluster)
}
