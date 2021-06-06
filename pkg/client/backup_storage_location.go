package client

import (
	"github.com/spectrocloud/hapi/models"
	userC "github.com/spectrocloud/hapi/user/client/v1alpha1"
)

func (h *V1alpha1Client) ListBackupStorageLocation(projectScope bool) ([]*models.V1alpha1UserAssetsLocation, error) {
	client, err := h.getUserClient()
	if err != nil {
		return nil, err
	}

	params := userC.NewV1alpha1UsersAssetsLocationGetParams()
	if projectScope {
		params.WithContext(h.ctx)
	}
	response, err := client.V1alpha1UsersAssetsLocationGet(params)
	if err != nil {
		return nil, err
	}

	bsls := make([]*models.V1alpha1UserAssetsLocation, len(response.Payload.Items))
	for i, account := range response.Payload.Items {
		bsls[i] = account
	}

	return bsls, nil
}

func (h *V1alpha1Client) GetBackupStorageLocation(uid string) (*models.V1alpha1UserAssetsLocation, error) {
	client, err := h.getUserClient()
	if err != nil {
		return nil, err
	}

	params := userC.NewV1alpha1UsersAssetsLocationGetParamsWithContext(h.ctx)
	response, err := client.V1alpha1UsersAssetsLocationGet(params)
	if err != nil {
		return nil, err
	}

	for _, account := range response.Payload.Items {
		if account.Metadata.UID == uid {
			return account, nil
		}
	}

	return nil, nil
}

func (h *V1alpha1Client) GetS3BackupStorageLocation(uid string) (*models.V1alpha1UserAssetsLocationS3, error) {
	client, err := h.getUserClient()
	if err != nil {
		return nil, err
	}

	params := userC.NewV1alpha1UsersAssetsLocationS3GetParamsWithContext(h.ctx).WithUID(uid)
	if response, err := client.V1alpha1UsersAssetsLocationS3Get(params); err != nil {
		return nil, err
	} else {
		return response.Payload, nil
	}
}

func (h *V1alpha1Client) CreateS3BackupStorageLocation(bsl *models.V1alpha1UserAssetsLocationS3) (string, error) {
	client, err := h.getUserClient()
	if err != nil {
		return "", err
	}

	params := userC.NewV1alpha1UsersAssetsLocationS3CreateParamsWithContext(h.ctx).WithBody(bsl)
	if resp, err := client.V1alpha1UsersAssetsLocationS3Create(params); err != nil {
		return "", err
	} else {
		return *resp.Payload.UID, nil
	}
}

func (h *V1alpha1Client) UpdateS3BackupStorageLocation(uid string, bsl *models.V1alpha1UserAssetsLocationS3) error {
	client, err := h.getUserClient()
	if err != nil {
		return err
	}

	params := userC.NewV1alpha1UsersAssetsLocationS3UpdateParamsWithContext(h.ctx).WithUID(uid).WithBody(bsl)
	if _, err := client.V1alpha1UsersAssetsLocationS3Update(params); err != nil {
		return err
	}
	return nil
}

func (h *V1alpha1Client) DeleteS3BackupStorageLocation(uid string) error {
	client, err := h.getUserClient()
	if err != nil {
		return err
	}

	params := userC.NewV1alpha1UsersAssetsLocationS3DeleteParamsWithContext(h.ctx).WithUID(uid)
	if _, err := client.V1alpha1UsersAssetsLocationS3Delete(params); err != nil {
		return err
	}
	return nil
}
