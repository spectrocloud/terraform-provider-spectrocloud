package spectrocloud

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"log"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/robfig/cron"
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

var (
	DefaultDiskType = "Standard_LRS"
	DefaultDiskSize = 60
)

var resourceClusterDeletePendingStates = []string{
	"Pending",
	"Provisioning",
	"Running",
	"Deleting",
	"Importing",
}
var resourceClusterCreatePendingStates = []string{
	"Pending",
	"Provisioning",
	"Importing",
}

//var resourceClusterUpdatePendingStates = []string{
//	"backing-up",
//	"modifying",
//	"resetting-master-credentials",
//	"upgrading",
//}
func waitForClusterDeletion(ctx context.Context, c *client.V1Client, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    resourceClusterDeletePendingStates,
		Target:     nil, // wait for deleted
		Refresh:    resourceClusterStateRefreshFunc(c, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func updateProfiles(c *client.V1Client, d *schema.ResourceData) error {
	log.Printf("Updating profiles")
	body := &models.V1SpectroClusterProfiles{
		Profiles: toProfiles(d),
	}
	if err := c.UpdateClusterProfileValues(d.Id(), body); err != nil {
		return err
	}
	return nil
}

func toTags(d *schema.ResourceData) map[string]string {
	tags := make(map[string]string)
	if d.Get("tags") != nil {
		for _, t := range d.Get("tags").(*schema.Set).List() {
			tag := t.(string)
			if strings.Contains(tag, ":") {
				tags[strings.Split(tag, ":")[0]] = strings.Split(tag, ":")[1]
			} else {
				tags[tag] = "spectro__tag"
			}
		}
	}
	return tags
}

func flattenTags(labels map[string]string) []interface{} {
	tags := make([]interface{}, 0)
	if len(labels) > 0 {
		for k, v := range labels {
			if v == "spectro__tag" {
				tags = append(tags, k)
			} else {
				tags = append(tags, fmt.Sprintf("%s:%s", k, v))
			}
		}
	}
	return tags
}

func toPolicies(d *schema.ResourceData) *models.V1SpectroClusterPolicies {
	return &models.V1SpectroClusterPolicies{
		BackupPolicy: toBackupPolicy(d),
		ScanPolicy:   toScanPolicy(d),
	}
}

func toBackupPolicy(d *schema.ResourceData) *models.V1ClusterBackupConfig {
	if policies, found := d.GetOk("backup_policy"); found {
		//policy := policies.([]interface{})[0]
		policy := policies.([]interface{})[0].(map[string]interface{})

		namespaces := make([]string, 0, 1)
		if policy["namespaces"] != nil {
			if nss, ok := policy["namespaces"].([]interface{}); ok {
				for _, ns := range nss {
					namespaces = append(namespaces, ns.(string))
				}
			}
		}

		return &models.V1ClusterBackupConfig{
			BackupLocationUID:       policy["backup_location_id"].(string),
			BackupPrefix:            policy["prefix"].(string),
			DurationInHours:         int64(policy["expiry_in_hour"].(int)),
			IncludeAllDisks:         policy["include_disks"].(bool),
			IncludeClusterResources: policy["include_cluster_resources"].(bool),
			Namespaces:              namespaces,
			Schedule: &models.V1ClusterFeatureSchedule{
				ScheduledRunTime: policy["schedule"].(string),
			},
		}
	}
	return nil
}

func flattenBackupPolicy(policy *models.V1ClusterBackupConfig) []interface{} {
	result := make([]interface{}, 0, 1)
	data := make(map[string]interface{})
	data["schedule"] = policy.Schedule.ScheduledRunTime
	data["backup_location_id"] = policy.BackupLocationUID
	data["prefix"] = policy.BackupPrefix
	data["namespaces"] = policy.Namespaces
	data["expiry_in_hour"] = policy.DurationInHours
	data["include_disks"] = policy.IncludeAllDisks
	data["include_cluster_resources"] = policy.IncludeClusterResources
	result = append(result, data)
	return result
}

func updateBackupPolicy(c *client.V1Client, d *schema.ResourceData) error {
	if policy := toBackupPolicy(d); policy != nil {
		return c.ApplyClusterBackupConfig(d.Id(), policy)
	}
	return nil
}

func toScanPolicy(d *schema.ResourceData) *models.V1ClusterComplianceScheduleConfig {
	if profiles, found := d.GetOk("scan_policy"); found {
		config := &models.V1ClusterComplianceScheduleConfig{}
		policy := profiles.([]interface{})[0].(map[string]interface{})
		if policy["configuration_scan_schedule"] != nil {
			config.KubeBench = &models.V1ClusterComplianceScanKubeBenchScheduleConfig{
				Schedule: &models.V1ClusterFeatureSchedule{
					ScheduledRunTime: policy["configuration_scan_schedule"].(string),
				},
			}
		}
		if policy["penetration_scan_schedule"] != nil {
			config.KubeHunter = &models.V1ClusterComplianceScanKubeHunterScheduleConfig{
				Schedule: &models.V1ClusterFeatureSchedule{
					ScheduledRunTime: policy["penetration_scan_schedule"].(string),
				},
			}
		}
		if policy["conformance_scan_schedule"] != nil {
			config.Sonobuoy = &models.V1ClusterComplianceScanSonobuoyScheduleConfig{
				Schedule: &models.V1ClusterFeatureSchedule{
					ScheduledRunTime: policy["conformance_scan_schedule"].(string),
				},
			}
		}
		return config
	}
	return nil
}

func flattenScanPolicy(driverSpec map[string]models.V1ComplianceScanDriverSpec) []interface{} {
	result := make([]interface{}, 0, 1)
	data := make(map[string]interface{})
	if v, found := driverSpec["kube-bench"]; found {
		data["configuration_scan_schedule"] = v.Config.Schedule.ScheduledRunTime
	}
	if v, found := driverSpec["kube-hunter"]; found {
		data["penetration_scan_schedule"] = v.Config.Schedule.ScheduledRunTime
	}
	if v, found := driverSpec["sonobuoy"]; found {
		data["conformance_scan_schedule"] = v.Config.Schedule.ScheduledRunTime
	}
	result = append(result, data)
	return result
}

func updateScanPolicy(c *client.V1Client, d *schema.ResourceData) error {
	if policy := toScanPolicy(d); policy != nil {
		return c.ApplyClusterScanConfig(d.Id(), policy)
	}
	return nil
}

func toProfiles(d *schema.ResourceData) []*models.V1SpectroClusterProfileEntity {
	resp := make([]*models.V1SpectroClusterProfileEntity, 0)
	profiles := d.Get("cluster_profile").([]interface{})
	if len(profiles) > 0 {
		for _, profile := range profiles {
			p := profile.(map[string]interface{})

			packValues := make([]*models.V1PackValuesEntity, 0)
			for _, pack := range p["pack"].([]interface{}) {
				p := toPack(pack)
				packValues = append(packValues, p)
			}
			resp = append(resp, &models.V1SpectroClusterProfileEntity{
				UID:        p["id"].(string),
				PackValues: packValues,
			})
		}
	} else {
		packValues := make([]*models.V1PackValuesEntity, 0)
		for _, pack := range d.Get("pack").([]interface{}) {
			p := toPack(pack)
			packValues = append(packValues, p)
		}
		resp = append(resp, &models.V1SpectroClusterProfileEntity{
			UID:        d.Get("cluster_profile_id").(string),
			PackValues: packValues,
		})
	}

	return resp
}

func toPack(pSrc interface{}) *models.V1PackValuesEntity {
	p := pSrc.(map[string]interface{})

	pack := &models.V1PackValuesEntity{
		Name: ptr.StringPtr(p["name"].(string)),
	}

	if val, found := p["values"]; found && len(val.(string)) > 0 {
		pack.Values = val.(string)
	}
	if val, found := p["tag"]; found && len(val.(string)) > 0 {
		pack.Tag = val.(string)
	}
	if val, found := p["type"]; found && len(val.(string)) > 0 {
		pack.Type = models.V1PackType(val.(string))
	}
	if val, found := p["manifest"]; found && len(val.([]interface{})) > 0 {
		manifestsData := val.([]interface{})
		manifests := make([]*models.V1ManifestRefUpdateEntity, len(manifestsData))
		for i := 0; i < len(manifestsData); i++ {
			data := manifestsData[i].(map[string]interface{})
			manifests[i] = &models.V1ManifestRefUpdateEntity{
				Name:    ptr.StringPtr(data["name"].(string)),
				Content: data["content"].(string),
			}
		}
		pack.Manifests = manifests
	}

	return pack
}

func resourceClusterStateRefreshFunc(c *client.V1Client, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := c.GetCluster(id)
		if err != nil {
			return nil, "", err
		} else if cluster == nil {
			return nil, "Deleted", nil
		}

		state := cluster.Status.State
		log.Printf("Cluster state (%s): %s", id, state)

		return cluster, state, nil
	}
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	err := c.DeleteCluster(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := waitForClusterDeletion(ctx, c, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceMachinePoolAzureHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	//d := m["disk"].([]interface{})[0].(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["azs"].(*schema.Set).GoString()))

	// TODO(saamalik) fix for disk
	//buf.WriteString(fmt.Sprintf("%d-", d["size_gb"].(int)))
	//buf.WriteString(fmt.Sprintf("%s-", d["type"].(string)))

	//d2 := m["disk"].([]interface{})
	//d := d2[0].(map[string]interface{})

	return int(hash(buf.String()))
}

func resourceMachinePoolAksHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["disk_size_gb"].(int)))
	buf.WriteString(fmt.Sprintf("%t-", m["is_system_node_pool"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["storage_account_type"].(string)))
	return int(hash(buf.String()))
}

func resourceMachinePoolGcpHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	//d := m["disk"].([]interface{})[0].(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["azs"].(*schema.Set).GoString()))

	return int(hash(buf.String()))
}

func resourceMachinePoolAwsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["capacity_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["max_price"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["azs"].(*schema.Set).GoString()))

	return int(hash(buf.String()))
}

func resourceMachinePoolEksHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["disk_size_gb"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["capacity_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["max_price"].(string)))

	for i, j := range m["az_subnets"].(map[string]interface{}) {
		buf.WriteString(fmt.Sprintf("%s-%s", i, j.(string)))
	}
	// TODO

	return int(hash(buf.String()))
}

func resourceMachinePoolVsphereHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	//d := m["disk"].([]interface{})[0].(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))

	if v, found := m["instance_type"]; found {
		ins := v.([]interface{})[0].(map[string]interface{})
		buf.WriteString(fmt.Sprintf("%d-", ins["cpu"].(int)))
		buf.WriteString(fmt.Sprintf("%d-", ins["disk_size_gb"].(int)))
		buf.WriteString(fmt.Sprintf("%d-", ins["memory_mb"].(int)))
	}

	return int(hash(buf.String()))
}

func resourceMachinePoolOpenStackHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))

	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["subnet_id"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["update_strategy"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["azs"].(*schema.Set).GoString()))

	return int(hash(buf.String()))
}

func resourceMachinePoolMaasHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	//d := m["disk"].([]interface{})[0].(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	if v, found := m["instance_type"]; found {
		ins := v.([]interface{})[0].(map[string]interface{})
		buf.WriteString(fmt.Sprintf("%d-", ins["min_cpu"].(int)))
		buf.WriteString(fmt.Sprintf("%d-", ins["min_memory_mb"].(int)))
	}
	buf.WriteString(fmt.Sprintf("%s-", m["azs"].(*schema.Set).GoString()))

	return int(hash(buf.String()))
}

func hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

func toClusterConfig(d *schema.ResourceData) *models.V1ClusterConfigEntity {
	return &models.V1ClusterConfigEntity{
		MachineManagementConfig: toMachineManagementConfig(d),
	}
}

func toMachineManagementConfig(d *schema.ResourceData) *models.V1MachineManagementConfig {
	return &models.V1MachineManagementConfig{
		OsPatchConfig: toOsPatchConfig(d),
	}
}

func toOsPatchConfig(d *schema.ResourceData) *models.V1OsPatchConfig {
	osPatchOnBoot := d.Get("os_patch_on_boot").(bool)
	osPatchOnSchedule := d.Get("os_patch_schedule").(string)
	osPatchAfter := d.Get("os_patch_after").(string)
	if osPatchOnBoot || len(osPatchOnSchedule) > 0 || len(osPatchAfter) > 0 {
		osPatchConfig := &models.V1OsPatchConfig{}
		if osPatchOnBoot {
			osPatchConfig.PatchOnBoot = osPatchOnBoot
		}
		if len(osPatchOnSchedule) > 0 {
			osPatchConfig.Schedule = osPatchOnSchedule
		}
		if len(osPatchAfter) > 0 {
			patchAfter, _ := time.Parse(time.RFC3339, osPatchAfter)
			osPatchConfig.OnDemandPatchAfter = models.V1Time(patchAfter)
		} else {
			//setting Zero time in request
			zeroTime, _ := time.Parse(time.RFC3339, "0001-01-01T00:00:00.000Z")
			osPatchConfig.OnDemandPatchAfter = models.V1Time(zeroTime)
		}
		return osPatchConfig
	}
	return nil
}

func validateOsPatchSchedule(data interface{}, _ cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	if data != nil {
		if _, err := cron.ParseStandard(data.(string)); err != nil {
			return diag.FromErr(errors.Wrap(err, "os patch schedule is invalid. Please see https://en.wikipedia.org/wiki/Cron for valid cron syntax"))
		}
	}
	return diags
}

func validateOsPatchOnDemandAfter(data interface{}, _ cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	if data != nil {
		if patchTime, err := time.Parse(time.RFC3339, data.(string)); err != nil {
			return diag.FromErr(errors.Wrap(err, "time for 'os_patch_after' is invalid. Please follow RFC3339 Date and Time Standards. Eg 2021-01-01T00:00:00.000Z "))
		} else {
			if time.Now().After(patchTime.Add(10 * time.Minute)) {
				return diag.FromErr(fmt.Errorf("valid timestamp is timestamp which is 10 mins ahead of current timestamp. Eg any timestamp ahead of %v", time.Now().Add(10*time.Minute).Format(time.RFC3339)))
			}
		}
	}

	return diags
}
