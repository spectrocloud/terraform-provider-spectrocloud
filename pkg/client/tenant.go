package client

import (
	userC "github.com/spectrocloud/hapi/user/client/v1"
)

func (h *V1Client) GetTenantUID() (string, error) {
	client, err := h.GetUserClient()
	if err != nil {
		return "", err
	}

	params := userC.NewV1UsersMeGetParams()
	me, err := client.V1UsersMeGet(params)
	if err != nil || me == nil {
		return "", err
	}
	return me.Payload.Status.Tenant.TenantUID, nil

}
