package cluster_profile_test

import (
	"errors"
	"testing"

	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schema"
	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mock"
)

func TestPublishClusterProfile(t *testing.T) {
	testCases := []struct {
		name           string
		uid            string
		ProfileContext string
		expectedError  error
		mock           *mock.ClusterClientMock
	}{
		{
			name:           "Success",
			uid:            "1",
			ProfileContext: "project",
			expectedError:  nil,
			mock: &mock.ClusterClientMock{
				PublishClusterProfileErr: nil,
			},
		},
		{
			name:           "Error",
			uid:            "2",
			ProfileContext: "tenant",
			expectedError:  errors.New("error publishing cluster profile"),
			mock: &mock.ClusterClientMock{
				PublishClusterProfileErr: errors.New("error publishing cluster profile"),
			},
		},
		{
			name:           "Invalid scope",
			uid:            "3",
			ProfileContext: "invalid",
			expectedError:  errors.New("invalid scope"),
			mock:           &mock.ClusterClientMock{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := &client.V1Client{}
			err := h.PublishClusterProfile(tc.mock, tc.uid, tc.ProfileContext)
			schema.CompareErrors(t, err, tc.expectedError)
		})
	}
}
