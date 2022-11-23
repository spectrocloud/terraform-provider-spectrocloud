package spectrocloud

import (
	"github.com/spectrocloud/hapi/models"
	"testing"
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

	testClusterProfileId := getClusterProfileUID(testAddonDeploymentId)
	if testClusterProfileId != clusterProfileId {
		t.Errorf("got %s, wanted %s", testClusterProfileId, clusterProfileId)
	}
}
