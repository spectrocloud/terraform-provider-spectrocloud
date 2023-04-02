package tests

import (
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

type HapiMock struct {
	clusterC.ClientService
	CreateClusterProfileErr      error
	DeleteClusterProfileErr      error
	GetClusterProfilesErr        error
	UpdateClusterProfileErr      error
	PatchClusterProfileErr       error
	PublishClusterProfileErr     error
	CreateClusterProfileResponse *clusterC.V1ClusterProfilesCreateCreated
	GetClusterProfilesResponse   *clusterC.V1ClusterProfilesGetOK
}

func (m *HapiMock) V1ClusterProfilesGet(params *clusterC.V1ClusterProfilesGetParams) (*clusterC.V1ClusterProfilesGetOK, error) {
	return m.GetClusterProfilesResponse, m.GetClusterProfilesErr
}

func (m *HapiMock) V1ClusterProfilesCreate(params *clusterC.V1ClusterProfilesCreateParams) (*clusterC.V1ClusterProfilesCreateCreated, error) {
	return m.CreateClusterProfileResponse, m.CreateClusterProfileErr
}

func (m *HapiMock) V1ClusterProfilesDelete(params *clusterC.V1ClusterProfilesDeleteParams) (*clusterC.V1ClusterProfilesDeleteNoContent, error) {
	return nil, m.DeleteClusterProfileErr
}

func (m *HapiMock) V1ClusterProfilesPublish(params *clusterC.V1ClusterProfilesPublishParams) (*clusterC.V1ClusterProfilesPublishNoContent, error) {
	return nil, m.PublishClusterProfileErr
}

func (m *HapiMock) V1ClusterProfilesUpdate(params *clusterC.V1ClusterProfilesUpdateParams) (*clusterC.V1ClusterProfilesUpdateNoContent, error) {
	return nil, m.UpdateClusterProfileErr
}

func (m *HapiMock) V1ClusterProfilesUIDMetadataUpdate(params *clusterC.V1ClusterProfilesUIDMetadataUpdateParams) (*clusterC.V1ClusterProfilesUIDMetadataUpdateNoContent, error) {
	return nil, m.PatchClusterProfileErr
}
