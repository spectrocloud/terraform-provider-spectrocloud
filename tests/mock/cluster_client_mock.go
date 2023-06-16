package mock

import (
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

type ClusterClientMock struct {
	clusterC.ClientService
	CreateClusterProfileErr      error
	DeleteClusterProfileErr      error
	GetClusterProfilesErr        error
	UpdateClusterProfileErr      error
	PatchClusterProfileErr       error
	PublishClusterProfileErr     error
	PatchSPCProfilesErr          error
	PatchSPCProfilesCount        int
	CreateClusterProfileResponse *clusterC.V1ClusterProfilesCreateCreated
	GetClusterProfilesResponse   *clusterC.V1ClusterProfilesGetOK

	DeleteEcrRegistryErr error

	// Cluster
	DeleteClusterErr      error
	ForceDeleteClusterErr error
}

func (m *ClusterClientMock) V1ClusterProfilesGet(params *clusterC.V1ClusterProfilesGetParams) (*clusterC.V1ClusterProfilesGetOK, error) {
	return m.GetClusterProfilesResponse, m.GetClusterProfilesErr
}

func (m *ClusterClientMock) V1ClusterProfilesCreate(params *clusterC.V1ClusterProfilesCreateParams) (*clusterC.V1ClusterProfilesCreateCreated, error) {
	return m.CreateClusterProfileResponse, m.CreateClusterProfileErr
}

func (m *ClusterClientMock) V1ClusterProfilesDelete(params *clusterC.V1ClusterProfilesDeleteParams) (*clusterC.V1ClusterProfilesDeleteNoContent, error) {
	return nil, m.DeleteClusterProfileErr
}

func (m *ClusterClientMock) V1ClusterProfilesPublish(params *clusterC.V1ClusterProfilesPublishParams) (*clusterC.V1ClusterProfilesPublishNoContent, error) {
	return nil, m.PublishClusterProfileErr
}

func (m *ClusterClientMock) V1ClusterProfilesUpdate(params *clusterC.V1ClusterProfilesUpdateParams) (*clusterC.V1ClusterProfilesUpdateNoContent, error) {
	return nil, m.UpdateClusterProfileErr
}

func (m *ClusterClientMock) V1ClusterProfilesUIDMetadataUpdate(params *clusterC.V1ClusterProfilesUIDMetadataUpdateParams) (*clusterC.V1ClusterProfilesUIDMetadataUpdateNoContent, error) {
	return nil, m.PatchClusterProfileErr
}

func (m *ClusterClientMock) V1SpectroClustersPatchProfiles(params *clusterC.V1SpectroClustersPatchProfilesParams) (*clusterC.V1SpectroClustersPatchProfilesNoContent, error) {
	m.PatchSPCProfilesCount++
	if m.PatchSPCProfilesCount < 3 {
		return nil, m.PatchSPCProfilesErr
	}
	return nil, nil
}

func (m *ClusterClientMock) V1EcrRegistriesUIDDelete(params *clusterC.V1EcrRegistriesUIDDeleteParams) (*clusterC.V1EcrRegistriesUIDDeleteNoContent, error) {
	return nil, m.DeleteEcrRegistryErr
}

func (m *ClusterClientMock) V1SpectroClustersDelete(params *clusterC.V1SpectroClustersDeleteParams) (*clusterC.V1SpectroClustersDeleteNoContent, error) {
	if params.ForceDelete != nil && *params.ForceDelete {
		return nil, m.ForceDeleteClusterErr
	} else {
		return nil, m.DeleteClusterErr
	}
}
