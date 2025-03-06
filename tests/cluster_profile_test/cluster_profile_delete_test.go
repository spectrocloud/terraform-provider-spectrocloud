package cluster_profile_test

import (
	"errors"
	"testing"

	clusterC "github.com/spectrocloud/palette-sdk-go/api/client/v1"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mock"
)

func TestDeleteClusterProfileError(t *testing.T) {
	testCases := []struct {
		name          string
		uid           string
		profile       *models.V1ClusterProfile
		expectedError error
		mock          *mock.ClusterClientMock
	}{
		{
			name:          "GetProfileError",
			uid:           "1",
			expectedError: errors.New("GetProfileError"),
			mock: &mock.ClusterClientMock{
				GetClusterProfilesResponse: nil,
				GetClusterProfilesErr:      errors.New("GetProfileError"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			//hapiClient := &client.V1Client{}
			//err := hapiClient.DeleteClusterProfile(tc.uid)
			//schema.CompareErrors(t, err, tc.expectedError)
		})
	}
}

func TestDeleteClusterProfile(t *testing.T) {
	testCases := []struct {
		name          string
		uid           string
		profile       *models.V1ClusterProfile
		expectedError error
		mock          *mock.ClusterClientMock
	}{
		{
			name:          "Success",
			uid:           "1",
			expectedError: nil,
			mock: &mock.ClusterClientMock{
				GetClusterProfilesResponse: &clusterC.V1ClusterProfilesGetOK{
					//Payload: &models.V1ClusterProfile{Metadata: &models.V1ObjectMeta{Annotations: map[string]string{"scope": "project"}}},
				},
				GetClusterProfilesErr: nil,
			},
		},
		{
			name:          "Success",
			uid:           "2",
			expectedError: nil,
			mock: &mock.ClusterClientMock{
				GetClusterProfilesResponse: &clusterC.V1ClusterProfilesGetOK{
					//Payload: &models.V1ClusterProfile{Metadata: &models.V1ObjectMeta{Annotations: map[string]string{"scope": "tenant"}}},
				},
				GetClusterProfilesErr: nil,
			},
		},
		{
			name:          "Invalid scope",
			uid:           "3",
			expectedError: errors.New("invalid scope"),
			mock: &mock.ClusterClientMock{
				GetClusterProfilesResponse: &clusterC.V1ClusterProfilesGetOK{
					//Payload: &models.V1ClusterProfile{Metadata: &models.V1ObjectMeta{Annotations: map[string]string{"scope": "invalid"}}},
				},
				GetClusterProfilesErr: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			//h := &client.V1Client{}
			//err := h.DeleteClusterProfile(tc.uid)
			//schema.CompareErrors(t, err, tc.expectedError)
		})
	}
}
