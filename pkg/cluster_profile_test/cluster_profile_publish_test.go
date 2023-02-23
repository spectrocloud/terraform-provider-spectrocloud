package cluster_profile_test

import (
	"errors"
	"testing"

	"github.com/spectrocloud/hapi/models"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schema"
)

func TestPublishClusterProfile(t *testing.T) {
	testCases := []struct {
		name               string
		uid                string
		ProfileContext     string
		expectedError      error
		getClientError     error
		publishError       error
		GetClusterClientFn func() (clusterC.ClientService, error)
		v1ClusterPublishFn func(params *clusterC.V1ClusterProfilesPublishParams) (*models.V1ClusterProfile, error)
	}{
		{
			name:           "Success",
			uid:            "1",
			ProfileContext: "project",
			expectedError:  nil,
			getClientError: nil,
			publishError:   nil,
			v1ClusterPublishFn: func(params *clusterC.V1ClusterProfilesPublishParams) (*models.V1ClusterProfile, error) {
				// Mock implementation of V1ClusterProfilesPublish goes here
				return nil, nil
			},
		},
		{
			name:           "Success",
			uid:            "2",
			ProfileContext: "tenant",
			expectedError:  errors.New("error publishing cluster profile"),
			getClientError: nil,
			publishError:   errors.New("error publishing cluster profile"),
			v1ClusterPublishFn: func(params *clusterC.V1ClusterProfilesPublishParams) (*models.V1ClusterProfile, error) {
				// Mock implementation of V1ClusterProfilesPublish goes here
				return nil, errors.New("error publishing cluster profile")
			},
		},
		{
			name:           "GetClientError",
			uid:            "3",
			ProfileContext: "project",
			expectedError:  errors.New("GetClientError"),
			getClientError: errors.New("GetClientError"),
			v1ClusterPublishFn: func(params *clusterC.V1ClusterProfilesPublishParams) (*models.V1ClusterProfile, error) {
				// Mock implementation of V1ClusterProfilesPublish goes here
				return nil, nil
			},
		},
		{
			name:           "Invalid scope",
			uid:            "4",
			ProfileContext: "invalid",
			expectedError:  errors.New("invalid scope"),
			getClientError: nil,
			v1ClusterPublishFn: func(params *clusterC.V1ClusterProfilesPublishParams) (*models.V1ClusterProfile, error) {
				// Mock implementation of V1ClusterProfilesPublish goes here
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
				V1ClusterProfilesPublishFn: tc.v1ClusterPublishFn,
			}
			err := h.PublishClusterProfile(tc.uid, tc.ProfileContext)
			schema.CompareErrors(t, err, tc.expectedError)
		})
	}
}
