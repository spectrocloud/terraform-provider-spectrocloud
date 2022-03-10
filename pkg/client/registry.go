package client

import (
	"fmt"
	"github.com/spectrocloud/hapi/models"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) GetRegistryOciByName(registryName string) (*models.V1OciRegistry, error) {
	client, err := h.getClusterClient()
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

func (h *V1Client) GetRegistryOci(uid string) (*models.V1EcrRegistry, error) {
	client, err := h.getClusterClient()
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
	client, err := h.getClusterClient()
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
	client, err := h.getClusterClient()
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
	client, err := h.getClusterClient()
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
