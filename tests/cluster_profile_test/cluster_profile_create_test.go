package cluster_profile_test

//func TestCreateClusterProfile(t *testing.T) {
//	testCases := []struct {
//		name                string
//		profileContext      string
//		expectedReturnedUID string
//		expectedError       error
//		clusterProfile      *models.V1ClusterProfileEntity
//		mock                *mock.ClusterClientMock
//	}{
//		{
//			name:                "Success",
//			clusterProfile:      &models.V1ClusterProfileEntity{},
//			profileContext:      "project",
//			expectedError:       nil,
//			expectedReturnedUID: "1",
//			mock: &mock.ClusterClientMock{
//				CreateClusterProfileErr:      nil,
//				CreateClusterProfileResponse: &clusterC.V1ClusterProfilesCreateCreated{Payload: (*models2.V1UID)(&models.V1UID{UID: types.Ptr("1")})},
//			},
//		},
//		{
//			name:                "Success",
//			clusterProfile:      &models.V1ClusterProfileEntity{},
//			profileContext:      "tenant",
//			expectedError:       nil,
//			expectedReturnedUID: "2",
//			mock: &mock.ClusterClientMock{
//				CreateClusterProfileErr:      nil,
//				CreateClusterProfileResponse: &clusterC.V1ClusterProfilesCreateCreated{Payload: (*models2.V1UID)(&models.V1UID{UID: types.Ptr("2")})},
//			},
//		},
//		{
//			name:           "Error",
//			clusterProfile: &models.V1ClusterProfileEntity{},
//			profileContext: "tenant",
//			expectedError:  errors.New("error creating cluster profile"),
//			mock: &mock.ClusterClientMock{
//				CreateClusterProfileErr:      errors.New("error creating cluster profile"),
//				CreateClusterProfileResponse: nil,
//			},
//		},
//		{
//			name:           "Invalid scope",
//			clusterProfile: &models.V1ClusterProfileEntity{},
//			profileContext: "invalid",
//			expectedError:  errors.New("invalid scope"),
//			mock:           &mock.ClusterClientMock{},
//		},
//	}
//	//for _, tc := range testCases {
//	//	t.Run(tc.name, func(t *testing.T) {
//	//		h := &client.V1Client{}
//	//		id, err := h.CreateClusterProfile(tc.clusterProfile)
//	//		if tc.expectedError != nil {
//	//			assert.EqualError(t, err, tc.expectedError.Error())
//	//		} else {
//	//			assert.NoError(t, err)
//	//		}
//	//		if tc.expectedReturnedUID != "" {
//	//			assert.Equal(t, id, tc.expectedReturnedUID)
//	//		}
//	//	})
//	//}
//}
