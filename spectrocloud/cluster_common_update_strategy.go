package spectrocloud

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// getUpdateStrategy returns the update strategy string from the machine pool configuration.
// This is a simple helper to extract the update_strategy field.
func getUpdateStrategy(m map[string]interface{}) string {
	if updateStrategy, ok := m["update_strategy"].(string); ok && updateStrategy != "" {
		return updateStrategy
	}
	return "RollingUpdateScaleOut" // Default value
}

// flattenUpdateStrategy flattens the SDK UpdateStrategy to Terraform state.
// Sets the update_strategy field from the API response.
func flattenUpdateStrategy(updateStrategy *models.V1UpdateStrategy, oi map[string]interface{}) {
	if updateStrategy != nil && updateStrategy.Type != "" {
		oi["update_strategy"] = updateStrategy.Type
	} else {
		// Set default if no strategy is provided
		oi["update_strategy"] = "RollingUpdateScaleOut"
	}
}
