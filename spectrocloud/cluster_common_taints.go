package spectrocloud

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func toClusterTaints(m map[string]interface{}) []*models.V1Taint {
	clusterTaints := make([]*models.V1Taint, 0)
	if m["taints"] == nil {
		return nil
	}
	for _, clusterTaint := range m["taints"].([]interface{}) {
		b := toClusterTaint(clusterTaint)
		clusterTaints = append(clusterTaints, b)
	}

	return clusterTaints
}

func toClusterTaint(clusterTaint interface{}) *models.V1Taint {
	m := clusterTaint.(map[string]interface{})

	key, _ := m["key"].(string)
	value, _ := m["value"].(string)
	effect, _ := m["effect"].(string)

	ret := &models.V1Taint{
		Effect: effect,
		Key:    key,
		Value:  value,
	}

	return ret
}

func flattenClusterTaints(items []*models.V1Taint) []interface{} {
	result := make([]interface{}, 0)
	for _, taint := range items {
		flattenTaint := make(map[string]interface{})

		flattenTaint["key"] = taint.Key
		flattenTaint["value"] = taint.Value
		flattenTaint["effect"] = taint.Effect

		result = append(result, flattenTaint)
	}
	return result
}

func FlattenAdditionalLabelsAndTaints(labels map[string]string, intaints []*models.V1Taint, oi map[string]interface{}) {
	if len(labels) == 0 {
		oi["additional_labels"] = make(map[string]interface{})
	} else {
		oi["additional_labels"] = labels
	}

	taints := flattenClusterTaints(intaints)
	if len(taints) > 0 {
		oi["taints"] = taints
	}
}
