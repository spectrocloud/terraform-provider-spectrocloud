package cluster_profile_test

import (
	"errors"
	"testing"

	"github.com/spectrocloud/hapi/client"
	"github.com/spectrocloud/hapi/models"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schema"
)

func TestDeleteClusterProfileError(t *testing.T) {
	testCases := []struct {
		name            string
		uid             string
		profile         *models.V1ClusterProfile
		expectedError   error
		getClientError  error
		getProfileError error
	}{
		{
			name:            "GetClientError",
			uid:             "1",
			profile:         nil,
			expectedError:   errors.New("GetClientError"),
			getClientError:  errors.New("GetClientError"),
			getProfileError: nil,
		},
		{
			name:            "GetProfileError",
			uid:             "2",
			profile:         nil,
			expectedError:   errors.New("GetProfileError"),
			getClientError:  nil,
			getProfileError: errors.New("GetProfileError"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := &client.V1Client{
				GetClusterClientFn: func() (clusterC.ClientService, error) {
					return &clusterC.Client{}, tc.getClientError
				},
				GetClusterProfileFn: func(uid string) (*models.V1ClusterProfile, error) {
					return tc.profile, tc.getProfileError
				},
			}

			err := h.DeleteClusterProfile(tc.uid)
			schema.CompareErrors(t, err, tc.expectedError)

		})

	}
}

func TestDeleteClusterProfile(t *testing.T) {
	testCases := []struct {
		name            string
		uid             string
		profile         *models.V1ClusterProfile
		expectedError   error
		getProfileError error
	}{
		{
			name:            "Success",
			uid:             "1",
			profile:         &models.V1ClusterProfile{Metadata: &models.V1ObjectMeta{Annotations: map[string]string{"scope": "project"}}},
			expectedError:   nil,
			getProfileError: nil,
		},
		{
			name:            "Success",
			uid:             "2",
			profile:         &models.V1ClusterProfile{Metadata: &models.V1ObjectMeta{Annotations: map[string]string{"scope": "tenant"}}},
			expectedError:   nil,
			getProfileError: nil,
		},
		{
			name:            "Invalid scope",
			uid:             "3",
			profile:         &models.V1ClusterProfile{Metadata: &models.V1ObjectMeta{Annotations: map[string]string{"scope": "invalid"}}},
			expectedError:   errors.New("invalid scope"),
			getProfileError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := &client.V1Client{
				GetClusterClientFn: func() (clusterC.ClientService, error) {
					return &clusterC.Client{}, nil
				},
				GetClusterProfileFn: func(uid string) (*models.V1ClusterProfile, error) {
					return tc.profile, tc.getProfileError
				},
				V1ClusterProfilesDeleteFn: func(params *clusterC.V1ClusterProfilesDeleteParams) (*clusterC.V1ClusterProfilesDeleteNoContent, error) {
					return &clusterC.V1ClusterProfilesDeleteNoContent{}, nil
				},
			}
			err := h.DeleteClusterProfile(tc.uid)
			schema.CompareErrors(t, err, tc.expectedError)
		})
	}
}
