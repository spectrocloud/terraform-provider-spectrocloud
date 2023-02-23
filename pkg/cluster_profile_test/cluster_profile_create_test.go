package cluster_profile_test

import (
	"errors"
	"testing"

	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/client"
	"github.com/spectrocloud/hapi/models"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
	"github.com/stretchr/testify/assert"
)

func TestCreateClusterProfile(t *testing.T) {
	testCases := []struct {
		name                      string
		clusterProfile            *models.V1ClusterProfileEntity
		profileContext            string
		expectedError             error
		expectedReturnedUID       string
		getClientError            error
		v1ClusterProfilesCreateFn func(params *clusterC.V1ClusterProfilesCreateParams) (*clusterC.V1ClusterProfilesCreateCreated, error)
	}{
		{
			name:                "Success",
			clusterProfile:      &models.V1ClusterProfileEntity{},
			profileContext:      "project",
			expectedError:       nil,
			expectedReturnedUID: "1",
			getClientError:      nil,
			v1ClusterProfilesCreateFn: func(params *clusterC.V1ClusterProfilesCreateParams) (*clusterC.V1ClusterProfilesCreateCreated, error) {
				response := &clusterC.V1ClusterProfilesCreateCreated{Payload: &models.V1UID{UID: ptr.StringPtr("1")}}
				return response, nil
			},
		},
		{
			name:                "Success",
			clusterProfile:      &models.V1ClusterProfileEntity{},
			profileContext:      "tenant",
			expectedError:       nil,
			expectedReturnedUID: "2",
			getClientError:      nil,
			v1ClusterProfilesCreateFn: func(params *clusterC.V1ClusterProfilesCreateParams) (*clusterC.V1ClusterProfilesCreateCreated, error) {
				response := &clusterC.V1ClusterProfilesCreateCreated{Payload: &models.V1UID{UID: ptr.StringPtr("2")}}
				return response, nil
			},
		},
		{
			name:           "Error",
			clusterProfile: &models.V1ClusterProfileEntity{},
			profileContext: "tenant",
			expectedError:  errors.New("error creating cluster profile"),
			getClientError: nil,
			v1ClusterProfilesCreateFn: func(params *clusterC.V1ClusterProfilesCreateParams) (*clusterC.V1ClusterProfilesCreateCreated, error) {
				// Mock implementation of V1ClusterProfilesCreate goes here
				return nil, errors.New("error creating cluster profile")
			},
		},
		{
			name:           "GetClientError",
			clusterProfile: &models.V1ClusterProfileEntity{},
			profileContext: "project",
			expectedError:  errors.New("GetClientError"),

			getClientError: errors.New("GetClientError"),
			v1ClusterProfilesCreateFn: func(params *clusterC.V1ClusterProfilesCreateParams) (*clusterC.V1ClusterProfilesCreateCreated, error) {
				// Mock implementation of V1ClusterProfilesCreate goes here
				return nil, nil
			},
		},
		{
			name:           "Invalid scope",
			clusterProfile: &models.V1ClusterProfileEntity{},
			profileContext: "invalid",
			expectedError:  errors.New("invalid scope"),
			getClientError: nil,
			v1ClusterProfilesCreateFn: func(params *clusterC.V1ClusterProfilesCreateParams) (*clusterC.V1ClusterProfilesCreateCreated, error) {
				// Mock implementation of V1ClusterProfilesCreate goes here
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
				V1ClusterProfilesCreateFn: tc.v1ClusterProfilesCreateFn,
			}
			id, err := h.CreateClusterProfile(tc.clusterProfile, tc.profileContext)
			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			if tc.expectedReturnedUID != "" {
				assert.Equal(t, id, tc.expectedReturnedUID)
			}
		})
	}
}
