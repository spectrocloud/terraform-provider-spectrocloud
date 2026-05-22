package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigureFeatureFlags(t *testing.T) {
	t.Cleanup(func() {
		disableAddonDeploymentResource = false
	})

	t.Run("defaults to disabled flag false", func(t *testing.T) {
		d := prepareProviderConfigWithFeatureFlags(nil)
		configureFeatureFlags(d)
		assert.False(t, disableAddonDeploymentResource)
	})

	t.Run("enables disable_addon_deployment_resource", func(t *testing.T) {
		d := prepareProviderConfigWithFeatureFlags(map[string]interface{}{
			featureFlagDisableAddonDeploymentResource: true,
		})
		configureFeatureFlags(d)
		assert.True(t, disableAddonDeploymentResource)
	})

	t.Run("ignores unknown feature flags", func(t *testing.T) {
		d := prepareProviderConfigWithFeatureFlags(map[string]interface{}{
			"future_flag": true,
		})
		configureFeatureFlags(d)
		assert.False(t, disableAddonDeploymentResource)
	})

	t.Run("provider configure resets flag when omitted", func(t *testing.T) {
		disableAddonDeploymentResource = true
		d := prepareBaseProviderConfig()
		configureFeatureFlags(d)
		assert.False(t, disableAddonDeploymentResource)
	})
}

func TestAddonDeploymentBlockedByFeatureFlag(t *testing.T) {
	t.Cleanup(func() {
		disableAddonDeploymentResource = false
	})
	disableAddonDeploymentResource = true

	d := prepareAddonDeploymentTestData("cluster-123_profile-1")
	require.NotNil(t, d)

	err := resourceAddonDeploymentCustomizeDiff(context.Background(), nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), featureFlagDisableAddonDeploymentResource)

	diags := resourceAddonDeploymentRead(context.Background(), d, unitTestMockAPIClient)
	assert.NotEmpty(t, diags)
	assert.Contains(t, diags[0].Summary+diags[0].Detail, featureFlagDisableAddonDeploymentResource)

	diags = resourceAddonDeploymentDelete(context.Background(), d, unitTestMockAPIClient)
	assert.NotEmpty(t, diags)
	assert.Contains(t, diags[0].Summary+diags[0].Detail, featureFlagDisableAddonDeploymentResource)
}

func prepareProviderConfigWithFeatureFlags(flags map[string]interface{}) *schema.ResourceData {
	d := prepareBaseProviderConfig()
	if flags != nil {
		_ = d.Set("feature_flag", flags)
	}
	return d
}
