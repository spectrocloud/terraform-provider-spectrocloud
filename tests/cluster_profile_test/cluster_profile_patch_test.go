package cluster_profile_test

import (
	"errors"
	"testing"

	"github.com/spectrocloud/palette-api-go/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mock"
	"github.com/stretchr/testify/assert"
)

func TestPatchClusterProfile(t *testing.T) {
	testCases := []struct {
		name           string
		clusterProfile *models.V1ClusterProfileUpdateEntity
		ProfileContext string
		expectedError  error
		mock           *mock.ClusterClientMock
	}{
		{
			name: "Success",
			clusterProfile: &models.V1ClusterProfileUpdateEntity{
				Metadata: &models.V1ObjectMeta{
					UID: "1",
				},
			},
			ProfileContext: "project",
			expectedError:  nil,
			mock: &mock.ClusterClientMock{
				PatchClusterProfileErr: nil,
			},
		},
		{
			name: "Error",
			clusterProfile: &models.V1ClusterProfileUpdateEntity{
				Metadata: &models.V1ObjectMeta{
					UID: "2",
				},
			},
			ProfileContext: "tenant",
			expectedError:  errors.New("error patching cluster profile"),
			mock: &mock.ClusterClientMock{
				PatchClusterProfileErr: errors.New("error patching cluster profile"),
			},
		},
		{
			name: "Invalid scope",
			clusterProfile: &models.V1ClusterProfileUpdateEntity{
				Metadata: &models.V1ObjectMeta{
					UID: "3",
				},
			},
			ProfileContext: "invalid",
			expectedError:  errors.New("invalid scope"),
			mock:           &mock.ClusterClientMock{},
		},
	}

	metadata := &models.V1ProfileMetaEntity{
		Metadata: &models.V1ObjectMetaInputEntity{
			Annotations: map[string]string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := &client.V1Client{}
			err := h.PatchClusterProfile(tc.mock, tc.clusterProfile, metadata, tc.ProfileContext)
			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
