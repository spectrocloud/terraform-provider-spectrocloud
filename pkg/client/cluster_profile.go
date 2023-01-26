package client

import (
	"errors"
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	hashboardC "github.com/spectrocloud/hapi/hashboard/client/v1"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) DeleteClusterProfile(uid string) error {
	if h.DeleteClusterProfileFn != nil {
		return h.DeleteClusterProfileFn(uid)
	}
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	profile, err := h.GetClusterProfile(uid)
	if err != nil || profile == nil {
		return err
	}

	var params *clusterC.V1ClusterProfilesDeleteParams
	switch profile.Metadata.Annotations["scope"] {
	case "project":
		params = clusterC.NewV1ClusterProfilesDeleteParamsWithContext(h.Ctx).WithUID(uid)
		break
	case "tenant":
		params = clusterC.NewV1ClusterProfilesDeleteParams().WithUID(uid)
		break
	default:
		return errors.New("invalid scope")
	}

	if h.V1ClusterProfilesDeleteFn != nil {
		_, err = h.V1ClusterProfilesDeleteFn(params)
	} else {
		_, err = client.V1ClusterProfilesDelete(params)
	}
	return err
}

func (h *V1Client) GetClusterProfile(uid string) (*models.V1ClusterProfile, error) {
	if h.GetClusterProfileFn != nil {
		return h.GetClusterProfileFn(uid)
	}
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	// no need to switch request context here as /v1/clusterprofiles/{uid} works for profile in any scope.
	params := clusterC.NewV1ClusterProfilesGetParamsWithContext(h.Ctx).WithUID(uid)
	success, err := client.V1ClusterProfilesGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) GetClusterProfiles() ([]*models.V1ClusterProfileMetadata, error) {
	client, err := h.GetHashboard()
	if err != nil {
		return nil, err
	}

	params := hashboardC.NewV1ClusterProfilesMetadataParamsWithContext(h.Ctx)
	response, err := client.V1ClusterProfilesMetadata(params)
	if err != nil {
		return nil, err
	}

	return response.Payload.Items, nil
}

func (h *V1Client) PatchClusterProfile(clusterProfile *models.V1ClusterProfileUpdateEntity, metadata *models.V1ProfileMetaEntity, ProfileContext string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	uid := clusterProfile.Metadata.UID
	var params *clusterC.V1ClusterProfilesUIDMetadataUpdateParams
	switch ProfileContext {
	case "project":
		params = clusterC.NewV1ClusterProfilesUIDMetadataUpdateParamsWithContext(h.Ctx).WithUID(uid).WithBody(metadata)
		break
	case "tenant":
		params = clusterC.NewV1ClusterProfilesUIDMetadataUpdateParams().WithUID(uid).WithBody(metadata)
		break
	default:
		return errors.New("invalid scope")
	}
	if h.V1ClusterProfilesUIDMetadataUpdateFn != nil {
		_, err = h.V1ClusterProfilesUIDMetadataUpdateFn(params)
	} else {
		_, err = client.V1ClusterProfilesUIDMetadataUpdate(params)
	}
	return err
}

func (h *V1Client) UpdateClusterProfile(clusterProfile *models.V1ClusterProfileUpdateEntity, ProfileContext string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	uid := clusterProfile.Metadata.UID
	var params *clusterC.V1ClusterProfilesUpdateParams
	switch ProfileContext {
	case "project":
		params = clusterC.NewV1ClusterProfilesUpdateParamsWithContext(h.Ctx).WithUID(uid).WithBody(clusterProfile)
		break
	case "tenant":
		params = clusterC.NewV1ClusterProfilesUpdateParams().WithUID(uid).WithBody(clusterProfile)
		break
	default:
		return errors.New("invalid scope")
	}

	if h.V1ClusterProfilesUpdateFn != nil {
		_, err = h.V1ClusterProfilesUpdateFn(params)
	} else {
		_, err = client.V1ClusterProfilesUpdate(params)
	}
	return err
}

func (h *V1Client) CreateClusterProfile(clusterProfile *models.V1ClusterProfileEntity, ProfileContext string) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	var params *clusterC.V1ClusterProfilesCreateParams
	switch ProfileContext {
	case "project":
		params = clusterC.NewV1ClusterProfilesCreateParamsWithContext(h.Ctx).WithBody(clusterProfile)
		break
	case "tenant":
		params = clusterC.NewV1ClusterProfilesCreateParams().WithBody(clusterProfile)
		break
	default:
		return "", errors.New("invalid scope")
	}

	var success *clusterC.V1ClusterProfilesCreateCreated
	if h.V1ClusterProfilesCreateFn != nil {
		success, err = h.V1ClusterProfilesCreateFn(params)
	} else {
		success, err = client.V1ClusterProfilesCreate(params)
	}
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) PublishClusterProfile(uid string, ProfileContext string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	var params *clusterC.V1ClusterProfilesPublishParams
	switch ProfileContext {
	case "project":
		params = clusterC.NewV1ClusterProfilesPublishParamsWithContext(h.Ctx).WithUID(uid)
		break
	case "tenant":
		params = clusterC.NewV1ClusterProfilesPublishParams().WithUID(uid)
		break
	default:
		return errors.New("invalid scope")
	}
	if h.V1ClusterProfilesPublishFn != nil {
		_, err = h.V1ClusterProfilesPublishFn(params)
	} else {
		_, err = client.V1ClusterProfilesPublish(params)
	}

	return err
}
