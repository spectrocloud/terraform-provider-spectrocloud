package test

import (
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud"
	"testing"
)

func TestFlattenMachinePoolConfigsAwsSubnetIds(t *testing.T) {
	var machinePoolConfig []*models.V1AwsMachinePoolConfig
	addLabels := make(map[string]string)
	addLabels["by"] = "Siva"
	addLabels["purpose"] = "unittest"

	addTags := make(map[string]string)
	addTags["owner"] = "Siva"
	addTags["project"] = "test"

	subnetIdsMaster := make(map[string]string)
	subnetIdsMaster["us-east-2a"] = "subnet-031a7ff4ff5e7fb9a"
	subnetIdsMaster["us-east-2a"] = "subnet-08864975df862eb58"

	subnetIdsWorker := make(map[string]string)
	subnetIdsWorker["us-east-2a"] = "subnet-08864975df862eb58"

	isControl := func(b bool) *bool { return &b }(true)
	machinePoolConfig = append(machinePoolConfig, &models.V1AwsMachinePoolConfig{
		Name:                    "master-pool",
		IsControlPlane:          isControl,
		InstanceType:            "t3.large",
		Size:                    1,
		AdditionalLabels:        addLabels,
		AdditionalTags:          addTags,
		RootDeviceSize:          10,
		UseControlPlaneAsWorker: true,
		UpdateStrategy: &models.V1UpdateStrategy{
			Type: "",
		},
		SubnetIds: subnetIdsMaster,
	})
	machinePoolConfig = append(machinePoolConfig, &models.V1AwsMachinePoolConfig{
		Name:             "worker-pool",
		InstanceType:     "t3.large",
		Size:             3,
		AdditionalLabels: addLabels,
		AdditionalTags:   addTags,
		UpdateStrategy: &models.V1UpdateStrategy{
			Type: "",
		},
		SubnetIds: subnetIdsWorker,
	})
	machinePools := spectrocloud.FlattenMachinePoolConfigsAws(machinePoolConfig)
	if len(machinePools) != 2 {
		t.Fail()
		t.Logf("Machine pool for master and worker is not returned by func - FlattenMachinePoolConfigsAws")
	} else {
		for i, _ := range machinePools {
			k := machinePools[i].(map[string]interface{})
			if k["update_strategy"] != "RollingUpdateScaleOut" {
				t.Fail()
				t.Logf("Machine pool update strategy is not set to it default - RollingUpdateScaleOut")
			}
			if k["name"] == "master-pool" {
				if k["control_plane_as_worker"].(bool) != machinePoolConfig[i].UseControlPlaneAsWorker ||
					k["count"].(int) != int(machinePoolConfig[i].Size) || k["disk_size_gb"].(int) != int(machinePoolConfig[i].RootDeviceSize) || k["instance_type"].(string) != string(machinePoolConfig[i].InstanceType) {
					t.Fail()
					t.Logf("Machine pool value not with defined schema")
				}

			} else {
				if k["count"].(int) != int(machinePoolConfig[i].Size) || k["instance_type"].(string) != string(machinePoolConfig[i].InstanceType) {
					t.Fail()
					t.Logf("Machine pool value not with defined schema")
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

	addTags := make(map[string]string)
	addTags["owner"] = "Siva"
	addTags["project"] = "test"
	azs := []string{"us-east-2a"}

	isControl := func(b bool) *bool { return &b }(true)
	machinePoolConfig = append(machinePoolConfig, &models.V1AwsMachinePoolConfig{
		Name:                    "master",
		IsControlPlane:          isControl,
		InstanceType:            "t3.xlarge",
		Size:                    1,
		AdditionalLabels:        addLabels,
		AdditionalTags:          addTags,
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
		AdditionalTags:   addTags,
		UpdateStrategy: &models.V1UpdateStrategy{
			Type: "Recreate",
		},
		Azs: azs,
	})
	machinePools := spectrocloud.FlattenMachinePoolConfigsAws(machinePoolConfig)
	if len(machinePools) != 2 {
		t.Fail()
		t.Logf("Machine pool for master and worker is not returned by func - FlattenMachinePoolConfigsAws")
	} else {
		for i, _ := range machinePools {
			k := machinePools[i].(map[string]interface{})
			if k["update_strategy"] != "Recreate" {
				t.Fail()
				t.Logf("Machine pool update strategy is not set to it default - RollingUpdateScaleOut")
			}
			if k["name"] == "master-pool" {
				if k["control_plane_as_worker"].(bool) != machinePoolConfig[i].UseControlPlaneAsWorker ||
					k["count"].(int) != int(machinePoolConfig[i].Size) || k["disk_size_gb"].(int) != int(machinePoolConfig[i].RootDeviceSize) || k["instance_type"].(string) != string(machinePoolConfig[i].InstanceType) {
					t.Fail()
					t.Logf("Machine pool value not with defined schema")
				}

			} else {
				if k["count"].(int) != int(machinePoolConfig[i].Size) || k["instance_type"].(string) != string(machinePoolConfig[i].InstanceType) {
					t.Fail()
					t.Logf("Machine pool value not with defined schema")
				}

			}
		}
	}

}
