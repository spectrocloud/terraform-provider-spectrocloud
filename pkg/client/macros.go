package client

import (
	"hash/fnv"
	"strconv"

	"github.com/spectrocloud/hapi/models"
	userC "github.com/spectrocloud/hapi/user/client/v1"
)

func (h *V1Client) CreateMacros(uid string, macros *models.V1Macros) error {
	client, err := h.GetUserClient()
	if err != nil {
		return err
	}
	if uid != "" {
		params := userC.NewV1ProjectsUIDMacrosCreateParams().WithContext(h.Ctx).WithUID(uid).WithBody(macros)
		_, err := client.V1ProjectsUIDMacrosCreate(params)
		if err != nil {
			return err
		}
	} else {
		tenantUID, err := h.GetTenantUID()
		if err != nil {
			return err
		}
		params := userC.NewV1TenantsUIDMacrosCreateParams().WithContext(h.Ctx).WithTenantUID(tenantUID).WithBody(macros)
		_, err = client.V1TenantsUIDMacrosCreate(params)
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
	id := h.GetMacroId(projectUID, name)

	for _, macro := range macros {
		if h.GetMacroId(projectUID, macro.Name) == id {
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

func (h *V1Client) UpdateMacros(uid string, macros *models.V1Macros) error {
	client, err := h.GetUserClient()
	if err != nil {
		return err
	}
	if uid != "" {
		params := userC.NewV1ProjectsUIDMacrosUpdateByMacroNameParams().WithContext(h.Ctx).WithUID(uid).WithBody(macros)
		_, err := client.V1ProjectsUIDMacrosUpdateByMacroName(params)
		return err

	} else {
		tenantUID, err := h.GetTenantUID()
		if err != nil || tenantUID == "" {
			return err
		}
		params := userC.NewV1TenantsUIDMacrosUpdateByMacroNameParams().WithContext(h.Ctx).WithTenantUID(tenantUID).WithBody(macros)
		_, err = client.V1TenantsUIDMacrosUpdateByMacroName(params)
		return err
	}
}

func (h *V1Client) DeleteMacros(uid string, body *models.V1Macros) error {

	client, err := h.GetUserClient()
	if err != nil {
		return err
	}

	if uid != "" {
		params := userC.NewV1ProjectsUIDMacrosDeleteByMacroNameParams().WithContext(h.Ctx).WithUID(uid).WithBody(body)
		_, err := client.V1ProjectsUIDMacrosDeleteByMacroName(params)
		if err != nil {
			return err
		}
	} else {
		tenantUID, err := h.GetTenantUID()
		if err != nil {
			return err
		}
		params := userC.NewV1TenantsUIDMacrosDeleteByMacroNameParams().WithContext(h.Ctx).WithTenantUID(tenantUID).WithBody(body)
		_, err = client.V1TenantsUIDMacrosDeleteByMacroName(params)
		if err != nil {
			return err
		}
	}
	_, err = h.GetMacros(uid)
	if err != nil {
		return err
	}
	return nil
}

func hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

func (h *V1Client) GetMacroId(uid string, name string) string {
	var hash string
	if uid != "" {
		hash = h.StringHash(name + uid)
	} else {
		hash = h.StringHash(name + "%tenant")
	}
	return hash
}
