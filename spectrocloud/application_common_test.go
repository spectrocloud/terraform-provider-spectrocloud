package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func prepareApplicationTestData(id string) *schema.ResourceData {
	d := resourceApplication().TestResourceData()
	d.SetId(id)
	return d
}

func TestResourceApplicationStateRefreshFunc(t *testing.T) {
	var cases []struct {
		name           string
		client         *client.V1Client
		schemaDiags    *schema.ResourceData
		retry          int
		duration       int
		expectedResult interface{}
		statusString   string
		errorMessage   error
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			refreshFunc := resourceApplicationStateRefreshFunc(tc.client, tc.schemaDiags, tc.retry, 1)
			result, statusString, errorMessage := refreshFunc()
			if tc.statusString == "Tier:Error" {
				if statusString != tc.statusString {
					t.Errorf("Expected %v, got %v", tc.statusString, statusString)
				}
				if errorMessage.Error() != tc.errorMessage.Error() {
					t.Errorf("Expected %v, got %v", tc.errorMessage.Error(), errorMessage.Error())
				}
			} else {
				if result != tc.expectedResult {
					t.Errorf("Expected %v, got %v", tc.expectedResult, result)
				}
			}
		})
	}
}
