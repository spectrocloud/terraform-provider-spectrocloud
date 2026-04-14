package spectrocloud

import (
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/require"
)

func TestParseTerraformAddonDeploymentAnnotation(t *testing.T) {
	t.Parallel()
	raw := `["uid-a","uid-b"]`
	uids := parseTerraformAddonDeploymentAnnotation(raw)
	require.Equal(t, []string{"uid-a", "uid-b"}, uids)

	set := terraformAddonManagedProfileUIDSet(&models.V1SpectroCluster{
		Metadata: &models.V1ObjectMeta{
			Annotations: map[string]string{
				clusterAnnotationTerraformAddonDeployments: raw,
			},
		},
	})
	require.Len(t, set, 2)
	_, ok := set["uid-a"]
	require.True(t, ok)
}

func TestSerializeTerraformAddonDeploymentUIDs(t *testing.T) {
	t.Parallel()
	s, err := serializeTerraformAddonDeploymentUIDs(map[string]struct{}{"z": {}, "a": {}})
	require.NoError(t, err)
	require.Equal(t, `["a","z"]`, s)
}

func TestTerraformAddonManagedProfileUIDSetNilCluster(t *testing.T) {
	t.Parallel()
	require.Empty(t, terraformAddonManagedProfileUIDSet(nil))
}
