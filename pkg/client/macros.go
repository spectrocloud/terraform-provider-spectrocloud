package client

import (
	"github.com/spectrocloud/hapi/models"
	"hash/fnv"
	"strconv"

	userC "github.com/spectrocloud/hapi/user/client/v1"
)

func (h *V1Client) CreateMacros(uid string, body *models.V1Macros) error {
	client, err := h.GetUserClient()
	if err != nil {
		return err
	}

	if uid != "" {
		params := userC.NewV1ProjectsUIDMacrosUpdateParams().WithContext(h.Ctx).WithUID(uid).WithBody(body)
		_, err := client.V1ProjectsUIDMacrosUpdate(params)
		if err != nil {
			return err
		}
	} else {
		tenantUID, err := h.GetTenantUID()
		if err != nil {
			return err
		}
		params := userC.NewV1TenantsUIDMacrosUpdateParams().WithContext(h.Ctx).WithTenantUID(tenantUID).WithBody(body)
		_, err = client.V1TenantsUIDMacrosUpdate(params)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *V1Client) GetMacro(name string, projectUID string) (*models.V1Macro, error) {
	macros, err := h.GetMacros(projectUID)
	if err != nil {
		return nil, err
	}

	id := h.StringHash(name)

	for _, macro := range macros {
		if h.StringHash(macro.Name) == id {
			return macro, nil
		}
	}

	return nil, nil
}

func (h *V1Client) GetMacros(projectUID string) ([]*models.V1Macro, error) {
	client, err := h.GetUserClient()
	if err != nil {
		return nil, err
	}

	var macros []*models.V1Macro

	if projectUID != "" {
		params := userC.NewV1ProjectsUIDMacrosListParams().WithContext(h.Ctx).WithUID(projectUID)
		macrosListOk, err := client.V1ProjectsUIDMacrosList(params)
		if err != nil {
			return nil, err
		}
		macros = macrosListOk.Payload.Macros

	} else {
		tenantUID, err := h.GetTenantUID()
		if err != nil || tenantUID == "" {
			return nil, err
		}

		params := userC.NewV1TenantsUIDMacrosListParams().WithTenantUID(tenantUID)
		macrosListOk, err := client.V1TenantsUIDMacrosList(params)
		if err != nil {
			return nil, err
		}
		macros = macrosListOk.Payload.Macros

	}
	return macros, nil
}

func (h *V1Client) StringHash(name string) string {
	return strconv.FormatUint(uint64(hash(name)), 10)
}

func (h *V1Client) UpdateMacros(macros []*models.V1Macro, uid string) error {
	client, err := h.GetUserClient()
	if err != nil {
		return err
	}

	body := &models.V1Macros{
		Macros: macros,
	}

	if uid != "" {
		params := userC.NewV1ProjectsUIDMacrosUpdateParams().WithContext(h.Ctx).WithUID(uid).WithBody(body)
		_, err := client.V1ProjectsUIDMacrosUpdate(params)
		return err

	} else {
		tenantUID, err := h.GetTenantUID()
		if err != nil || tenantUID == "" {
			return err
		}

		params := userC.NewV1TenantsUIDMacrosUpdateParams().WithContext(h.Ctx).WithTenantUID(tenantUID).WithBody(body)
		_, err = client.V1TenantsUIDMacrosUpdate(params)
		return err

	}
}

func (h *V1Client) DeleteMacros(name string, uid string) error {
	macros, err := h.GetMacros(uid)
	if err != nil {
		return err
	}

	keep_macros := make([]*models.V1Macro, 0)

	for _, macro := range macros {
		if macro.Name != name {
			keep_macros = append(keep_macros, macro)
		}
	}

	return h.UpdateMacros(keep_macros, uid)
}

func hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
