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

// toUpdateStrategy builds the V1UpdateStrategy model from the machine pool configuration.
// Handles both simple update_strategy and OverrideScaling with maxSurge and maxUnavailable.
func toUpdateStrategy(m map[string]interface{}) *models.V1UpdateStrategy {
	strategy := &models.V1UpdateStrategy{
		Type: getUpdateStrategy(m),
	}

	// If using OverrideScaling, populate maxSurge and maxUnavailable from override_scaling
	if strategy.Type == "OverrideScaling" {
		if overrideScaling, ok := m["override_scaling"].([]interface{}); ok && len(overrideScaling) > 0 {
			scalingConfig := overrideScaling[0].(map[string]interface{})
			if maxSurge, ok := scalingConfig["max_surge"].(string); ok {
				strategy.MaxSurge = maxSurge
			}
			if maxUnavailable, ok := scalingConfig["max_unavailable"].(string); ok {
				strategy.MaxUnavailable = maxUnavailable
			}
		}
	}

	return strategy
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

// flattenOverrideScaling flattens the MaxSurge and MaxUnavailable values from V1UpdateStrategy
// to the override_scaling Terraform field.
func flattenOverrideScaling(updateStrategy *models.V1UpdateStrategy, oi map[string]interface{}) {
	if updateStrategy != nil && updateStrategy.Type == "OverrideScaling" {
		if updateStrategy.MaxSurge != "" || updateStrategy.MaxUnavailable != "" {
			overrideScaling := make(map[string]interface{})
			overrideScaling["max_surge"] = updateStrategy.MaxSurge
			overrideScaling["max_unavailable"] = updateStrategy.MaxUnavailable
			oi["override_scaling"] = []interface{}{overrideScaling}
		}
	}
}
