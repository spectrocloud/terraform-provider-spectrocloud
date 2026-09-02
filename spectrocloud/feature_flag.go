package spectrocloud

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const featureFlagDisableAddonDeploymentResource = "disable_addon_deployment_resource"

// disableAddonDeploymentResource is set from provider feature_flag configuration.
var disableAddonDeploymentResource bool

const addonDeploymentResourceDisabledMessage = "spectrocloud_addon_deployment is disabled by provider feature flag " +
	"`disable_addon_deployment_resource`. Remove this resource from configuration or set the flag to false."

func configureFeatureFlags(d *schema.ResourceData) {
	disableAddonDeploymentResource = false

	raw, ok := d.GetOk("feature_flag")
	if !ok {
		return
	}

	flags, ok := raw.(map[string]interface{})
	if !ok {
		return
	}

	if v, ok := flags[featureFlagDisableAddonDeploymentResource]; ok {
		if disabled, ok := v.(bool); ok {
			disableAddonDeploymentResource = disabled
		}
	}
}

func addonDeploymentResourceDisabled() bool {
	return disableAddonDeploymentResource
}

func addonDeploymentResourceDisabledError() error {
	return errors.New(addonDeploymentResourceDisabledMessage)
}
