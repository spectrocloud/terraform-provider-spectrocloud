package spectrocloud

import (
	"bytes"
	"context"
	"fmt"
	"github.com/robfig/cron"
	"hash/fnv"
	"log"
	"time"

	"github.com/go-openapi/strfmt"

	"github.com/pkg/errors"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

var (
	DefaultDiskType = "Standard_LRS"
	DefaultDiskSize = 60
)

var resourceClusterDeletePendingStates = []string{
	string(pending),
	string(provisioning),
	string(running),
	string(deleting),
	string(importing),
}
var resourceClusterCreatePendingStates = []string{
	string(pending),
	string(provisioning),
	string(importing),
}

//var resourceClusterUpdatePendingStates = []string{
//	"backing-up",
//	"modifying",
//	"resetting-master-credentials",
//	"upgrading",
//}
func waitForClusterDeletion(ctx context.Context, c *client.V1alpha1Client, id string, timeout time.Duration) error {
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

func toPack(pSrc interface{}) *models.V1alpha1PackValuesEntity {
	p := pSrc.(map[string]interface{})

	pack := &models.V1alpha1PackValuesEntity{
		Name:   ptr.StringPtr(p[name].(string)),
		Tag:    ptr.StringPtr(p[tag].(string)),
		Values: p[values].(string),
	}
	return pack
}

func resourceClusterStateRefreshFunc(c *client.V1alpha1Client, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := c.GetCluster(id)
		if err != nil {
			return nil, "", err
		} else if cluster == nil {
			return nil, string(deleted), nil
		}

		state := cluster.Status.State
		log.Printf("Cluster state (%s): %s", id, state)

		return cluster, state, nil
	}
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

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

func cloudClusterImportFunc(c *client.V1alpha1Client, d *schema.ResourceData) (string, error) {
	meta := toClusterMeta(d)
	cloudType := d.Get(cloud).(string)
	switch CloudType(cloudType) {
	case cloud_type_aws:
		return c.ImportClusterAws(meta)
	case cloud_type_azure:
		return c.ImportClusterAzure(meta)
	case cloud_type_gcp:
		return c.ImportClusterGcp(meta)
	case cloud_type_vsphere:
		return c.ImportClusterVsphere(meta)
	}
	return "", fmt.Errorf("failed to find cloud type %s", cloudType)
}

func cloudClusterReadFunc(ctx context.Context, d *schema.ResourceData, m interface{}) {
	cloudType := d.Get(cloud).(string)
	switch CloudType(cloudType) {
	case cloud_type_aws:
		resourceClusterAwsRead(ctx, d, m)
	case cloud_type_azure:
		resourceClusterAzureRead(ctx, d, m)
	case cloud_type_gcp:
		resourceClusterGcpRead(ctx, d, m)
	case cloud_type_vsphere:
		resourceClusterVsphereRead(ctx, d, m)
	}
}

func resourceCloudClusterImport(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)
	var diags diag.Diagnostics
	uid, err := cloudClusterImportFunc(c, d)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)
	stateConf := &resource.StateChangeConf{
		Target:     []string{string(pending)},
		Refresh:    resourceClusterStateRefreshFunc(c, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate) - 1*time.Minute,
		MinTimeout: 1 * time.Second,
		Delay:      5 * time.Second,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	cloudClusterReadFunc(ctx, d, m)

	if profiles := getCloudClusterProfiles(d); profiles != nil {
		if err := c.UpdateBrownfieldCluster(uid, profiles); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func resourceCloudClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cloudType := d.Get(cloud).(string)
	switch CloudType(cloudType) {
	case cloud_type_aws:
		return resourceClusterAwsRead(ctx, d, m)
	case cloud_type_azure:
		return resourceClusterAzureRead(ctx, d, m)
	case cloud_type_gcp:
		return resourceClusterGcpRead(ctx, d, m)
	case cloud_type_vsphere:
		return resourceClusterVsphereRead(ctx, d, m)
	}
	return diag.FromErr(fmt.Errorf("failed to import cluster as cloud type '%s' is invalid", cloudType))
}

func resourceCloudClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)
	var diags diag.Diagnostics

	clusterProfileId := d.Get(cluster_prrofile_id).(string)
	profiles := make([]*models.V1alpha1SpectroClusterProfileEntity, 0)
	packValues := make([]*models.V1alpha1PackValuesEntity, 0)
	for _, pack := range d.Get(pack).(*schema.Set).List() {
		p := toPack(pack)
		packValues = append(packValues, p)
	}

	profiles = append(profiles, &models.V1alpha1SpectroClusterProfileEntity{
		PackValues: packValues,
		UID:        clusterProfileId,
	})

	err := c.UpdateBrownfieldCluster("", &models.V1alpha1SpectroClusterProfiles{
		Profiles: profiles,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourcePackHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(fmt.Sprintf("%s-", m[name].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m[tag].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m[values].(string)))

	return int(hash(buf.String()))
}

func resourceMachinePoolAzureHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	//d := m[disk].([]interface{})[0].(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%t-", m[control_plane].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m[control_plane_as_worker].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m[name].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m[count].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m[instance_type].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m[azs].(*schema.Set).GoString()))

	// TODO(saamalik) fix for disk
	//buf.WriteString(fmt.Sprintf("%d-", d[size_gb].(int)))
	//buf.WriteString(fmt.Sprintf("%s-", d[disk_type].(string)))

	//d2 := m[disk].([]interface{})
	//d := d2[0].(map[string]interface{})

	return int(hash(buf.String()))
}

func resourceMachinePoolGcpHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	//d := m[disk].([]interface{})[0].(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%t-", m[control_plane].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m[control_plane_as_worker].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m[name].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m[count].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m[instance_type].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m[azs].(*schema.Set).GoString()))

	return int(hash(buf.String()))
}

func resourceMachinePoolAwsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	//d := m[disk].([]interface{})[0].(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%t-", m[control_plane].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m[control_plane_as_worker].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m[name].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m[count].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m[instance_type].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m[azs].(*schema.Set).GoString()))

	return int(hash(buf.String()))
}

func resourceMachinePoolVsphereHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	//d := m[disk].([]interface{})[0].(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%t-", m[control_plane].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m[control_plane_as_worker].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m[name].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m[count].(int)))

	// TODO(saamalik) MORE

	// TODO(saamalik) fix for disk
	//buf.WriteString(fmt.Sprintf("%d-", d[size_gb].(int)))
	//buf.WriteString(fmt.Sprintf("%s-", d[disk_type].(string)))

	//d2 := m[disk].([]interface{})
	//d := d2[0].(map[string]interface{})

	return int(hash(buf.String()))
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func toClusterMeta(d *schema.ResourceData) *models.V1ObjectMetaInputEntity {
	return &models.V1ObjectMetaInputEntity{
		Name: d.Get(name).(string),
	}
}

func getCloudClusterProfiles(d *schema.ResourceData) *models.V1alpha1SpectroClusterProfiles {
	if clusterProfileUid := d.Get(cluster_prrofile_id); clusterProfileUid != nil {
		profileEntities := make([]*models.V1alpha1SpectroClusterProfileEntity, 0)
		packValues := make([]*models.V1alpha1PackValuesEntity, 0)
		for _, pack := range d.Get(pack).(*schema.Set).List() {
			p := toPack(pack)
			packValues = append(packValues, p)
		}

		profileEntities = append(profileEntities, &models.V1alpha1SpectroClusterProfileEntity{
			PackValues: packValues,
			UID:        clusterProfileUid.(string),
		})
		return &models.V1alpha1SpectroClusterProfiles{
			Profiles: profileEntities,
		}
	}
	return nil
}

func getClusterConfig(d *schema.ResourceData) *models.V1alpha1ClusterConfig {
	return &models.V1alpha1ClusterConfig{
		MachineManagementConfig: getMachineManagementConfig(d),
	}
}

func getMachineManagementConfig(d *schema.ResourceData) *models.V1alpha1MachineManagementConfig {
	return &models.V1alpha1MachineManagementConfig{
		OsPatchConfig: getOsPatchConfig(d),
	}
}

func getOsPatchConfig(d *schema.ResourceData) *models.V1alpha1OsPatchConfig {
	osPatchOnBoot := d.Get(os_patch_on_boot).(bool)
	osPatchOnSchedule := d.Get(os_patch_schedule).(string)
	osPatchAfter := d.Get(os_patch_after).(string)
	if osPatchOnBoot || len(osPatchOnSchedule) > 0 || len(osPatchAfter) > 0{
		osPatchConfig := &models.V1alpha1OsPatchConfig{}
		if osPatchOnBoot {
			osPatchConfig.PatchOnBoot = osPatchOnBoot
		}
		if len(osPatchOnSchedule) > 0 {
			osPatchConfig.Schedule = osPatchOnSchedule
		}
		if len(osPatchAfter) > 0 {
			dateTime, _ := strfmt.ParseDateTime(osPatchAfter)
			osPatchConfig.OnDemandPatchAfter = dateTime
		} else {
			//setting Zero time in request
			zeroTime, _ := strfmt.ParseDateTime("0001-01-01T00:00:00.000Z")
			osPatchConfig.OnDemandPatchAfter = zeroTime
		}
		return osPatchConfig
	}
	return nil
}

func getSpectroClusterProfiles(d *schema.ResourceData) []*models.V1alpha1SpectroClusterProfileEntity {
	profiles := make([]*models.V1alpha1SpectroClusterProfileEntity, 0)
	packValues := make([]*models.V1alpha1PackValuesEntity, 0)
	for _, pack := range d.Get(pack).(*schema.Set).List() {
		p := toPack(pack)
		packValues = append(packValues, p)
	}
	profile := &models.V1alpha1SpectroClusterProfileEntity{
		UID: d.Get("cluster_profile_id").(string),
	}
	if len(packValues) > 0 {
		profile.PackValues = packValues
	}
	profiles = append(profiles, profile)
	return profiles
}

func validateCloudType(data interface{}, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	inCloudType := data.(string)
	for _, cloudType := range cloud_types {
		if cloudType == inCloudType {
			return diags
		}
	}
	return diag.FromErr(fmt.Errorf("cloud type '%s' is invalid. valid cloud types are %v", inCloudType, cloud_types))
}

func validateOsPatchSchedule(data interface{}, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	if data != nil {
		if _, err := cron.ParseStandard(data.(string)); err != nil {
			return diag.FromErr(errors.Wrap(err, "os patch schedule is invalid. Please see https://en.wikipedia.org/wiki/Cron for valid cron syntax"))
		}
	}
	return diags
}

func validateOsPatchOnDemandAfter(data interface{}, path cty.Path) diag.Diagnostics {
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
