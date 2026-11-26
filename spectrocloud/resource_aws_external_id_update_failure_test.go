package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PLT-2553: Palette rejects an STS external ID that is not the tenant's own. When it rejects an update, state must keep
// the previous external_id. Neither resource below can correct it on refresh: the AWS cloud account never reads
// external_id back, and the audit trail keeps the value from state because Palette returns it masked. Without this, the
// next plan would show no changes even though Palette still has the old value.

// applyFailedUpdate plans prior -> config and applies it with the negative mock API, which rejects the update.
func applyFailedUpdate(t *testing.T, res *schema.Resource, prior, config map[string]interface{}) *terraform.InstanceState {
	t.Helper()
	ctx := context.Background()

	priorRD := schema.TestResourceDataRaw(t, res.Schema, prior)
	priorRD.SetId("resource-uid")
	state := priorRD.State()
	require.NotNil(t, state)

	diff, err := res.Diff(ctx, state, terraform.NewResourceConfigRaw(config), unitTestMockAPINegativeClient)
	require.NoError(t, err)
	require.NotNil(t, diff, "precondition: the plan has changes to apply")

	newState, diags := res.Apply(ctx, state, diff, unitTestMockAPINegativeClient)
	require.True(t, diags.HasError(), "precondition: the update must fail")
	require.NotNil(t, newState)
	assert.Equal(t, "resource-uid", newState.ID)
	return newState
}

func TestCloudAccountAwsUpdateFailureKeepsPriorExternalID(t *testing.T) {
	account := func(externalID string) map[string]interface{} {
		return map[string]interface{}{
			"name":        "aws-sts-account",
			"context":     "project",
			"type":        "sts",
			"arn":         "arn:aws:iam::123456789012:role/palette",
			"external_id": externalID,
		}
	}
	newState := applyFailedUpdate(t, resourceCloudAccountAws(), account("tenant-external-id"), account("rejected-external-id"))
	assert.Equal(t, "tenant-external-id", newState.Attributes["external_id"])
}

func TestAuditTrailUpdateFailureKeepsPriorExternalID(t *testing.T) {
	trail := func(externalID string) map[string]interface{} {
		return map[string]interface{}{
			"name": "cloudwatch-trail",
			"type": "cloudwatch",
			"cloudwatch": []interface{}{map[string]interface{}{
				"group":           "palette-audit",
				"region":          "us-east-1",
				"stream":          "audit",
				"credential_type": "sts",
				"arn":             "arn:aws:iam::123456789012:role/palette",
				"external_id":     externalID,
			}},
		}
	}
	newState := applyFailedUpdate(t, resourceAuditTrail(), trail("tenant-external-id"), trail("rejected-external-id"))
	assert.Equal(t, "tenant-external-id", newState.Attributes["cloudwatch.0.external_id"])
}
