package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func expandMetadata(list []interface{}) *models.V1ObjectMetaInputEntity {
	if len(list) == 0 || list[0] == nil {
		return nil
	}

	m := list[0].(map[string]interface{})

	return &models.V1ObjectMetaInputEntity{
		Name: m["name"].(string),
	}
}

func expandSpec(list []interface{}) *models.V1TagFilterSpec {
	if len(list) == 0 || list[0] == nil {
		return nil
	}

	m := list[0].(map[string]interface{})
	filterGroup := m["filter_group"].([]interface{})

	return &models.V1TagFilterSpec{
		FilterGroup: expandFilterGroup(filterGroup),
	}
}

func expandFilterGroup(list []interface{}) *models.V1TagFilterGroup {
	if len(list) == 0 || list[0] == nil {
		return nil
	}

	m := list[0].(map[string]interface{})

	// Handle both TypeSet and TypeList for backward compatibility
	var filtersList []interface{}
	if filtersRaw, ok := m["filters"]; ok && filtersRaw != nil {
		if filtersSet, ok := filtersRaw.(*schema.Set); ok {
			filtersList = filtersSet.List()
		} else if filtersListRaw, ok := filtersRaw.([]interface{}); ok {
			// Fallback for backward compatibility during migration
			filtersList = filtersListRaw
		}
	}

	// filters := m["filters"].([]interface{})

	conjunction := models.V1SearchFilterConjunctionOperator(m["conjunction"].(string))

	return &models.V1TagFilterGroup{
		Conjunction: &conjunction,
		Filters:     expandFilters(filtersList),
	}
}

func expandFilters(list []interface{}) []*models.V1TagFilterItem {
	filters := make([]*models.V1TagFilterItem, len(list))

	for i, item := range list {
		m := item.(map[string]interface{})
		var values []string
		if m["values"] != nil {
			interfaceValues := m["values"].([]interface{})
			for _, v := range interfaceValues {
				values = append(values, v.(string))
			}
		}

		filters[i] = &models.V1TagFilterItem{
			Key:      m["key"].(string),
			Negation: m["negation"].(bool),
			Operator: models.V1SearchFilterKeyValueOperator(m["operator"].(string)),
			Values:   values,
		}
	}

	return filters
}

func flattenMetadata(metadata *models.V1ObjectMeta) []interface{} {
	if metadata == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"name": metadata.Name,
	}

	return []interface{}{m}
}

func flattenSpec(spec *models.V1TagFilterSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"filter_group": flattenFilterGroup(spec.FilterGroup),
	}

	return []interface{}{m}
}

func flattenFilters(filters []*models.V1TagFilterItem) []interface{} {
	if filters == nil {
		return []interface{}{}
	}

	fs := make([]interface{}, len(filters))

	for i, filter := range filters {
		m := map[string]interface{}{
			"key":      filter.Key,
			"negation": filter.Negation,
			"operator": string(filter.Operator),
			"values":   filter.Values,
		}
		fs[i] = m
	}

	return fs
}

func flattenFilterGroup(filterGroup *models.V1TagFilterGroup) []interface{} {
	if filterGroup == nil {
		return []interface{}{}
	}
	// Convert filters to schema.Set for proper TypeSet handling
	var filtersSet *schema.Set
	if filterGroup.Filters != nil && len(filterGroup.Filters) > 0 {
		filtersList := make([]interface{}, len(filterGroup.Filters))
		for i, filter := range filterGroup.Filters {
			filtersList[i] = map[string]interface{}{
				"key":      filter.Key,
				"negation": filter.Negation,
				"operator": string(filter.Operator),
				"values":   filter.Values,
			}
		}
		// Create a schema.Set using the hash function
		filtersSet = schema.NewSet(resourceFilterItemHash, filtersList)
	} else {
		filtersSet = schema.NewSet(resourceFilterItemHash, []interface{}{})
	}

	m := map[string]interface{}{
		"conjunction": string(*filterGroup.Conjunction),
		"filters":     flattenFilters(filterGroup.Filters),
	}

	return []interface{}{m}
}
