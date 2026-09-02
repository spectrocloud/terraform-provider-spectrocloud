package routes

import (
	"net/http"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// mockKubevirtVMPayload returns a minimal V1ClusterVirtualMachine that
// exercises the flatten path (metadata + one data-volume template) with
// a Status whose PrintableStatus is "Deleted" — the target state used by
// resourceKubevirtVirtualMachineDelete's waitForVirtualMachineToTargetState
// call. Reporting the target on the first refresh lets the waiter exit
// immediately and keeps unit tests fast.
func mockKubevirtVMPayload() *models.V1ClusterVirtualMachine {
	dv := "boot-vol"
	return &models.V1ClusterVirtualMachine{
		APIVersion: "kubevirt.io/v1",
		Kind:       "VirtualMachine",
		Metadata: &models.V1VMObjectMeta{
			Name:      "test-vm",
			Namespace: "default",
			UID:       "test-vm-uid",
		},
		Spec: &models.V1ClusterVirtualMachineSpec{
			RunStrategy: "Always",
			DataVolumeTemplates: []*models.V1VMDataVolumeTemplateSpec{
				{
					Metadata: &models.V1VMObjectMeta{
						Name:      dv,
						Namespace: "default",
					},
					Spec: &models.V1VMDataVolumeSpec{},
				},
			},
		},
		Status: &models.V1ClusterVirtualMachineStatus{
			PrintableStatus: "Deleted",
			Ready:           false,
			Created:         true,
		},
	}
}

// KubevirtVMRoutes exposes the /v1/spectroclusters/{uid}/vms/* endpoints
// consumed by resource_kubevirt_virtual_machine + resource_kubevirt_datavolume.
// The routes here cover: Create, Get, Update, Delete on the VM plus
// AddVolume / RemoveVolume for the datavolume resource. All responses
// echo the same fixture so read-after-write asserts round-trip cleanly.
func KubevirtVMRoutes() []Route {
	vm := mockKubevirtVMPayload()
	return []Route{
		{
			// SDK client reads a 200 OK envelope, not 201 Created.
			Method: "POST",
			Path:   "/v1/spectroclusters/{uid}/vms",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    vm,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/{uid}/vms/{vmName}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    vm,
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/spectroclusters/{uid}/vms/{vmName}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    vm,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/spectroclusters/{uid}/vms/{vmName}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			// AddVolume — SDK expects 204 No Content.
			Method: "PUT",
			Path:   "/v1/spectroclusters/{uid}/vms/{vmName}/addVolume",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			// RemoveVolume — SDK expects 204 No Content.
			Method: "PUT",
			Path:   "/v1/spectroclusters/{uid}/vms/{vmName}/removeVolume",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			// VM list endpoint — used by GetVirtualMachines. Return the
			// same VM in a one-element list.
			Method: "GET",
			Path:   "/v1/spectroclusters/{uid}/vms",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1ClusterVirtualMachineList{
					Items: []*models.V1ClusterVirtualMachine{vm},
				},
			},
		},
	}
}
