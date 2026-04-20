package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func prepareBaseResourceApplicationData() *schema.ResourceData {
	d := resourceApplication().TestResourceData()
	d.SetId("test-application-id")
	_ = d.Set("name", "test-application")
	_ = d.Set("tags", []string{"test:dev"})
	_ = d.Set("application_profile_uid", "test-application-profile-id")
	var con []interface{}
	con = append(con, map[string]interface{}{
		"cluster_uid":       "test-cluster-id",
		"cluster_group_uid": "test-cluster-group-id",
		"cluster_context":   "project",
		"cluster_name":      "test-cluster",
		"limits": []interface{}{
			map[string]interface{}{
				"cpu":     2,
				"memory":  1000,
				"storage": 100,
			},
		},
	})

	_ = d.Set("config", con)
	return d
}

func TestResourceApplicationCRUD(t *testing.T) {
	testResourceCRUD(t, prepareBaseResourceApplicationData, unitTestMockAPIClient,
		resourceApplicationCreate, resourceApplicationRead, resourceApplicationUpdate, resourceApplicationDelete)
}
