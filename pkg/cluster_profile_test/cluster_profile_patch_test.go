package cluster_profile_test

import (
	"errors"
	"github.com/spectrocloud/hapi/models"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPatchClusterProfile(t *testing.T) {
	testCases := []struct {
		name                                 string
		clusterProfile                       *models.V1ClusterProfileUpdateEntity
		ProfileContext                       string
		expectedError                        error
		getClientError                       error
		patchError                           error
		GetClusterClientFn                   func() (clusterC.ClientService, error)
		v1ClusterProfilesUIDMetadataUpdateFn func(params *clusterC.V1ClusterProfilesUIDMetadataUpdateParams) (*clusterC.V1ClusterProfilesUIDMetadataUpdateNoContent, error)
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
			getClientError: nil,
			patchError:     nil,
			v1ClusterProfilesUIDMetadataUpdateFn: func(params *clusterC.V1ClusterProfilesUIDMetadataUpdateParams) (*clusterC.V1ClusterProfilesUIDMetadataUpdateNoContent, error) { // Mock implementation of V1ClusterProfilesUIDMetadataUpdate goes here
				return nil, nil
			},
		},
		{
			name: "Success",
			clusterProfile: &models.V1ClusterProfileUpdateEntity{
				Metadata: &models.V1ObjectMeta{
					UID: "2",
				},
			},
			ProfileContext: "tenant",
			expectedError:  errors.New("error patching cluster profile"),
			getClientError: nil,
			patchError:     errors.New("error patching cluster profile"),
			v1ClusterProfilesUIDMetadataUpdateFn: func(params *clusterC.V1ClusterProfilesUIDMetadataUpdateParams) (*clusterC.V1ClusterProfilesUIDMetadataUpdateNoContent, error) {
				// Mock implementation of V1ClusterProfilesUIDMetadataUpdate goes here
				return nil, errors.New("error patching cluster profile")
			},
		},
		{
			name: "GetClientError",
			clusterProfile: &models.V1ClusterProfileUpdateEntity{
				Metadata: &models.V1ObjectMeta{
					UID: "3",
				},
			},
			ProfileContext: "project",
			expectedError:  errors.New("GetClientError"),
			getClientError: errors.New("GetClientError"),
			v1ClusterProfilesUIDMetadataUpdateFn: func(params *clusterC.V1ClusterProfilesUIDMetadataUpdateParams) (*clusterC.V1ClusterProfilesUIDMetadataUpdateNoContent, error) {
				// Mock implementation of V1ClusterProfilesUIDMetadataUpdate goes here
				return nil, nil
			},
		},
		{
			name: "Invalid scope",

			clusterProfile: &models.V1ClusterProfileUpdateEntity{
				Metadata: &models.V1ObjectMeta{
					UID: "4",
				},
			},
			ProfileContext: "invalid",
			expectedError:  errors.New("invalid scope"),
			getClientError: nil,
			v1ClusterProfilesUIDMetadataUpdateFn: func(params *clusterC.V1ClusterProfilesUIDMetadataUpdateParams) (*clusterC.V1ClusterProfilesUIDMetadataUpdateNoContent, error) {
				// Mock implementation of V1ClusterProfilesUIDMetadataUpdate goes here
				return nil, nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := &client.V1Client{
				GetClusterClientFn: func() (clusterC.ClientService, error) {
					return &clusterC.Client{}, tc.getClientError
				},
				V1ClusterProfilesUIDMetadataUpdateFn: tc.v1ClusterProfilesUIDMetadataUpdateFn,
			}

			err := h.PatchClusterProfile(tc.clusterProfile, &models.V1ProfileMetaEntity{
				Metadata: &models.V1ObjectMetaInputEntity{
					Annotations: map[string]string{},
				},
			}, tc.ProfileContext)

			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
