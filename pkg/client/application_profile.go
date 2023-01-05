package client

import (
	"fmt"
	"github.com/pkg/errors"
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	hashboardC "github.com/spectrocloud/hapi/hashboard/client/v1"
	"github.com/spectrocloud/hapi/models"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client/herr"
)

func (h *V1Client) GetApplicationProfileByName(profileName string) (*models.V1AppProfileSummary, error) {
	client, err := h.GetHashboard()
	if err != nil {
		return nil, err
	}

	limit := int64(0)
	params := hashboardC.NewV1DashboardAppProfilesParamsWithContext(h.Ctx).WithLimit(&limit)
	profiles, err := client.V1DashboardAppProfiles(params)
	if err != nil {
		return nil, err
	}

	for _, profile := range profiles.Payload.AppProfiles {
		if profile.Metadata.Name == profileName {
			return profile, nil
		}
	}

	return nil, fmt.Errorf("Application profile '%s' not found.", profileName)
}

func (h *V1Client) GetApplicationProfile(uid string) (*models.V1AppProfile, error) {
	// Unit test mock handler
	if h.GetApplicationProfileFn != nil {
		return h.GetApplicationProfileFn(uid)
	}
	//
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1AppProfilesUIDGetParamsWithContext(h.Ctx).WithUID(uid)
	response, err := client.V1AppProfilesUIDGet(params)
	if err != nil {
		if herr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return response.Payload, nil
}

func (h *V1Client) GetApplicationProfileTiers(applicationProfileUID string) ([]*models.V1AppTier, error) {
	if h.GetApplicationProfileTiersFn != nil {
		return h.GetApplicationProfileTiersFn(applicationProfileUID)
	}

	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1AppProfilesUIDTiersGetParamsWithContext(h.Ctx).WithUID(applicationProfileUID)
	success, err := client.V1AppProfilesUIDTiersGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload.Spec.AppTiers, nil
}

func (h *V1Client) GetApplicationProfileTierManifestContent(applicationProfileUID string, tierUID string, manifestUID string) (string, error) {
	if h.GetApplicationProfileTierManifestContentFn != nil {
		return h.GetApplicationProfileTierManifestContentFn(applicationProfileUID, tierUID, manifestUID)
	}

	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}
	params := &clusterC.V1AppProfilesUIDTiersUIDManifestsUIDGetParams{
		UID:         applicationProfileUID,
		TierUID:     tierUID,
		ManifestUID: manifestUID,
		Context:     h.Ctx,
	}
	success, err := client.V1AppProfilesUIDTiersUIDManifestsUIDGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return success.Payload.Spec.Published.Content, nil
}

func (h *V1Client) PatchApplicationProfile(appProfileUID string, metadata *models.V1AppProfileMetaEntity, ProfileContext string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	var params *clusterC.V1AppProfilesUIDMetadataUpdateParams
	switch ProfileContext {
	case "project":
		params = clusterC.NewV1AppProfilesUIDMetadataUpdateParamsWithContext(h.Ctx).WithUID(appProfileUID).WithBody(metadata)
		break
	case "tenant":
		params = clusterC.NewV1AppProfilesUIDMetadataUpdateParams().WithUID(appProfileUID).WithBody(metadata)
		break
	default:
		break
	}
	_, err = client.V1AppProfilesUIDMetadataUpdate(params)
	return err
}

func (h *V1Client) CreateApplicationProfileTiers(appProfileUID string, appTiers []*models.V1AppTierEntity, ProfileContext string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	for _, appTier := range appTiers {
		var params *clusterC.V1AppProfilesUIDTiersCreateParams
		switch ProfileContext {
		case "project":
			params = clusterC.NewV1AppProfilesUIDTiersCreateParamsWithContext(h.Ctx).WithUID(appProfileUID).WithBody(appTier)
			break
		case "tenant":
			params = clusterC.NewV1AppProfilesUIDTiersCreateParams().WithUID(appProfileUID).WithBody(appTier)
			break
		default:
			break
		}
		_, tmp_err := client.V1AppProfilesUIDTiersCreate(params)
		if tmp_err != nil {
			err = errors.Wrap(err, tmp_err.Error())
		}
	}
	return err
}

func (h *V1Client) UpdateApplicationProfileTiers(appProfileUID string, tierUID string, appTier *models.V1AppTierUpdateEntity, ProfileContext string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	var params *clusterC.V1AppProfilesUIDTiersUIDUpdateParams
	switch ProfileContext {
	case "project":
		params = clusterC.NewV1AppProfilesUIDTiersUIDUpdateParamsWithContext(h.Ctx).WithUID(appProfileUID).WithTierUID(tierUID).WithBody(appTier)
		break
	case "tenant":
		params = clusterC.NewV1AppProfilesUIDTiersUIDUpdateParams().WithUID(appProfileUID).WithTierUID(tierUID).WithBody(appTier)
		break
	default:
		break
	}

	_, err = client.V1AppProfilesUIDTiersUIDUpdate(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		return nil
	} else if err != nil {
		return err
	}

	return err
}

func (h *V1Client) DeleteApplicationProfileTiers(appProfileUID string, appTiers []string, ProfileContext string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	for _, appTierUID := range appTiers {
		var params *clusterC.V1AppProfilesUIDTiersUIDDeleteParams
		switch ProfileContext {
		case "project":
			params = clusterC.NewV1AppProfilesUIDTiersUIDDeleteParamsWithContext(h.Ctx).WithUID(appProfileUID).WithTierUID(appTierUID)
			break
		case "tenant":
			params = clusterC.NewV1AppProfilesUIDTiersUIDDeleteParams().WithUID(appProfileUID).WithTierUID(appTierUID)
			break
		default:
			break
		}
		_, tmp_err := client.V1AppProfilesUIDTiersUIDDelete(params)
		if tmp_err != nil {
			err = errors.Wrap(err, tmp_err.Error())
		}
	}
	return err
}

func (h *V1Client) CreateApplicationProfile(appProfile *models.V1AppProfileEntity, ProfileContext string) (string, error) {
	// Unit test mock handler
	if h.CreateApplicationProfileFn != nil {
		return h.CreateApplicationProfileFn(appProfile, ProfileContext)
	}
	//

	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	var params *clusterC.V1AppProfilesCreateParams
	switch ProfileContext {
	case "project":
		params = clusterC.NewV1AppProfilesCreateParams().WithContext(h.Ctx).WithBody(appProfile)
		break
	case "tenant":
		params = clusterC.NewV1AppProfilesCreateParams().WithBody(appProfile)
		break
	default:
		break
	}
	success, err := client.V1AppProfilesCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) DeleteApplicationProfile(uid string) error {
	// Unit test mock handler
	if h.DeleteApplicationProfileFn != nil {
		return h.DeleteApplicationProfileFn(uid)
	}
	//
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	profile, err := h.GetApplicationProfile(uid)
	if err != nil {
		return nil
	}

	var params *clusterC.V1AppProfilesUIDDeleteParams
	switch profile.Metadata.Annotations["scope"] {
	case "project":
		params = clusterC.NewV1AppProfilesUIDDeleteParamsWithContext(h.Ctx).WithUID(uid)
		break
	case "tenant":
		params = clusterC.NewV1AppProfilesUIDDeleteParams().WithUID(uid)
		break
	default:
		break
	}

	_, err = client.V1AppProfilesUIDDelete(params)
	return err
}
