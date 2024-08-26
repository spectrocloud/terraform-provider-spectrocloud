package spectrocloud

import (
	"reflect"
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func TestFlattenMachinePoolConfigsAwsSubnetIds(t *testing.T) {
	var machinePoolConfig []*models.V1AwsMachinePoolConfig
	addLabels := make(map[string]string)
	addLabels["by"] = "Siva"
	addLabels["purpose"] = "unittest"

	subnetIdsCP := make(map[string]string)
	subnetIdsCP["us-east-2a"] = "subnet-031a7ff4ff5e7fb9a"

	subnetIdsWorker := make(map[string]string)
	subnetIdsWorker["us-east-2a"] = "subnet-08864975df862eb58"

	isControl := func(b bool) *bool { return &b }(true)
	machinePoolConfig = append(machinePoolConfig, &models.V1AwsMachinePoolConfig{
		Name:                    "cp-pool",
		IsControlPlane:          isControl,
		InstanceType:            "t3.large",
		Size:                    1,
		AdditionalLabels:        addLabels,
		RootDeviceSize:          10,
		UseControlPlaneAsWorker: true,
		UpdateStrategy: &models.V1UpdateStrategy{
			Type: "",
		},
		SubnetIds: subnetIdsCP,
	})
	machinePoolConfig = append(machinePoolConfig, &models.V1AwsMachinePoolConfig{
		Name:             "worker-pool",
		InstanceType:     "t3.large",
		Size:             3,
		AdditionalLabels: addLabels,
		UpdateStrategy: &models.V1UpdateStrategy{
			Type: "",
		},
		SubnetIds: subnetIdsWorker,
	})
	machinePools := flattenMachinePoolConfigsAws(machinePoolConfig)
	if len(machinePools) != 2 {
		t.Fail()
		t.Logf("Machine pool for control-plane and worker is not returned by func - FlattenMachinePoolConfigsAws")
	} else {
		for i := range machinePools {
			k := machinePools[i].(map[string]interface{})
			if k["update_strategy"] != "RollingUpdateScaleOut" {
				t.Errorf("Machine pool - update strategy is not matching got %v, wanted %v", k["update_strategy"], "RollingUpdateScaleOut")
				t.Fail()
			}
			if k["count"].(int) != int(machinePoolConfig[i].Size) {
				t.Errorf("Machine pool - count is not matching got %v, wanted %v", k["count"].(string), int(machinePoolConfig[i].Size))
				t.Fail()
			}
			if k["instance_type"].(string) != string(machinePoolConfig[i].InstanceType) {
				t.Errorf("Machine pool - instance_type is not matching got %v, wanted %v", k["instance_type"].(string), string(machinePoolConfig[i].InstanceType))
				t.Fail()
			}
			if !validateMapString(addLabels, k["additional_labels"].(map[string]string)) {
				t.Errorf("Machine pool - additional labels is not matching got %v, wanted %v", addLabels, k["additional_labels"])
				t.Fail()
			}
			if k["name"] == "cp-pool" {
				if k["control_plane_as_worker"].(bool) != machinePoolConfig[i].UseControlPlaneAsWorker {
					t.Errorf("Machine pool - control_plane_as_worker is not matching got %s, wanted %v", k["control_plane_as_worker"].(string), machinePoolConfig[i].UseControlPlaneAsWorker)
					t.Fail()
				}
				if k["disk_size_gb"].(int) != int(machinePoolConfig[i].RootDeviceSize) {
					t.Errorf("Machine pool - disk_size_gb is not matching got %v, wanted %v", k["disk_size_gb"].(int), int(machinePoolConfig[i].RootDeviceSize))
					t.Fail()
				}
				if !validateMapString(subnetIdsCP, k["az_subnets"].(map[string]string)) {
					t.Errorf("Machine pool - additional labels is not matching got %v, wanted %v", subnetIdsCP, k["az_subnets"])
					t.Fail()
				}

			} else {
				if !validateMapString(subnetIdsWorker, k["az_subnets"].(map[string]string)) {
					t.Errorf("Machine pool - additional labels is not matching got %v, wanted %v", subnetIdsWorker, k["az_subnets"])
					t.Fail()
				}
			}
		}
	}

}

func TestFlattenMachinePoolConfigsAwsAZ(t *testing.T) {
	var machinePoolConfig []*models.V1AwsMachinePoolConfig
	addLabels := make(map[string]string)
	addLabels["by"] = "Siva"
	addLabels["purpose"] = "unittest"

	azs := []string{"us-east-2a"}

	isControl := func(b bool) *bool { return &b }(true)
	machinePoolConfig = append(machinePoolConfig, &models.V1AwsMachinePoolConfig{
		Name:                    "cp",
		IsControlPlane:          isControl,
		InstanceType:            "t3.xlarge",
		Size:                    1,
		AdditionalLabels:        addLabels,
		RootDeviceSize:          10,
		UseControlPlaneAsWorker: true,
		UpdateStrategy: &models.V1UpdateStrategy{
			Type: "Recreate",
		},
		Azs: azs,
	})
	machinePoolConfig = append(machinePoolConfig, &models.V1AwsMachinePoolConfig{
		Name:             "worker",
		InstanceType:     "t3.xlarge",
		Size:             3,
		AdditionalLabels: addLabels,
		UpdateStrategy: &models.V1UpdateStrategy{
			Type: "Recreate",
		},
		Azs: azs,
	})
	machinePools := flattenMachinePoolConfigsAws(machinePoolConfig)
	if len(machinePools) != 2 {
		t.Fail()
		t.Logf("Machine pool for control-plane and worker is not returned by func - FlattenMachinePoolConfigsAws")
	} else {
		for i := range machinePools {
			k := machinePools[i].(map[string]interface{})
			if k["update_strategy"] != "Recreate" {
				t.Errorf("Machine pool - update strategy is not matching got %v, wanted %v", k["update_strategy"], "Recreate")
				t.Fail()
			}
			if k["count"].(int) != int(machinePoolConfig[i].Size) {
				t.Errorf("Machine pool - count is not matching got %s, wanted %v", k["count"].(string), int(machinePoolConfig[i].Size))
				t.Fail()
			}
			if k["instance_type"].(string) != string(machinePoolConfig[i].InstanceType) {
				t.Errorf("Machine pool - instance_type is not matching got %s, wanted %s", k["instance_type"].(string), string(machinePoolConfig[i].InstanceType))
				t.Fail()
			}
			if !validateMapString(addLabels, k["additional_labels"].(map[string]string)) {
				t.Errorf("Machine pool - additional labels is not matching got %v, wanted %v", addLabels, k["additional_labels"])
				t.Fail()
			}
			if !reflect.DeepEqual(azs, k["azs"]) {
				t.Errorf("Machine pool - AZS is not matching got %v, wanted %v", azs, k["azs"])
				t.Fail()
			}
			if k["name"] == "cp-pool" {
				if k["control_plane_as_worker"].(bool) != machinePoolConfig[i].UseControlPlaneAsWorker {
					t.Errorf("Machine pool - control_plane_as_worker is not matching got %s, wanted %v", k["control_plane_as_worker"].(string), machinePoolConfig[i].UseControlPlaneAsWorker)
					t.Fail()
				}
				if k["disk_size_gb"].(int) != int(machinePoolConfig[i].RootDeviceSize) {
					t.Errorf("Machine pool - disk_size_gb is not matching got %v, wanted %v", k["disk_size_gb"].(int), int(machinePoolConfig[i].RootDeviceSize))
					t.Fail()
				}
			}
		}
	}
}

func validateMapString(src map[string]string, dest map[string]string) bool {
	return reflect.DeepEqual(src, dest)
}
