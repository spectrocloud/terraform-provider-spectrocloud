package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
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

func TestResourceApplicationCreate(t *testing.T) {
	d := prepareBaseResourceApplicationData()

	diags := resourceApplicationCreate(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}

func TestResourceApplicationRead(t *testing.T) {
	d := prepareBaseResourceApplicationData()

	diags := resourceApplicationRead(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}

func TestResourceApplicationUpdate(t *testing.T) {
	d := prepareBaseResourceApplicationData()

	diags := resourceApplicationUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}

func TestResourceApplicationDelete(t *testing.T) {
	d := prepareBaseResourceApplicationData()

	diags := resourceApplicationDelete(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}
