package spectrocloud

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-api-go/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"strings"
)

var (
	DefaultDiskType = "Standard_LRS"
	DefaultDiskSize = 60
	NameToCloudType = map[string]string{
		"spectrocloud_cluster_aks":          "aks",
		"spectrocloud_cluster_aws":          "aws",
		"spectrocloud_cluster_azure":        "azure",
		"spectrocloud_cluster_edge_native":  "edge-native",
		"spectrocloud_cluster_eks":          "eks",
		"spectrocloud_cluster_edge_vsphere": "edge-vsphere",
		"spectrocloud_cluster_gcp":          "gcp",
		"spectrocloud_cluster_maas":         "maas",
		"spectrocloud_cluster_openstack":    "openstack",
		"spectrocloud_cluster_tke":          "tke",
		"spectrocloud_cluster_vsphere":      "vsphere",
		"spectrocloud_cluster_gke":          "gke",
	}
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
		ClusterMetaAttribute:    toClusterMetaAttribute(d),
		MachineManagementConfig: toMachineManagementConfig(d),
		Resources:               toClusterResourceConfig(d),
		HostClusterConfig:       toClusterHostConfigs(d),
		Location:                toClusterLocationConfigs(d),
	}
}

func toClusterMetaAttribute(d *schema.ResourceData) string {
	clusterMetadataAttribute := ""
	if v, ok := d.GetOk("cluster_meta_attribute"); ok {
		clusterMetadataAttribute = v.(string)
	}
	return clusterMetadataAttribute
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
	if cloudConfig["ssh_key"] != nil {
		sshKey := cloudConfig["ssh_key"].(string)
		if sshKey != "" {
			sshKeys = append(sshKeys, strings.TrimSpace(sshKey))
		}
	}
	if len(sshKeysList) > 0 || len(sshKeys) > 0 {
		for _, sk := range sshKeysList {
			sshKeys = append(sshKeys, strings.TrimSpace(sk.(string)))
		}
		return sshKeys, nil
	}
	return nil, errors.New("validation ssh_key: Kindly specify any one attribute ssh_key or ssh_keys")
}

func FlattenControlPlaneAndRepaveInterval(isControlPlane *bool, oi map[string]interface{}, nodeRepaveInterval int32) {
	if isControlPlane != nil {
		oi["control_plane"] = *isControlPlane
		if !*isControlPlane {
			oi["node_repave_interval"] = int32(nodeRepaveInterval)
		}
	}
}

func ValidationNodeRepaveIntervalForControlPlane(nodeRepaveInterval int) error {
	if nodeRepaveInterval != 0 {
		errMsg := fmt.Sprintf("Validation error: The `node_repave_interval` attribute is not applicable for the control plane. Attempted value: %d.", nodeRepaveInterval)
		return errors.New(errMsg)
	}
	return nil
}

func ValidateContext(context string) error {
	if context != "project" && context != "tenant" {
		return fmt.Errorf("invalid Context set - %s", context)
	}
	return nil
}

func ValidateCloudType(resourceName string, cluster *models.V1SpectroCluster) error {
	if cluster.Spec == nil {
		return fmt.Errorf("cluster spec is nil in cluster %s", cluster.Metadata.UID)
	}
	if cluster.Spec.CloudType != NameToCloudType[resourceName] {
		return fmt.Errorf("resource with id %s is not of type %s, need to correct resource type", cluster.Metadata.UID, resourceName)
	}
	return nil
}

func updateAgentUpgradeSetting(c *client.V1Client, d *schema.ResourceData) error {
	clusterContext := d.Get("context").(string)
	if v, ok := d.GetOk("pause_agent_upgrades"); ok {
		setting := &models.V1ClusterUpgradeSettingsEntity{
			SpectroComponents: v.(string),
		}
		if err := c.UpdatePauseAgentUpgradeSettingCluster(setting, d.Id(), clusterContext); err != nil {
			return err
		}
	}
	return nil
}

// This function is called during import cluster from palette to set default terraform value
func flattenCommonAttributeForClusterImport(c *client.V1Client, d *schema.ResourceData) error {
	clusterProfiles, err := flattenClusterProfileForImport(c, d)
	if err != nil {
		return err
	}
	err = d.Set("cluster_profile", clusterProfiles)
	if err != nil {
		return err
	}

	var diags diag.Diagnostics
	cluster, err := resourceClusterRead(d, c, diags)
	if err != nil {
		return err
	}

	if cluster.Metadata.Annotations["description"] != "" {
		if err := d.Set("description", cluster.Metadata.Annotations["description"]); err != nil {
			return err
		}
	}

	if cluster.Status.SpcApply != nil {
		err = d.Set("apply_setting", cluster.Status.SpcApply.ActionType)
		if err != nil {
			return err
		}
	}

	err = d.Set("pause_agent_upgrades", getSpectroComponentsUpgrade(cluster))
	if err != nil {
		return err
	}
	if cluster.Spec.ClusterConfig.MachineManagementConfig != nil {
		err = d.Set("os_patch_on_boot", cluster.Spec.ClusterConfig.MachineManagementConfig.OsPatchConfig.PatchOnBoot)
		if err != nil {
			return err
		}
		err = d.Set("os_patch_schedule", cluster.Spec.ClusterConfig.MachineManagementConfig.OsPatchConfig.Schedule)
		if err != nil {
			return err
		}
	}
	if cluster.Status.Repave != nil {
		if err = d.Set("review_repave_state", cluster.Status.Repave.State); err != nil {
			return err
		}
	}
	err = d.Set("force_delete", false)
	if err != nil {
		return err
	}
	err = d.Set("force_delete_delay", 20)
	if err != nil {
		return err
	}
	err = d.Set("skip_completion", false)
	if err != nil {
		return err
	}
	return nil
}

func GetCommonCluster(d *schema.ResourceData, c *client.V1Client) error {
	// parse resource ID and scope
	_, clusterID, err := ParseResourceID(d)
	if err != nil {
		return err
	}

	// Use the IDs to retrieve the cluster data from the API
	cluster, err := c.GetCluster(clusterID)
	if err != nil {
		return fmt.Errorf("unable to retrieve cluster data: %s", err)
	}
	if cluster != nil {
		err = d.Set("name", cluster.Metadata.Name)
		if err != nil {
			return err
		}
		err = d.Set("context", cluster.Metadata.Annotations["scope"])
		if err != nil {
			return err
		}

		// Set the ID of the resource in the state. This ID is used to track the
		// resource and must be set in the state during the import.
		d.SetId(clusterID)
	} else {
		return fmt.Errorf("couldn’t find cluster. Kindly check the cluster UID and context")
	}

	return nil
}
