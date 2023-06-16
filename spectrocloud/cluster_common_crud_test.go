package spectrocloud

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mock"
)

func prepareClusterTestData() *schema.ResourceData {
	d := resourceClusterEks().TestResourceData()
	d.Set("context", "project")

	return d
}

func TestResourceClusterDelete(t *testing.T) {
	testCases := []struct {
		name                  string
		forceDeleteTimeout    string
		expectedReturnedDiags diag.Diagnostics
		mock                  *mock.ClusterClientMock
	}{
		{
			name:                  "ForceDeleteClusterErr",
			forceDeleteTimeout:    "2m",
			expectedReturnedDiags: diag.FromErr(errors.New("covering error case")),
			mock: &mock.ClusterClientMock{
				DeleteClusterErr:      nil,
				ForceDeleteClusterErr: errors.New("covering error case"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := prepareClusterTestData()
			d.Set("force_delete_timeout", tc.forceDeleteTimeout)

			h := &client.V1Client{
				GetClusterFn: func(uid string) (*models.V1SpectroCluster, error) {
					isHost := new(bool)
					*isHost = true
					cluster := &models.V1SpectroCluster{
						APIVersion: "v1",
						Metadata: &models.V1ObjectMeta{
							Annotations:       map[string]string{"scope": "project"},
							CreationTimestamp: models.V1Time{},
							DeletionTimestamp: models.V1Time{},
							Labels: map[string]string{
								"owner": "siva",
							},
							LastModifiedTimestamp: models.V1Time{},
							Name:                  "test-vsphere-cluster-unit-test",
							Namespace:             "",
							ResourceVersion:       "",
							SelfLink:              "",
							UID:                   "vsphere-uid",
						},
						Spec: &models.V1SpectroClusterSpec{
							CloudConfigRef: &models.V1ObjectReference{
								APIVersion:      "",
								FieldPath:       "",
								Kind:            "",
								Name:            "",
								Namespace:       "",
								ResourceVersion: "",
								UID:             "test-cloud-config-uid",
							},
							CloudType: "",
							ClusterConfig: &models.V1ClusterConfig{
								ClusterRbac:                    nil,
								ClusterResources:               nil,
								ControlPlaneHealthCheckTimeout: "",
								Fips:                           nil,
								HostClusterConfig: &models.V1HostClusterConfig{
									ClusterEndpoint: &models.V1HostClusterEndpoint{
										Config: nil,
										Type:   "LoadBalancer",
									},
									ClusterGroup:  nil,
									HostCluster:   nil,
									IsHostCluster: isHost,
								},
								LifecycleConfig:             nil,
								MachineHealthConfig:         nil,
								MachineManagementConfig:     nil,
								UpdateWorkerPoolsInParallel: false,
							},
							ClusterProfileTemplates: nil,
							ClusterType:             "",
						},
						Status: &models.V1SpectroClusterStatus{
							State: "running",
						},
					}
					return cluster, nil
				},
				ClusterC: tc.mock,
			}

			ctx := context.Background()
			diags := resourceClusterDelete(ctx, d, h)

			if len(diags) != len(tc.expectedReturnedDiags) {
				t.Fail()
				t.Logf("Expected diags count: %v", len(tc.expectedReturnedDiags))
				t.Logf("Actual diags count: %v", len(diags))
			} else {
				for i := range diags {
					if diags[i].Severity != tc.expectedReturnedDiags[i].Severity {
						t.Fail()
						t.Logf("Expected severity: %v", tc.expectedReturnedDiags[i].Severity)
						t.Logf("Actual severity: %v", diags[i].Severity)
					}
					if diags[i].Summary != tc.expectedReturnedDiags[i].Summary {
						t.Fail()
						t.Logf("Expected summary: %v", tc.expectedReturnedDiags[i].Summary)
						t.Logf("Actual summary: %v", diags[i].Summary)
					}
					if diags[i].Detail != tc.expectedReturnedDiags[i].Detail {
						t.Fail()
						t.Logf("Expected detail: %v", tc.expectedReturnedDiags[i].Detail)
						t.Logf("Actual detail: %v", diags[i].Detail)
					}
				}
			}
		})
	}
}
