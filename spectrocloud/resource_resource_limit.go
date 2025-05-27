package spectrocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"time"
)

func resourceResourceLimit() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceResourceLimitsCreate,
		ReadContext:   resourceResourceLimitsRead,
		UpdateContext: resourceResourceLimitsUpdate,
		DeleteContext: resourceResourceLimitsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceResourceLimitsImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Description:   "",
		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"alert": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The maximum number of alerts that can be created. Must be between 1 and 10,000.",
			},
			"api_keys": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The maximum number of API keys that can be generated. Must be between 1 and 10,000.",
			},
			"application_deployment": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The maximum number of application deployments allowed. Must be between 1 and 10,000.",
			},
			"application_profile": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The maximum number of application profiles that can be configured. Must be between 1 and 10,000.",
			},
			"certificate": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The maximum number of certificates that can be managed. Must be between 1 and 10,000.",
			},
			"cloud_account": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      200,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The maximum number of cloud accounts that can be added. Must be between 1 and 10,000.",
			},
			"cluster_group": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The maximum number of cluster groups that can be created. Must be between 1 and 10,000.",
			},
			"cluster_profile": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      200,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The maximum number of cluster profiles that can be configured. Must be between 1 and 10,000.",
			},
			"appliance": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      200,
				ValidateFunc: validation.IntBetween(1, 50000),
				Description:  "The maximum number of appliances that can be managed. Must be between 1 and 50,000.",
			},
			"appliance_token": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      200,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The maximum number of appliance tokens that can be issued. Must be between 1 and 10,000.",
			},
			"filter": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The maximum number of filters that can be defined. Must be between 1 and 10,000.",
			},
			"location": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The maximum number of locations that can be configured. Must be between 1 and 10,000.",
			},
			"macro": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      200,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The maximum number of macros that can be created. Must be between 1 and 10,000.",
			},
			"private_gateway": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      50,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The maximum number of private gateways that can be managed. Must be between 1 and 10,000.",
			},
			"project": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      50,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The maximum number of projects that can be created. Must be between 1 and 10,000.",
			},
			"registry": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      50,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The maximum number of registries that can be configured. Must be between 1 and 10,000.",
			},
			"role": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The maximum number of roles that can be assigned. Must be between 1 and 10,000.",
			},
			"cluster": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      10000,
				ValidateFunc: validation.IntBetween(1, 50000),
				Description:  "The maximum number of clusters that can be created. Must be between 1 and 50,000.",
			},
			"ssh_key": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      300,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The maximum number of SSH keys that can be managed. Must be between 1 and 10,000.",
			},
			"team": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The maximum number of teams that can be created. Must be between 1 and 10,000.",
			},
			"user": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      300,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The maximum number of users that can be added. Must be between 1 and 10,000.",
			},
			"workspace": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      50,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The maximum number of workspaces that can be created. Must be between 1 and 10,000.",
			},
		},
	}
}

var KindToFieldMapping = []struct {
	Kind    models.V1ResourceLimitType
	Field   string
	Default int64
}{
	{models.V1ResourceLimitTypeAlert, "alert", 100},
	{models.V1ResourceLimitTypeAPIKey, "api_keys", 20},
	{models.V1ResourceLimitTypeAppdeployment, "application_deployment", 100},
	{models.V1ResourceLimitTypeAppprofile, "application_profile", 100},
	{models.V1ResourceLimitTypeCertificate, "certificate", 20},
	{models.V1ResourceLimitTypeCloudaccount, "cloud_account", 200},
	{models.V1ResourceLimitTypeClustergroup, "cluster_group", 100},
	{models.V1ResourceLimitTypeClusterprofile, "cluster_profile", 200},
	{models.V1ResourceLimitTypeEdgehost, "appliance", 200},
	{models.V1ResourceLimitTypeEdgetoken, "appliance_token", 200},
	{models.V1ResourceLimitTypeFilter, "filter", 100},
	{models.V1ResourceLimitTypeLocation, "location", 100},
	{models.V1ResourceLimitTypeMacro, "macro", 200},
	{models.V1ResourceLimitTypePrivategateway, "private_gateway", 50},
	{models.V1ResourceLimitTypeProject, "project", 50},
	{models.V1ResourceLimitTypeRegistry, "registry", 50},
	{models.V1ResourceLimitTypeRole, "role", 400},
	{models.V1ResourceLimitTypeSpectrocluster, "cluster", 1000},
	{models.V1ResourceLimitTypeSshkey, "ssh_key", 300},
	{models.V1ResourceLimitTypeTeam, "team", 100},
	{models.V1ResourceLimitTypeUser, "user", 300},
	{models.V1ResourceLimitTypeWorkspace, "workspace", 50},
}

func toResourceLimits(d *schema.ResourceData) (*models.V1TenantResourceLimitsEntity, error) {

	resourceLimit := make([]*models.V1TenantResourceLimitEntity, len(KindToFieldMapping))
	for i, mapping := range KindToFieldMapping {
		resourceLimit[i] = &models.V1TenantResourceLimitEntity{
			Kind:  mapping.Kind,
			Limit: int64(d.Get(mapping.Field).(int)),
		}
	}

	return &models.V1TenantResourceLimitsEntity{Resources: resourceLimit}, nil
}

func toResourceDefaultLimits(d *schema.ResourceData) (*models.V1TenantResourceLimitsEntity, error) {

	resourceLimit := make([]*models.V1TenantResourceLimitEntity, len(KindToFieldMapping))
	for i, limit := range KindToFieldMapping {
		resourceLimit[i] = &models.V1TenantResourceLimitEntity{
			Kind:  limit.Kind,
			Limit: limit.Default,
		}
	}

	return &models.V1TenantResourceLimitsEntity{Resources: resourceLimit}, nil
}

func flattenResourceLimits(resourceLimits *models.V1TenantResourceLimits, d *schema.ResourceData) error {
	kindToField := make(map[models.V1ResourceLimitType]string, len(KindToFieldMapping))
	for _, mapping := range KindToFieldMapping {
		kindToField[mapping.Kind] = mapping.Field
	}

	for _, resource := range resourceLimits.Resources {
		if field, exists := kindToField[resource.Kind]; exists {
			if err := d.Set(field, resource.Limit); err != nil {
				return err
			}
		}
	}

	return nil
}

func resourceResourceLimitsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	resourceLimits, err := toResourceLimits(d)
	if err != nil {
		return diag.FromErr(err)
	}
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.UpdateResourceLimits(tenantUID, resourceLimits)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("default-resource-limit-id")
	return diags
}

func resourceResourceLimitsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return handleReadError(d, err, diags)
	}
	resp, err := c.GetResourceLimits(tenantUID)
	if err != nil {
		return handleReadError(d, err, diags)
	}
	err = flattenResourceLimits(resp, d)
	if err != nil {
		return nil
	}
	return diags
}

func resourceResourceLimitsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	resourceLimits, err := toResourceLimits(d)
	if err != nil {
		return diag.FromErr(err)
	}
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.UpdateResourceLimits(tenantUID, resourceLimits)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceResourceLimitsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	// We can't delete the base resource limit, instead
	resourceLimits, err := toResourceDefaultLimits(d)
	if err != nil {
		return diag.FromErr(err)
	}
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.UpdateResourceLimits(tenantUID, resourceLimits)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}

func resourceResourceLimitsImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics
	givenTenantId := d.Id()
	actualTenantId, err := c.GetTenantUID()
	if err != nil {
		return nil, err
	}
	if givenTenantId != actualTenantId {
		return nil, fmt.Errorf("tenant id is not valid with curent user: %v", diags)
	}
	diags = resourceResourceLimitsRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read resource limits for import: %v", diags)
	}
	return []*schema.ResourceData{d}, nil
}
