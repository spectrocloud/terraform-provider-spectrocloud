package spectrocloud

func toAdditionalNodePoolLabels(m map[string]interface{}) map[string]string {
	additionalLabels := make(map[string]string)
	if m["additional_labels"] != nil && len(m["additional_labels"].(map[string]interface{})) > 0 {
		additionalLabels = expandStringMap(m["additional_labels"].(map[string]interface{}))
	}
	return additionalLabels
}
