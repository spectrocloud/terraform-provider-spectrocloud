package spectrocloud

import (
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"strings"
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
	if cloudConfig["ssh_key"] != "" && len(cloudConfig["ssh_keys"].(*schema.Set).List()) == 0 {
		sshKeys = []string{strings.TrimSpace(cloudConfig["ssh_key"].(string))}
	} else if cloudConfig["ssh_key"] == "" && len(cloudConfig["ssh_keys"].(*schema.Set).List()) >= 0 {
		for _, sk := range cloudConfig["ssh_keys"].(*schema.Set).List() {
			sshKeys = append(sshKeys, strings.TrimSpace(sk.(string)))
		}
	} else {
		return nil, errors.New("ssh_key: Kindly specify any one attribute ssh_key or ssh_keys")
	}
	return sshKeys, nil
}
