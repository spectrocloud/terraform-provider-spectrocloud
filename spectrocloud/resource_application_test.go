package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func prepareBaseResourceApplicationData() *schema.ResourceData {
	d := resourceApplication().TestResourceData()
	return d
}

//func TestResourceApplicationCreate(t *testing.T) {
//	d := prepareBaseResourceApplicationData()
//	_ = d.Set("config", []interface{}{
//		map[string]string{
//			"cluster_uid": "test-cluster-uid",
//		},
//	})
//	_ = d.Set("name", "test-application")
//	diags := resourceApplicationCreate(context.Background(), d, unitTestMockAPIClient)
//	assert.Empty(t, diags)
//}
