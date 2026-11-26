package spectrocloud

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
)

// PLT-2514: a Day 2 cluster_profile -> cluster_template attach that Palette will reject must fail
// at plan time, and a rejected attach at apply must not be written to state.

const unknownValue = "74D93920-ED26-11E3-AC10-0800200C9A66" // hcl2shim.UnknownVariableValue

// planEksAttach plans replacing cluster_profile "p1" with cluster_template templateID on an
// existing EKS cluster, running CustomizeDiff with the given provider meta.
func planEksAttach(t *testing.T, templateID string, meta interface{}) error {
	t.Helper()
	res := resourceClusterEks()

	oldRaw := baseEksUpdateRaw(eksCloudConfigRaw(nil), []interface{}{}, []interface{}{})
	oldRaw["cluster_profile"] = []interface{}{map[string]interface{}{"id": "p1"}}
	oldRD := schema.TestResourceDataRaw(t, res.Schema, oldRaw)
	oldRD.SetId(eksClusterID)
	require.NoError(t, oldRD.Set("cloud_config_id", eksCloudConfigUID))

	newRaw := baseEksUpdateRaw(eksCloudConfigRaw(nil), []interface{}{}, []interface{}{})
	newRaw["cluster_template"] = []interface{}{map[string]interface{}{"id": templateID}}

	_, err := res.Diff(context.Background(), oldRD.State(), terraform.NewResourceConfigRaw(newRaw), meta)
	return err
}

func TestClusterTemplateAttachEligibility_Plan(t *testing.T) {
	t.Run("ineligible template fails the plan with the reasons", func(t *testing.T) {
		err := planEksAttach(t, routes.ClusterTemplateIneligibleUID, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "k8s-two-minors-ahead")
		assert.Contains(t, err.Error(), "K8sUpgradeNotSupported")
	})

	t.Run("eligible template passes", func(t *testing.T) {
		assert.NoError(t, planEksAttach(t, "test-cluster-config-template-id", unitTestMockAPIClient))
	})

	t.Run("template not in the listing is skipped", func(t *testing.T) {
		assert.NoError(t, planEksAttach(t, "template-not-listed", unitTestMockAPIClient))
	})

	t.Run("API error is skipped", func(t *testing.T) {
		// The negative mock server has no cluster template routes, so the listing call fails.
		assert.NoError(t, planEksAttach(t, routes.ClusterTemplateIneligibleUID, unitTestMockAPINegativeClient))
	})

	t.Run("template id not known yet is skipped", func(t *testing.T) {
		assert.NoError(t, planEksAttach(t, unknownValue, unitTestMockAPIClient))
	})

	t.Run("no provider client is skipped", func(t *testing.T) {
		assert.NoError(t, planEksAttach(t, routes.ClusterTemplateIneligibleUID, nil))
	})
}

func TestCheckClusterTemplateAttachEligibility(t *testing.T) {
	c := mustUnitClient(t, false)
	assert.NoError(t, checkClusterTemplateAttachEligibility(c, eksClusterID, "test-cluster-config-template-id"))
	assert.ErrorContains(t, checkClusterTemplateAttachEligibility(c, eksClusterID, routes.ClusterTemplateIneligibleUID),
		"upgrade from 1.34 to 1.36 is not supported")
}

func TestUpdateCommonFields_RejectedAttachKeepsPriorState(t *testing.T) {
	oldRaw := baseEksUpdateRaw(eksCloudConfigRaw(nil), []interface{}{}, []interface{}{})
	oldRaw["cluster_profile"] = []interface{}{map[string]interface{}{"id": "p1"}}

	newRaw := baseEksUpdateRaw(eksCloudConfigRaw(nil), []interface{}{}, []interface{}{})
	newRaw["cluster_template"] = []interface{}{map[string]interface{}{"id": routes.ClusterTemplateAttachRejectUID}}

	d := buildEksUpdateResourceData(t, oldRaw, newRaw, eksCloudConfigUID)
	attach, _ := classifyClusterTemplateTransition(d)
	require.True(t, attach)

	diags, errorSet := updateCommonFields(d, mustUnitClient(t, false))
	require.True(t, errorSet, "the rejected attach must surface as an error")
	require.NotEmpty(t, diags)

	// What Terraform saves after the failed apply: the cluster still uses its profile and the
	// rejected attach is not recorded, so the next plan proposes it again.
	state := d.State()
	require.NotNil(t, state)
	assert.Equal(t, "1", state.Attributes["cluster_profile.#"])
	profileIDs := []string{}
	for k, v := range state.Attributes {
		if strings.HasPrefix(k, "cluster_profile.") && strings.HasSuffix(k, ".id") {
			profileIDs = append(profileIDs, v)
		}
	}
	assert.Equal(t, []string{"p1"}, profileIDs)
	assert.Contains(t, []string{"", "0"}, state.Attributes["cluster_template.#"], "the rejected attach must not be saved")
}
