package spectrocloud

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/palette-sdk-go/client/herr"
)

var (
	DefaultDiskType = "Standard_LRS"
	DefaultDiskSize = 60
	NameToCloudType = map[string]string{
		"spectrocloud_cluster_aks":               "aks",
		"spectrocloud_cluster_aws":               "aws",
		"spectrocloud_cluster_azure":             "azure",
		"spectrocloud_cluster_edge_native":       "edge-native",
		"spectrocloud_cluster_eks":               "eks",
		"spectrocloud_cluster_edge_vsphere":      "edge-vsphere",
		"spectrocloud_cluster_gcp":               "gcp",
		"spectrocloud_cluster_maas":              "maas",
		"spectrocloud_cluster_openstack":         "openstack",
		"spectrocloud_cluster_vsphere":           "vsphere",
		"spectrocloud_cluster_gke":               "gke",
		"spectrocloud_cluster_apache_cloudstack": "apache-cloudstack",
	}
	//clusterVsphereKeys = []string{"name", "context", "tags", "description", "cluster_meta_attribute", "cluster_profile", "apply_setting", "cloud_account_id", "cloud_config_id", "review_repave_state", "pause_agent_upgrades", "os_patch_on_boot", "os_patch_schedule", "os_patch_after", "kubeconfig", "admin_kube_config", "cloud_config", "machine_pool", "backup_policy", "scan_policy", "cluster_rbac_binding", "namespaces", "host_config", "location_config", "skip_completion", "force_delete", "force_delete_delay"}
)

const (
	tenantString  = "tenant"
	projectString = "project"
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
	config := &models.V1ClusterConfigEntity{
		ClusterMetaAttribute:    toClusterMetaAttribute(d),
		MachineManagementConfig: toMachineManagementConfig(d),
		Resources:               toClusterResourceConfig(d),
		HostClusterConfig:       toClusterHostConfigs(d),
		Location:                toClusterLocationConfigs(d),
		Timezone:                toClusterTimezone(d),
	}

	// Set UpdateWorkerPoolsInParallel if specified
	if v, ok := d.GetOk("update_worker_pools_in_parallel"); ok {
		config.UpdateWorkerPoolsInParallel = v.(bool)
	}

	return config
}

func toClusterMetaAttribute(d *schema.ResourceData) string {
	clusterMetadataAttribute := ""
	if v, ok := d.GetOk("cluster_meta_attribute"); ok {
		clusterMetadataAttribute = v.(string)
	}
	return clusterMetadataAttribute
}

func toClusterTimezone(d *schema.ResourceData) string {
	timezone := ""
	if v, ok := d.GetOk("cluster_timezone"); ok {
		timezone = v.(string)
	}
	return timezone
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
	if v, ok := d.GetOk("pause_agent_upgrades"); ok {
		setting := &models.V1ClusterUpgradeSettingsEntity{
			SpectroComponents: v.(string),
		}
		if err := c.UpdatePauseAgentUpgradeSettingCluster(setting, d.Id()); err != nil {
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

	if cluster.Spec.ClusterConfig.Timezone != "" {
		if err := d.Set("cluster_timezone", cluster.Spec.ClusterConfig.Timezone); err != nil {
			return err
		}
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

func GetCommonCluster(d *schema.ResourceData, m interface{}) (*client.V1Client, error) {
	// Parse resource ID and scope (import ID format: id_or_name:context e.g. "my-cluster:project" or "uid-123:project")
	resourceContext, clusterIDOrName, err := ParseResourceID(d)
	if err != nil {
		return nil, err
	}
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Try by UID first, then by name (physical then virtual)
	cluster, err := c.GetCluster(clusterIDOrName)
	if err == nil && cluster != nil {
		return setClusterImportState(d, c, cluster.Metadata.UID, cluster)
	}
	if err != nil && !herr.IsNotFound(err) {
		return c, fmt.Errorf("unable to retrieve cluster data: %s", err)
	}
	cluster, err = c.GetClusterByName(clusterIDOrName, false)
	if err == nil && cluster != nil {
		return setClusterImportState(d, c, cluster.Metadata.UID, cluster)
	}

	return c, fmt.Errorf("unable to retrieve cluster by UID or name '%s': not found in context %s", clusterIDOrName, resourceContext)
}

// setClusterImportState sets name, context, and ID on d after a successful cluster lookup (import).
func setClusterImportState(d *schema.ResourceData, c *client.V1Client, clusterUID string, cluster *models.V1SpectroCluster) (*client.V1Client, error) {
	if err := d.Set("name", cluster.Metadata.Name); err != nil {
		return c, err
	}
	if err := d.Set("context", cluster.Metadata.Annotations["scope"]); err != nil {
		return c, err
	}
	d.SetId(clusterUID)
	return c, nil
}

func generalWarningForRepave(diags *diag.Diagnostics) {
	message := "Please note that certain day 2 operations on a running cluster may trigger a node pool repave or a full repave of your cluster. This process might temporarily affect your cluster’s performance or configuration. For more details, please refer to the https://docs.spectrocloud.com/clusters/cluster-management/node-pool/"
	*diags = append(*diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "Warning",
		Detail:   message,
	})
}

func flattenCommonAttributeForCustomClusterImport(c *client.V1Client, d *schema.ResourceData) error {
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

	if cluster.Spec.ClusterConfig.Timezone != "" {
		if err := d.Set("cluster_timezone", cluster.Spec.ClusterConfig.Timezone); err != nil {
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

func flattenCloudConfigGeneric(configUID string, d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err)
	}
	return diag.Diagnostics{}
}

func validateCloudType(data interface{}, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	inCloudType := data.(string)
	for _, cloudType := range []string{"aws", "azure", "gcp", "vsphere", "generic"} {
		if cloudType == inCloudType {
			return diags
		}
	}
	return diag.FromErr(fmt.Errorf("cloud type '%s' is invalid. valid cloud types are %v", inCloudType, "cloud_types"))
}

func toTagsMap(d *schema.ResourceData) map[string]string {
	tags := make(map[string]string)
	if d.Get("tags_map") != nil {
		for k, v := range d.Get("tags_map").(map[string]interface{}) {
			vStr := v.(string)
			if v != "" {
				tags[k] = vStr
			} else {
				tags[k] = "spectro__tag"
			}
		}
		return tags
	} else {
		return nil
	}
}

func flattenTagsMap(labels map[string]string) map[string]string {
	tags := make(map[string]string)
	if len(labels) > 0 {
		for k, v := range labels {
			tags[k] = v
		}
		return tags
	} else {
		return nil
	}
}

// updateClusterTimezone updates the timezone configuration for a cluster.
func updateClusterTimezone(c *client.V1Client, d *schema.ResourceData) error {
	if v, ok := d.GetOk("cluster_timezone"); ok {
		timezone := v.(string)
		if err := c.UpdateClusterTimezone(d.Id(), timezone); err != nil {
			return err
		}
	}
	return nil
}

// toClusterType converts the terraform cluster_type value to the API model.
// Returns nil if cluster_type is not set.
func toClusterType(d *schema.ResourceData) *models.V1ClusterType {
	if v, ok := d.GetOk("cluster_type"); ok {
		clusterType := models.V1ClusterType(v.(string))
		return &clusterType
	}
	return nil
}

// validateTimezone validates that the provided timezone is in valid IANA format.
// Valid examples: "America/New_York", "Asia/Kolkata", "Europe/London", "UTC"
func validateTimezone(val interface{}, key string) (warns []string, errs []error) {
	timezone := val.(string)
	if timezone == "" {
		return warns, errs
	}

	// Common validation patterns for IANA timezone format
	// IANA timezones are in format: Area/Location or Area/Location/Sublocation
	// Examples: America/New_York, Asia/Kolkata, Europe/London, UTC, GMT

	// Check for basic IANA timezone format
	// Valid patterns: UTC, GMT, or Area/Location format
	if timezone == "UTC" || timezone == "GMT" {
		return warns, errs
	}

	// Check if it contains at least one '/' for Area/Location format
	if !strings.Contains(timezone, "/") {
		errs = append(errs, fmt.Errorf(
			"%q must be a valid IANA timezone string (e.g., 'America/New_York', 'Asia/Kolkata', 'Europe/London', 'UTC'). Got: %s",
			key, timezone))
		return warns, errs
	}

	// Additional validation: timezone shouldn't have spaces or invalid characters
	if strings.Contains(timezone, " ") {
		errs = append(errs, fmt.Errorf(
			"%q timezone cannot contain spaces. Got: %s", key, timezone))
		return warns, errs
	}

	return warns, errs
}

// ValidateClusterTypeUpdate checks if cluster_type has been modified during an update operation.
// Returns an error if the cluster_type field has changed, as it is a create-only field.
func ValidateClusterTypeUpdate(d *schema.ResourceData) error {
	if d.HasChange("cluster_type") {
		oldVal, newVal := d.GetChange("cluster_type")
		return fmt.Errorf("cluster_type cannot be modified after cluster creation. "+
			"Current value: %q, attempted new value: %q. "+
			"To change the cluster type, you must delete and recreate the cluster", oldVal, newVal)
	}
	return nil
}
