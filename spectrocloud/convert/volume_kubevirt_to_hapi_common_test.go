package convert

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func TestToHapiVolume(t *testing.T) {
	// Step 1: Prepare sample DataVolume object
	volume := &cdiv1.DataVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:                       "test-volume",
			Namespace:                  "default",
			Annotations:                map[string]string{"test-annotation-key": "test-annotation-value"},
			DeletionGracePeriodSeconds: int64Ptr(30),
			Finalizers:                 []string{"test-finalizer"},
			GenerateName:               "test-",
			Generation:                 1,
			Labels:                     map[string]string{"test-label-key": "test-label-value"},
			ManagedFields:              []metav1.ManagedFieldsEntry{},
			OwnerReferences:            []metav1.OwnerReference{},
			ResourceVersion:            "123456",
			UID:                        "test-uid",
		},
		Spec: cdiv1.DataVolumeSpec{
			// Populate DataVolumeSpec fields
		},
	}

	// Step 2: Prepare AddVolumeOptions
	addVolumeOptions := &models.V1VMAddVolumeOptions{
		// Populate AddVolumeOptions fields
	}

	// Step 3: Call the function
	hapiVolume, err := ToHapiVolume(volume, addVolumeOptions)

	// Step 4: Validate the result
	assert.NoError(t, err)
	assert.NotNil(t, hapiVolume)
	assert.Equal(t, "test-volume", hapiVolume.DataVolumeTemplate.Metadata.Name)
	assert.Equal(t, "default", hapiVolume.DataVolumeTemplate.Metadata.Namespace)
	assert.Equal(t, "test-uid", hapiVolume.DataVolumeTemplate.Metadata.UID)
	assert.Equal(t, int64(30), hapiVolume.DataVolumeTemplate.Metadata.DeletionGracePeriodSeconds)
	// Add more assertions as necessary to validate other fields
}

func int64Ptr(i int64) *int64 {
	return &i
}

func TestFromHapiVolume(t *testing.T) {
	// Step 1: Prepare sample V1VMAddVolumeEntity object
	hapiVolume := &models.V1VMAddVolumeEntity{
		DataVolumeTemplate: &models.V1VMDataVolumeTemplateSpec{
			Metadata: &models.V1VMObjectMeta{
				Annotations:                map[string]string{"test-annotation-key": "test-annotation-value"},
				DeletionGracePeriodSeconds: 30,
				Finalizers:                 []string{"test-finalizer"},
				GenerateName:               "test-",
				Generation:                 1,
				Labels:                     map[string]string{"test-label-key": "test-label-value"},
				Name:                       "test-volume",
				Namespace:                  "default",
				ResourceVersion:            "123456",
				UID:                        "test-uid",
			},
			Spec: &models.V1VMDataVolumeSpec{},
		},
	}

	// Step 2: Call the function
	volume, err := FromHapiVolume(hapiVolume)

	// Step 3: Validate the result
	assert.NoError(t, err)
	assert.NotNil(t, volume)
	assert.Equal(t, "test-volume", volume.Name)
	assert.Equal(t, "default", volume.Namespace)
	assert.Equal(t, "test-uid", string(volume.UID))
	assert.Equal(t, "test-annotation-value", volume.Annotations["test-annotation-key"])
	assert.Equal(t, int64(30), *volume.DeletionGracePeriodSeconds)

	// Validate the Spec
	specJson, err := json.Marshal(hapiVolume.DataVolumeTemplate.Spec)
	assert.NoError(t, err)

	var expectedSpec cdiv1.DataVolumeSpec
	err = json.Unmarshal(specJson, &expectedSpec)
	assert.NoError(t, err)
	assert.Equal(t, expectedSpec, volume.Spec)
}

func TestToHapiVmStatusM(t *testing.T) {
	// Step 1: Prepare sample VirtualMachineStatus object
	status := kubevirtapiv1.VirtualMachineStatus{

		PrintableStatus: "Running",
	}

	// Step 2: Call the function
	hapiVmStatus, err := ToHapiVmStatusM(status)

	// Step 3: Validate the result
	assert.NoError(t, err)
	assert.NotNil(t, hapiVmStatus)

	// Validate the fields that are mapped
	assert.Equal(t, "Running", hapiVmStatus.PrintableStatus)

}

func TestToHapiVmSpecM(t *testing.T) {
	spec := kubevirtapiv1.VirtualMachineSpec{
		Running: func(b bool) *bool { return &b }(true),
	}

	hapiVmSpec, err := ToHapiVmSpecM(spec)
	assert.NoError(t, err)
	assert.NotNil(t, hapiVmSpec)
}

func TestToHapiVm(t *testing.T) {
	vm := &kubevirtapiv1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:                       "test-vm",
			Namespace:                  "default",
			Annotations:                map[string]string{"key": "value"},
			DeletionGracePeriodSeconds: func(i int64) *int64 { return &i }(30),
			Finalizers:                 []string{"finalizer"},
			GenerateName:               "test-vm-",
			Generation:                 1,
			Labels:                     map[string]string{"label": "value"},
			ManagedFields: []metav1.ManagedFieldsEntry{
				{
					APIVersion: "v1",
					FieldsType: "FieldsV1",
					FieldsV1:   &metav1.FieldsV1{Raw: []byte("raw-data")},
					Manager:    "manager",
					Operation:  "Update",
				},
			},
			ResourceVersion: "12345",
			UID:             "uid12345",
		},
		Spec: kubevirtapiv1.VirtualMachineSpec{
			Running: func(b bool) *bool { return &b }(true),
		},
		Status: kubevirtapiv1.VirtualMachineStatus{
			Ready:           false,
			PrintableStatus: "Running",
		},
	}

	hapiVM, err := ToHapiVm(vm)
	assert.NoError(t, err)
	assert.NotNil(t, hapiVM)

	_, err = json.Marshal(vm.Spec)
	assert.NoError(t, err)

	assert.Equal(t, vm.ObjectMeta.Name, hapiVM.Metadata.Name)
	assert.Equal(t, vm.ObjectMeta.Namespace, hapiVM.Metadata.Namespace)
	assert.Equal(t, vm.ObjectMeta.Annotations, hapiVM.Metadata.Annotations)
	assert.Equal(t, *vm.DeletionGracePeriodSeconds, hapiVM.Metadata.DeletionGracePeriodSeconds)
	assert.Equal(t, vm.ObjectMeta.Finalizers, hapiVM.Metadata.Finalizers)
	assert.Equal(t, vm.ObjectMeta.GenerateName, hapiVM.Metadata.GenerateName)
	assert.Equal(t, vm.ObjectMeta.Generation, hapiVM.Metadata.Generation)
	assert.Equal(t, vm.ObjectMeta.Labels, hapiVM.Metadata.Labels)
	assert.Equal(t, vm.ObjectMeta.ResourceVersion, hapiVM.Metadata.ResourceVersion)
	assert.Equal(t, string(vm.ObjectMeta.UID), hapiVM.Metadata.UID)
}

func TestToKubevirtVMStatusM(t *testing.T) {
	status := &models.V1ClusterVirtualMachineStatus{
		Created:            true,
		Ready:              true,
		SnapshotInProgress: "snapshot-in-progress",
		RestoreInProgress:  "restore-in-progress",
		PrintableStatus:    "Running",
	}

	kubevirtStatus, err := ToKubevirtVMStatusM(status)
	assert.NoError(t, err)
	assert.NotNil(t, kubevirtStatus)

}

func TestToKubevirtVMStatus(t *testing.T) {
	status := &models.V1ClusterVirtualMachineStatus{
		Created:            true,
		Ready:              true,
		SnapshotInProgress: "snapshot-in-progress",
		RestoreInProgress:  "restore-in-progress",
		PrintableStatus:    "Running",
		Conditions: []*models.V1VMVirtualMachineCondition{
			{
				Type:    ptr.To("Ready"),
				Status:  ptr.To("True"),
				Reason:  "Ready",
				Message: "VM is ready",
			},
		},
	}

	kubevirtStatus := ToKubevirtVMStatus(status)
	assert.NotNil(t, kubevirtStatus)

	assert.Equal(t, status.Created, kubevirtStatus.Created)
	assert.Equal(t, status.Ready, kubevirtStatus.Ready)
	assert.Equal(t, status.SnapshotInProgress, *kubevirtStatus.SnapshotInProgress)
	assert.Equal(t, status.RestoreInProgress, *kubevirtStatus.RestoreInProgress)
	assert.Equal(t, kubevirtapiv1.VirtualMachinePrintableStatus(status.PrintableStatus), kubevirtStatus.PrintableStatus)

	assert.Len(t, kubevirtStatus.Conditions, len(status.Conditions))
	for i, condition := range kubevirtStatus.Conditions {
		assert.Equal(t, kubevirtapiv1.VirtualMachineConditionType(*status.Conditions[i].Type), condition.Type)
		assert.Equal(t, k8sv1.ConditionStatus(*status.Conditions[i].Status), condition.Status)
		assert.Equal(t, status.Conditions[i].Reason, condition.Reason)
		assert.Equal(t, status.Conditions[i].Message, condition.Message)
	}
}

func TestToKubevirtVMSpecM(t *testing.T) {
	// Create a test input for V1ClusterVirtualMachineSpec
	testSpec := &models.V1ClusterVirtualMachineSpec{
		DataVolumeTemplates: []*models.V1VMDataVolumeTemplateSpec{
			{
				APIVersion: "",
				Kind:       "",
				Metadata: &models.V1VMObjectMeta{
					Name: "test-volume",
					UID:  "test-uid-volume",
				},
				Spec: nil,
			},
		},
		Instancetype: &models.V1VMInstancetypeMatcher{
			InferFromVolume: "",
			Kind:            "",
			Name:            "test-instance-type",
			RevisionName:    "testins",
		},
		Preference: &models.V1VMPreferenceMatcher{
			InferFromVolume: "test-vol",
			Kind:            "node",
			Name:            "test-pref",
			RevisionName:    "testpref",
		},
		RunStrategy: "new",
		Running:     true,
		Template: &models.V1VMVirtualMachineInstanceTemplateSpec{
			Metadata: &models.V1VMObjectMeta{
				Name: "test-tem",
				UID:  "test-uid-template",
			},
			Spec: &models.V1VMVirtualMachineInstanceSpec{
				AccessCredentials:             nil,
				Affinity:                      nil,
				DNSConfig:                     nil,
				DNSPolicy:                     "test-1",
				Domain:                        nil,
				EvictionStrategy:              "ready",
				Hostname:                      "127.0.0.1",
				LivenessProbe:                 nil,
				Networks:                      nil,
				NodeSelector:                  nil,
				PriorityClassName:             "level",
				ReadinessProbe:                nil,
				SchedulerName:                 "auto",
				StartStrategy:                 "run",
				Subdomain:                     "test.test.com",
				TerminationGracePeriodSeconds: 10,
				Tolerations:                   nil,
				TopologySpreadConstraints:     nil,
				Volumes:                       nil,
			},
		},
	}

	// Call the function under test
	_, err := ToKubevirtVMSpecM(testSpec)

	// Assert no error occurred
	assert.NoError(t, err)

}

func TestToKubevirtVM(t *testing.T) {
	hapiVM := &models.V1ClusterVirtualMachine{
		Kind:       "VirtualMachine",
		APIVersion: "kubevirt.io/v1",
		Metadata: &models.V1VMObjectMeta{
			Name:      "test-vm",
			Namespace: "test-namespace",
			UID:       "123456",
			OwnerReferences: []*models.V1VMOwnerReference{
				{
					APIVersion:         ptr.To("v1"),
					BlockOwnerDeletion: true,
					Controller:         true,
					Kind:               ptr.To("ReplicaSet"),
					Name:               ptr.To("test-owner"),
					UID:                ptr.To("654321"),
				},
			},
			ManagedFields: []*models.V1VMManagedFieldsEntry{
				{
					APIVersion: "v1",
					FieldsType: "FieldsV1",
					FieldsV1: &models.V1VMFieldsV1{
						Raw: []strfmt.Base64{strfmt.Base64("c29tZS1maWVsZHM=")},
					},
					Manager:   "kubectl",
					Operation: "Apply",
				},
			},
		},
		Spec: &models.V1ClusterVirtualMachineSpec{
			Running: true,
		},
		Status: &models.V1ClusterVirtualMachineStatus{
			Created: true,
		},
	}

	kubevirtVM, err := ToKubevirtVM(hapiVM)
	assert.NoError(t, err)
	assert.NotNil(t, kubevirtVM)
	assert.Equal(t, "test-vm", kubevirtVM.Name)
	assert.Equal(t, "test-namespace", kubevirtVM.Namespace)
	assert.Equal(t, 1, len(kubevirtVM.OwnerReferences))
	assert.Equal(t, "654321", string(kubevirtVM.OwnerReferences[0].UID))
	assert.Equal(t, "Apply", string(kubevirtVM.ManagedFields[0].Operation))
}

func TestToKubevirtVM_NilInput(t *testing.T) {
	kubevirtVM, err := ToKubevirtVM(nil)
	assert.Error(t, err)
	assert.Nil(t, kubevirtVM)
	assert.Equal(t, errors.New("hapiVM is nil"), err)
}
