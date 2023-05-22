package client

import (
	"errors"
	"fmt"
	"github.com/spectrocloud/hapi/models"
	userC "github.com/spectrocloud/hapi/user/client/v1"
	"hash/fnv"
	"strconv"
	"strings"
	"time"
)

func (h *V1Client) CreateMacros(uid string, macros *models.V1Macros) error {
	client, err := h.GetUserClient()
	if err != nil {
		return err
	}
	if uid != "" {
		params := userC.NewV1ProjectsUIDMacrosCreateParams().WithContext(h.Ctx).WithUID(uid).WithBody(macros)
		_, err = client.V1ProjectsUIDMacrosCreate(params)
	} else {
		var tenantUID string
		tenantUID, err = h.GetTenantUID()
		if err != nil {
			return err
		}
		params := userC.NewV1TenantsUIDMacrosCreateParams().WithContext(h.Ctx).WithTenantUID(tenantUID).WithBody(macros)
		_, err = client.V1TenantsUIDMacrosCreate(params)
	}
	if err != nil {
		err = h.handleMacroDuplicateForbiddenError(macros, uid, err)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *V1Client) handleMacroDuplicateForbiddenError(macros *models.V1Macros, projectUID string, srcErr error) error {
	if strings.Contains(srcErr.Error(), "Code:DuplicateMacroNamesForbidden") {
		var srcMacro *models.V1Macro
		srcMacro = macros.Macros[0]
		actualMacro, err := h.retryGetMacro(3, 10*time.Second, srcMacro.Name, projectUID)
		if err != nil {
			return err
		}
		if actualMacro != nil {
			if (srcMacro.Name == actualMacro.Name) && (srcMacro.Value == actualMacro.Value) {
				return nil
			} else if (srcMacro.Name == actualMacro.Name) && (srcMacro.Value != actualMacro.Value) {
				return srcErr
			}
		} else {
			return errors.New("code: Failed to create macro")
		}
	} else {
		return srcErr
	}
	return nil
}
func (h *V1Client) retryGetMacro(attempts int, sleep time.Duration, macroName string, projectUid string) (*models.V1Macro, error) {
	var macro *models.V1Macro
	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(sleep)
		}
		macro, _ = h.GetMacro(macroName, projectUid)
		if macro != nil {
			return macro, nil
		}
	}
	return nil, fmt.Errorf("after %d attempts, last error: %s", attempts, errors.New("macro get error"))
}

func (h *V1Client) handleMacroNotFoundError(err error) error {
	if strings.Contains(err.Error(), "Code:MacrosNotFound") || strings.Contains(err.Error(), "Code:ResourceNotFound") {
		return nil
	} else {
		return err
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
		_, err = client.V1ProjectsUIDMacrosDeleteByMacroName(params)
		if err != nil {
			err = h.handleMacroNotFoundError(err)
		}
	} else {
		var tenantUID string
		tenantUID, err = h.GetTenantUID()
		if err != nil {
			return err
		}
		params := userC.NewV1TenantsUIDMacrosDeleteByMacroNameParams().WithContext(h.Ctx).WithTenantUID(tenantUID).WithBody(body)
		_, err = client.V1TenantsUIDMacrosDeleteByMacroName(params)
		if err != nil {
			err = h.handleMacroNotFoundError(err)
		}
	}
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
