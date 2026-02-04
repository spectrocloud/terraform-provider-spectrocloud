package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"testing"
)

func prepareBaseFilterTestData() *schema.ResourceData {
	d := resourceFilter().TestResourceData()
	_ = d.Set("metadata", []interface{}{
		map[string]interface{}{
			"name": "test-filter-name",
		},
	})
	_ = d.Set("spec", []interface{}{
		map[string]interface{}{
			"filter_group": []interface{}{
				map[string]interface{}{
					"conjunction": "AND",
					"filters": []interface{}{
						map[string]interface{}{
							"key":      "test-key",
							"negation": false,
							"operator": "eq",
							"values":   []string{"test-value"},
						},
					},
				},
			},
		},
	})
	d.SetId("test-filter-id")
	return d
}

func TestResourceFilterCRUD(t *testing.T) {
	testResourceCRUD(t, prepareBaseFilterTestData, unitTestMockAPIClient,
		resourceFilterCreate, resourceFilterRead, resourceFilterUpdate, resourceFilterDelete)
}
