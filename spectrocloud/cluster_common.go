package spectrocloud

import (
	"errors"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
)

var (
	DefaultDiskType = "Standard_LRS"
	DefaultDiskSize = 60
)

func toNtpServers(in map[string]interface{}) []string {
	servers := make([]string, 0, 1)
	if _, ok := in["ntp_servers"]; ok {
		for _, t := range in["ntp_servers"].(*schema.Set).List() {
			ntp := t.(string)
			servers = append(servers, ntp)
		}
	}
	return servers
}

func toClusterConfig(d *schema.ResourceData) *models.V1ClusterConfigEntity {
	return &models.V1ClusterConfigEntity{
		MachineManagementConfig: toMachineManagementConfig(d),
		Resources:               toClusterResourceConfig(d),
		HostClusterConfig:       toClusterHostConfigs(d),
		Location:                toClusterLocationConfigs(d),
	}
}

func toMachineManagementConfig(d *schema.ResourceData) *models.V1MachineManagementConfig {
	return &models.V1MachineManagementConfig{
		OsPatchConfig: toOsPatchConfig(d),
	}
}

func toClusterResourceConfig(d *schema.ResourceData) *models.V1ClusterResourcesEntity {
	return &models.V1ClusterResourcesEntity{
		Namespaces: toClusterNamespaces(d),
		Rbacs:      toClusterRBACsInputEntities(d),
	}
}

func toSSHKeys(cloudConfig map[string]interface{}) ([]string, error) {
	var sshKeys []string
	var sshKeysList []interface{}
	if cloudConfig["ssh_keys"] != nil {
		sshKeysList = cloudConfig["ssh_keys"].(*schema.Set).List()
	}
	sshKey := cloudConfig["ssh_key"].(string)
	if sshKey != "" && len(sshKeysList) == 0 {
		sshKeys = []string{strings.TrimSpace(sshKey)}
		return sshKeys, nil
	}
	if sshKey == "" && len(sshKeysList) > 0 {
		for _, sk := range sshKeysList {
			sshKeys = append(sshKeys, strings.TrimSpace(sk.(string)))
		}
		return sshKeys, nil
	}
	return nil, errors.New("validation ssh_key: Kindly specify any one attribute ssh_key or ssh_keys")
}
