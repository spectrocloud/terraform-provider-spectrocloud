package spectrocloud

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func prepareApplicationTestData(id string) *schema.ResourceData {
	d := resourceApplication().TestResourceData()
	d.SetId(id)
	return d
}

func TestResourceApplicationStateRefreshFunc(t *testing.T) {
	cases := []struct {
		name            string
		client          *client.V1Client
		schemaDiags     *schema.ResourceData
		retry           int
		duration        int
		expected_result interface{}
		status_string   string
		error_message   error
	}{
		{
			name: "tier error",
			client: &client.V1Client{
				GetApplicationFn: func(id string) (*models.V1AppDeployment, error) {
					return &models.V1AppDeployment{
						Status: &models.V1AppDeploymentStatus{
							AppTiers: []*models.V1ClusterPackStatus{
								{
									Name: "test",
									Condition: &models.V1ClusterCondition{
										Type:    types.Ptr("Error"),
										Status:  types.Ptr("True"),
										Message: "error message",
									},
								},
							},
							State: "NotDeployed",
						},
					}, nil
				},
			},
			schemaDiags:     prepareApplicationTestData("test_id"),
			retry:           5,
			duration:        1,
			expected_result: nil,
			status_string:   "Tier:Error",
			error_message:   errors.New("error message"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			refreshFunc := resourceApplicationStateRefreshFunc(tc.client, tc.schemaDiags, tc.retry, 1)
			result, status_string, error_message := refreshFunc()
			if tc.status_string == "Tier:Error" {
				if status_string != tc.status_string {
					t.Errorf("Expected %v, got %v", tc.status_string, status_string)
				}
				if error_message.Error() != tc.error_message.Error() {
					t.Errorf("Expected %v, got %v", tc.error_message.Error(), error_message.Error())
				}
			} else {
				if result != tc.expected_result {
					t.Errorf("Expected %v, got %v", tc.expected_result, result)
				}
			}
		})
	}
}
