package spectrocloud

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"log"
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

func updateProfiles(c *client.V1alpha1Client, d *schema.ResourceData) error {
	log.Printf("Updating profiles")
	body := &models.V1alpha1SpectroClusterProfiles{
		Profiles: toProfiles(d),
	}
	if err := c.UpdateClusterProfileValues(d.Id(), body); err != nil {
		return err
	}
	return nil
}

func toProfiles(d *schema.ResourceData) []*models.V1alpha1SpectroClusterProfileEntity {
	resp := make([]*models.V1alpha1SpectroClusterProfileEntity, 0)
	profiles := d.Get("cluster_profile").(*schema.Set).List()
	if len(profiles) > 0 {
		for _, profile := range profiles {
			p := profile.(map[string]interface{})

			packValues := make([]*models.V1alpha1PackValuesEntity, 0)
			for _, pack := range p["pack"].([]interface{}) {
				p := toPack(pack)
				packValues = append(packValues, p)
			}
			resp = append(resp, &models.V1alpha1SpectroClusterProfileEntity{
				UID:        p["cluster_profile_id"].(string),
				PackValues: packValues,
			})
		}
	} else {
		packValues := make([]*models.V1alpha1PackValuesEntity, 0)
		for _, pack := range d.Get("pack").([]interface{}) {
			p := toPack(pack)
			packValues = append(packValues, p)
		}
		resp = append(resp, &models.V1alpha1SpectroClusterProfileEntity{
			UID:        d.Get("cluster_profile_id").(string),
			PackValues: packValues,
		})
	}

	return resp
}

func toPack(pSrc interface{}) *models.V1alpha1PackValuesEntity {
	p := pSrc.(map[string]interface{})

	pack := &models.V1alpha1PackValuesEntity{
		Name:   ptr.StringPtr(p["name"].(string)),
		Tag:    p["tag"].(string),
		Values: p["values"].(string),
	}
	return pack
}

func resourceClusterStateRefreshFunc(c *client.V1alpha1Client, id string) resource.StateRefreshFunc {
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

func resourceClusterProfileHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["cluster_profile_id"].(string)))
	return int(hash(buf.String()))
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
	//d := m["disk"].([]interface{})[0].(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
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

	for i, j := range m["az_subnets"].(map[string]interface{}) {
		buf.WriteString(fmt.Sprintf("%s-%s", i, j.(string)))
	}
	//buf.WriteString(fmt.Sprintf("%s-", m["az_subnets"].(*schema.Map).GoString()))
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

func hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

func toClusterConfig(d *schema.ResourceData) *models.V1alpha1ClusterConfig {
	return &models.V1alpha1ClusterConfig{
		MachineManagementConfig: toMachineManagementConfig(d),
	}
}

func toMachineManagementConfig(d *schema.ResourceData) *models.V1alpha1MachineManagementConfig {
	return &models.V1alpha1MachineManagementConfig{
		OsPatchConfig: toOsPatchConfig(d),
	}
}

func toOsPatchConfig(d *schema.ResourceData) *models.V1alpha1OsPatchConfig {
	osPatchOnBoot := d.Get("os_patch_on_boot").(bool)
	osPatchOnSchedule := d.Get("os_patch_schedule").(string)
	osPatchAfter := d.Get("os_patch_after").(string)
	if osPatchOnBoot || len(osPatchOnSchedule) > 0 || len(osPatchAfter) > 0 {
		osPatchConfig := &models.V1alpha1OsPatchConfig{}
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

