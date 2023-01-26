package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

// Cloud Account

func (h *V1Client) CreateCloudAccountAws(account *models.V1AwsAccount, AccountContext string) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	var params *clusterC.V1CloudAccountsAwsCreateParams
	switch AccountContext {
	case "project":
		params = clusterC.NewV1CloudAccountsAwsCreateParamsWithContext(h.Ctx).WithBody(account)
		break
	case "tenant":
		params = clusterC.NewV1CloudAccountsAwsCreateParams().WithBody(account)
		break
	default:
		break
	}
	success, err := client.V1CloudAccountsAwsCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) UpdateCloudAccountAws(account *models.V1AwsAccount) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	uid := account.Metadata.UID
	var params *clusterC.V1CloudAccountsAwsUpdateParams
	switch account.Metadata.Annotations["scope"] {
	case "project":
		params = clusterC.NewV1CloudAccountsAwsUpdateParamsWithContext(h.Ctx).WithUID(uid).WithBody(account)
		break
	case "tenant":
		params = clusterC.NewV1CloudAccountsAwsUpdateParams().WithBody(account)
		break
	default:
		break
	}
	_, err = client.V1CloudAccountsAwsUpdate(params)
	return err
}

func (h *V1Client) DeleteCloudAccountAws(uid string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	account, err := h.GetCloudAccountAws(uid)
	if err != nil {
		return err
	}

	var params *clusterC.V1CloudAccountsAwsDeleteParams
	switch account.Metadata.Annotations["scope"] {
	case "project":
		params = clusterC.NewV1CloudAccountsAwsDeleteParamsWithContext(h.Ctx).WithUID(uid)
		break
	case "tenant":
		params = clusterC.NewV1CloudAccountsAwsDeleteParams().WithUID(uid)
		break
	default:
		break
	}
	_, err = client.V1CloudAccountsAwsDelete(params)
	return err
}

func (h *V1Client) GetCloudAccountAws(uid string) (*models.V1AwsAccount, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudAccountsAwsGetParamsWithContext(h.Ctx).WithUID(uid)
	success, err := client.V1CloudAccountsAwsGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) GetCloudAccountsAws() ([]*models.V1AwsAccount, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudAccountsAwsListParamsWithContext(h.Ctx)
	response, err := client.V1CloudAccountsAwsList(params)
	if err != nil {
		return nil, err
	}

	accounts := make([]*models.V1AwsAccount, len(response.Payload.Items))
	for i, account := range response.Payload.Items {
		accounts[i] = account
	}

	return accounts, nil
}
