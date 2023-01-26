package client

import (
	"errors"
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

const OverlordUID = "overlordUid"

// convert V1VsphereAccount to V1OverlordVsphereAccountEntity
func toV1OverlordsUIDVsphereAccountValidateBody(account *models.V1VsphereAccount) clusterC.V1OverlordsUIDVsphereAccountValidateBody {
	return clusterC.V1OverlordsUIDVsphereAccountValidateBody{
		Account: account.Spec,
	}
}

// Cloud Account
func (h *V1Client) CreateCloudAccountVsphere(account *models.V1VsphereAccount, AccountContext string) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	// validate account
	err = validateAccount(account, h)
	if err != nil {
		return "", err
	}

	// create account
	var params *clusterC.V1CloudAccountsVsphereCreateParams
	switch AccountContext {
	case "project":
		params = clusterC.NewV1CloudAccountsVsphereCreateParamsWithContext(h.Ctx)
		break
	case "tenant":
		params = clusterC.NewV1CloudAccountsVsphereCreateParams()
		break
	default:
		return "", errors.New("invalid account context")
	}
	params = params.WithBody(account)
	success, err := client.V1CloudAccountsVsphereCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func validateAccount(account *models.V1VsphereAccount, h *V1Client) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	paramsValidate := clusterC.NewV1OverlordsUIDVsphereAccountValidateParams().WithUID(account.Metadata.Annotations[OverlordUID])
	paramsValidate = paramsValidate.WithBody(toV1OverlordsUIDVsphereAccountValidateBody(account))
	_, err = client.V1OverlordsUIDVsphereAccountValidate(paramsValidate)
	if err != nil {
		return err
	}

	return nil
}

func (h *V1Client) UpdateCloudAccountVsphere(account *models.V1VsphereAccount, AccountContext string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	// validate account
	err = validateAccount(account, h)
	if err != nil {
		return err
	}

	uid := account.Metadata.UID
	var params *clusterC.V1CloudAccountsVsphereUpdateParams
	switch AccountContext {
	case "project":
		params = clusterC.NewV1CloudAccountsVsphereUpdateParamsWithContext(h.Ctx).WithUID(uid)
		break
	case "tenant":
		params = clusterC.NewV1CloudAccountsVsphereUpdateParams().WithUID(uid)
		break
	default:
		return errors.New("invalid account context")
	}
	params = params.WithBody(account)
	_, err = client.V1CloudAccountsVsphereUpdate(params)
	return err
}

func (h *V1Client) DeleteCloudAccountVsphere(uid string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudAccountsVsphereDeleteParamsWithContext(h.Ctx).WithUID(uid)
	_, err = client.V1CloudAccountsVsphereDelete(params)
	return err
}

func (h *V1Client) GetCloudAccountVsphere(uid string) (*models.V1VsphereAccount, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	getParams := clusterC.NewV1CloudAccountsVsphereGetParamsWithContext(h.Ctx).WithUID(uid)
	success, err := client.V1CloudAccountsVsphereGet(getParams)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) GetCloudAccountsVsphere() ([]*models.V1VsphereAccount, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudAccountsVsphereListParamsWithContext(h.Ctx)
	response, err := client.V1CloudAccountsVsphereList(params)
	if err != nil {
		return nil, err
	}

	accounts := make([]*models.V1VsphereAccount, len(response.Payload.Items))
	for i, account := range response.Payload.Items {
		accounts[i] = account
	}

	return accounts, nil
}
