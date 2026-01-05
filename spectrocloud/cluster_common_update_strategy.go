package spectrocloud

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// getUpdateStrategy returns the UpdateStrategy model based on either the new rolling_update_strategy
// or the legacy update_strategy field. Prioritizes rolling_update_strategy for backward compatibility.
func getUpdateStrategy(m map[string]interface{}) string {
	updateStrategy := ""

	// Check for new rolling_update_strategy first
	if rollingUpdateList, ok := m["rolling_update_strategy"].([]interface{}); ok && len(rollingUpdateList) > 0 {
		if rollingUpdate, ok := rollingUpdateList[0].(map[string]interface{}); ok {
			if strategyType, ok := rollingUpdate["type"].(string); ok && strategyType != "" {
				updateStrategy = strategyType
			}
		}
	}

	// Fall back to legacy update_strategy if rolling_update_strategy didn't provide a valid value
	if updateStrategy == "" && m["update_strategy"] != nil {
		if strategyValue, ok := m["update_strategy"].(string); ok && strategyValue != "" {
			updateStrategy = strategyValue
		}
	}

	// Default to RollingUpdateScaleOut if no valid strategy was found
	if updateStrategy == "" {
		updateStrategy = "RollingUpdateScaleOut"
	}

	return updateStrategy
}

// toUpdateStrategy converts Terraform rolling_update_strategy or update_strategy to SDK UpdateStrategy model.
// Supports both new (rolling_update_strategy) and legacy (update_strategy) fields for backward compatibility.
func toUpdateStrategy(m map[string]interface{}) *models.V1UpdateStrategy {
	updateStrategy := &models.V1UpdateStrategy{
		Type: getUpdateStrategy(m),
	}

	// Check for new rolling_update_strategy with maxSurge and maxUnavailable
	if rollingUpdateList, ok := m["rolling_update_strategy"].([]interface{}); ok && len(rollingUpdateList) > 0 {
		if rollingUpdate, ok := rollingUpdateList[0].(map[string]interface{}); ok {
			if maxSurge, ok := rollingUpdate["max_surge"].(string); ok && maxSurge != "" {
				updateStrategy.MaxSurge = maxSurge
			}

			if maxUnavailable, ok := rollingUpdate["max_unavailable"].(string); ok && maxUnavailable != "" {
				updateStrategy.MaxUnavailable = maxUnavailable
			}
		}
	}

	return updateStrategy
}

// flattenUpdateStrategy flattens the SDK UpdateStrategy to Terraform state.
// Populates both update_strategy (legacy) and rolling_update_strategy (new) for backward compatibility.
func flattenUpdateStrategy(updateStrategy *models.V1UpdateStrategy, oi map[string]interface{}) {
	if updateStrategy == nil || updateStrategy.Type == "" {
		// Set defaults
		oi["update_strategy"] = "RollingUpdateScaleOut"
		return
	}

	// Always set legacy update_strategy field for backward compatibility
	oi["update_strategy"] = updateStrategy.Type

	// Set rolling_update_strategy if there are additional fields or if it's already in use
	// Check if user is using the new field by seeing if it exists in the original config
	if _, hasRollingUpdate := oi["rolling_update_strategy"]; hasRollingUpdate || updateStrategy.MaxSurge != "" || updateStrategy.MaxUnavailable != "" {
		rollingUpdate := map[string]interface{}{
			"type": updateStrategy.Type,
		}

		if updateStrategy.MaxSurge != "" {
			rollingUpdate["max_surge"] = updateStrategy.MaxSurge
		}

		if updateStrategy.MaxUnavailable != "" {
			rollingUpdate["max_unavailable"] = updateStrategy.MaxUnavailable
		}

		oi["rolling_update_strategy"] = []interface{}{rollingUpdate}
	}
}
