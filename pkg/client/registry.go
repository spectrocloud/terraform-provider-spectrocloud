package client

import (
	"fmt"
	"github.com/spectrocloud/hapi/models"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func (h *V1Client) GetPackRegistryCommonByName(registryName string) (*models.V1RegistryMetadata, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1RegistriesMetadataParams().WithScope(types.Ptr(""))
	registries, err := client.V1RegistriesMetadata(params)
	if err != nil {
		return nil, err
	}

	for _, registry := range registries.Payload.Items {
		if registry.Name == registryName {
			return registry, nil
		}
	}

	return nil, fmt.Errorf("Registry '%s' not found.", registryName)
}

func (h *V1Client) GetPackRegistryByName(registryName string) (*models.V1PackRegistry, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1RegistriesPackListParams()
	registries, err := client.V1RegistriesPackList(params)
	if err != nil {
		return nil, err
	}

	for _, registry := range registries.Payload.Items {
		if registry.Metadata.Name == registryName {
			return registry, nil
		}
	}

	return nil, fmt.Errorf("Registry '%s' not found.", registryName)
}

func (h *V1Client) GetHelmRegistryByName(registryName string) (*models.V1HelmRegistry, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1RegistriesHelmListParams()
	registries, err := client.V1RegistriesHelmList(params)
	if err != nil {
		return nil, err
	}

	for _, registry := range registries.Payload.Items {
		if registry.Metadata.Name == registryName {
			return registry, nil
		}
	}

	return nil, fmt.Errorf("Registry '%s' not found.", registryName)
}

func (h *V1Client) GetHelmRegistry(uid string) (*models.V1HelmRegistry, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1RegistriesHelmUIDGetParams().WithUID(uid)
	response, err := client.V1RegistriesHelmUIDGet(params)
	if err != nil {
		return nil, err
	}

	return response.Payload, nil
}

func (h *V1Client) CreateHelmRegistry(registry *models.V1HelmRegistryEntity) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1RegistriesHelmCreateParams().WithBody(registry)
	if resp, err := client.V1RegistriesHelmCreate(params); err != nil {
		return "", err
	} else {
		return *resp.Payload.UID, nil
	}
}

func (h *V1Client) UpdateHelmRegistry(uid string, registry *models.V1HelmRegistry) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1RegistriesHelmUIDUpdateParams().WithBody(registry).WithUID(uid)
	if _, err := client.V1RegistriesHelmUIDUpdate(params); err != nil {
		return err
	} else {
		return nil
	}
}

func (h *V1Client) DeleteHelmRegistry(uid string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1RegistriesHelmUIDDeleteParams().WithUID(uid)
	if _, err := client.V1RegistriesHelmUIDDelete(params); err != nil {
		return err
	} else {
		return nil
	}
}

func (h *V1Client) GetOciRegistryByName(registryName string) (*models.V1OciRegistry, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1OciRegistriesSummaryParams()
	registries, err := client.V1OciRegistriesSummary(params)
	if err != nil {
		return nil, err
	}

	for _, registry := range registries.Payload.Items {
		if registry.Metadata.Name == registryName {
			return registry, nil
		}
	}

	return nil, fmt.Errorf("Registry '%s' not found.", registryName)
}

func (h *V1Client) GetOciRegistry(uid string) (*models.V1EcrRegistry, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1EcrRegistriesUIDGetParams().WithUID(uid)
	response, err := client.V1EcrRegistriesUIDGet(params)
	if err != nil {
		return nil, err
	}

	return response.Payload, nil
}

func (h *V1Client) CreateOciEcrRegistry(registry *models.V1EcrRegistry) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1EcrRegistriesCreateParams().WithBody(registry)
	if resp, err := client.V1EcrRegistriesCreate(params); err != nil {
		return "", err
	} else {
		return *resp.Payload.UID, nil
	}
}

func (h *V1Client) UpdateEcrRegistry(uid string, registry *models.V1EcrRegistry) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1EcrRegistriesUIDUpdateParams().WithBody(registry).WithUID(uid)
	_, err = client.V1EcrRegistriesUIDUpdate(params)
	if err != nil {
		return err
	}

	return nil
}

func (h *V1Client) DeleteRegistry(uid string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1BasicOciRegistriesUIDDeleteParams().WithUID(uid)
	_, err = client.V1BasicOciRegistriesUIDDelete(params)
	if err != nil {
		return err
	}

	return nil
}
