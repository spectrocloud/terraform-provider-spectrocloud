package client

import (
	"errors"
	"github.com/spectrocloud/hapi/models"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdateClusterProfile(t *testing.T) {
	testCases := []struct {
		name               string
		clusterProfile     *models.V1ClusterProfileUpdateEntity
		ProfileContext     string
		expectedError      error
		getClientError     error
		updateError        error
		GetClusterClientFn func() (clusterC.ClientService, error)
		v1ClusterUpdateFn  func(params *clusterC.V1ClusterProfilesUpdateParams) (*clusterC.V1ClusterProfilesUpdateNoContent, error)
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
			updateError:    nil,
			v1ClusterUpdateFn: func(params *clusterC.V1ClusterProfilesUpdateParams) (*clusterC.V1ClusterProfilesUpdateNoContent, error) {
				// Mock implementation of V1ClusterProfilesUpdate goes here
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
			expectedError:  errors.New("error updating cluster profile"),
			getClientError: nil,
			updateError:    errors.New("error updating cluster profile"),
			v1ClusterUpdateFn: func(params *clusterC.V1ClusterProfilesUpdateParams) (*clusterC.V1ClusterProfilesUpdateNoContent, error) {
				// Mock implementation of V1ClusterProfilesUpdate goes here
				return nil, errors.New("error updating cluster profile")
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
			v1ClusterUpdateFn: func(params *clusterC.V1ClusterProfilesUpdateParams) (*clusterC.V1ClusterProfilesUpdateNoContent, error) {
				// Mock implementation of V1ClusterProfilesUpdate goes here
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
			v1ClusterUpdateFn: func(params *clusterC.V1ClusterProfilesUpdateParams) (*clusterC.V1ClusterProfilesUpdateNoContent, error) {
				// Mock implementation of V1ClusterProfilesUpdate goes here
				return nil, nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := &V1Client{
				GetClusterClientFn: func() (clusterC.ClientService, error) {
					return &clusterC.Client{}, tc.getClientError
				},
				v1ClusterProfilesUpdateFn: tc.v1ClusterUpdateFn,
			}

			err := h.UpdateClusterProfile(tc.clusterProfile, tc.ProfileContext)

			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
