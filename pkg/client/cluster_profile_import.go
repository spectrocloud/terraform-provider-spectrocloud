package client

import (
	"github.com/go-openapi/runtime"
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func NewTrue() *bool {
	b := true
	return &b
}

func (h *V1Client) CreateClusterProfileImport(importFile runtime.NamedReadCloser, ProfileContext string) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	var params *clusterC.V1ClusterProfilesImportFileParams
	switch ProfileContext {
	case "project":
		params = clusterC.NewV1ClusterProfilesImportFileParamsWithContext(h.Ctx)
		break
	case "tenant":
		params = clusterC.NewV1ClusterProfilesImportFileParams()
		break
	default:
		break
	}

	params = params.WithPublish(NewTrue()).WithImportFile(importFile)
	success, err := client.V1ClusterProfilesImportFile(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) ClusterProfileExport(uid string) (*models.V1ClusterProfile, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	// no need to switch request context here as /v1/clusterprofiles/{uid} works for profile in any scope.
	params := clusterC.NewV1ClusterProfilesGetParamsWithContext(h.Ctx).WithUID(uid)
	success, err := client.V1ClusterProfilesGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {

		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}
