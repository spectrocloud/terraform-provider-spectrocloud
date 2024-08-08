package spectrocloud

import (
	"github.com/spectrocloud/palette-api-go/models"
)

func getUpdateStrategy(m map[string]interface{}) string {
	updateStrategy := "RollingUpdateScaleOut"
	if m["update_strategy"] != nil {
		updateStrategy = m["update_strategy"].(string)
	}
	return updateStrategy
}

func flattenUpdateStrategy(updateStrategy *models.V1UpdateStrategy, oi map[string]interface{}) {
	if updateStrategy != nil && updateStrategy.Type != "" {
		oi["update_strategy"] = updateStrategy.Type
	} else {
		oi["update_strategy"] = "RollingUpdateScaleOut"
	}
}
