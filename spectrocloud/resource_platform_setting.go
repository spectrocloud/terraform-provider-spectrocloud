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

func resourcePlatformSetting() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePlatformSettingCreate,
		ReadContext:   resourcePlatformSettingRead,
		UpdateContext: resourcePlatformSettingUpdate,
		DeleteContext: resourcePlatformSettingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcePlatformSettingImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "tenant",
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"project", "tenant"}, false),
				Description: "Defines the scope of the platform setting. Valid values are `project` or `tenant`. " +
					"By default, it is set to `tenant`. " + PROJECT_NAME_NUANCE,
			},
			"session_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Specifies the duration (in minutes) of inactivity before a user is automatically logged out. The default is 240 minutes allowed in Palette",
			},
			"pause_agent_upgrades": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "unlock",
				ValidateFunc: validation.StringInSlice([]string{"lock", "unlock"}, false),
				Description: "Controls automatic upgrades for Palette components and agents in clusters deployed under a tenant or project. " +
					"Setting it to `lock` disables automatic upgrades, while `unlock` (default) allows automatic upgrades.",
			},
			"enable_auto_remediation": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Enables automatic remediation. set only with `project' context",
			},
			"cluster_auto_remediation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				Description: "Enables automatic remediation for unhealthy nodes in Palette-provisioned clusters by replacing them with new nodes. " +
					"Disabling this feature prevents auto-remediation. Not applicable to `EKS`, `AKS`, or `TKE` clusters.",
			},
			"non_fips_addon_pack": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Allows users in this tenant to use non-FIPS-compliant addon packs when creating cluster profiles. The `non_fips_addon_pack` only supported in palette vertex environment.",
			},
			"non_fips_features": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Allows users in this tenant to access non-FIPS-compliant features such as backup, restore, and scans. The `non_fips_features` only supported in palette vertex environment.",
			},
			"non_fips_cluster_import": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Allows users in this tenant to import clusters, but the imported clusters may not be FIPS-compliant.  The `non_fips_cluster_import` only supported in palette vertex environment.",
			},
			"login_banner": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Configure a login banner that users must acknowledge before signing in.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"title": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Specify the title of the login banner.",
						},
						"message": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Specify the message displayed in the login banner.",
						},
					},
				},
			},
		},
		CustomizeDiff: validateContextDependencies,
	}
}

func validateContextDependencies(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	contextVal := d.Get("context").(string)

	if contextVal == "project" {
		disallowedFields := []string{"session_timeout", "login_banner", "non_fips_addon_pack", "non_fips_features", "non_fips_cluster_import"}

		for _, field := range disallowedFields {
			if _, exists := d.GetOk(field); exists {
				return fmt.Errorf("attribute %q is not allowed when context is set to 'project'", field)
			}
		}
	}
	return nil
}

func updatePlatformSettings(d *schema.ResourceData, m interface{}) diag.Diagnostics {
	platformSettingContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, platformSettingContext)
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	var diags diag.Diagnostics

	remediationSettings := &models.V1NodesAutoRemediationSettings{
		DisableNodesAutoRemediation: d.Get("cluster_auto_remediation").(bool),
		IsEnabled:                   d.Get("enable_auto_remediation").(bool), // when ever we are setting `cluster_auto_remediation` we need enable it hence set same attribute
	}
	if platformSettingContext == tenantString {
		// session timeout
		if sessionTime, ok := d.GetOk("session_timeout"); ok {
			err = c.UpdateSessionTimeout(tenantUID,
				&models.V1AuthTokenSettings{ExpiryTimeMinutes: int32(sessionTime.(int))})
			if err != nil {
				return diag.FromErr(err)
			}
		}

		loginBanner := d.Get("login_banner").([]interface{})
		// login banner
		if len(loginBanner) == 1 {
			bannerData := loginBanner[0].(map[string]interface{})
			bannerSetting := &models.V1LoginBannerSettings{
				Message:   bannerData["message"].(string),
				IsEnabled: true,
				Title:     bannerData["title"].(string),
			}
			err = c.UpdateLoginBanner(tenantUID, bannerSetting)
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			bannerSetting := &models.V1LoginBannerSettings{
				Message:   "",
				IsEnabled: false,
				Title:     "",
			}
			err = c.UpdateLoginBanner(tenantUID, bannerSetting)
			if err != nil {
				return diag.FromErr(err)
			}
		}
		// cluster node remediation for tenant
		err = c.UpdateClusterAutoRemediationForTenant(tenantUID, remediationSettings)
		if err != nil {
			return diag.FromErr(err)
		}

		// non fip related setting
		fipsAddonPack := "nonFipsDisabled"
		fipsFeatures := "nonFipsDisabled"
		fipsClusterImport := "nonFipsDisabled"

		fp, fpOk := d.GetOk("non_fips_addon_pack")
		ff, ffOk := d.GetOk("non_fips_features")
		fi, fiOk := d.GetOk("non_fips_cluster_import")

		if fpOk {
			fipsAddonPack = convertFIPSBool(fp.(bool))
		}
		if ffOk {
			fipsFeatures = convertFIPSBool(ff.(bool))
		}
		if fiOk {
			fipsClusterImport = convertFIPSBool(fi.(bool))
		}

		if fiOk || ffOk || fpOk {
			err = c.UpdateFIPSPreference(tenantUID, &models.V1FipsSettings{
				FipsClusterFeatureConfig: &models.V1NonFipsConfig{Mode: &fipsFeatures},
				FipsClusterImportConfig:  &models.V1NonFipsConfig{Mode: &fipsClusterImport},
				FipsPackConfig:           &models.V1NonFipsConfig{Mode: &fipsAddonPack},
			})
			if err != nil {
				return diag.FromErr(err)
			}
		}

	} else {
		// cluster node remediation for project
		err = c.UpdateClusterAutoRemediationForProject(ProviderInitProjectUid, remediationSettings)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	// pause agent upgrade setting according to context
	err = c.UpdatePlatformClusterUpgradeSetting(&models.V1ClusterUpgradeSettingsEntity{
		SpectroComponents: d.Get("pause_agent_upgrades").(string)})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func convertFIPSBool(flag bool) string {
	if flag {
		return "nonFipsEnabled"
	}
	return "nonFipsDisabled"
}

func convertFIPSString(flag string) bool {
	return flag == "nonFipsEnabled"
}

func resourcePlatformSettingCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	platformSettingContext := d.Get("context").(string)
	diags = updatePlatformSettings(d, m)
	d.SetId(fmt.Sprintf("default-platform-setting-%s", platformSettingContext))
	return diags
}

func resourcePlatformSettingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	platformSettingContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, platformSettingContext)
	var diags diag.Diagnostics
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return handleReadError(d, err, diags)
	}
	// handling case for cross-plane for singleton resource
	if d.Id() != fmt.Sprintf("default-platform-setting-%s", platformSettingContext) {
		d.SetId("")
		return diags
	}

	if platformSettingContext == tenantString {
		// read session timeout
		var respSessionTimeout *models.V1AuthTokenSettings
		respSessionTimeout, err = c.GetSessionTimeout(tenantUID)
		if err != nil {
			return handleReadError(d, err, diags)
		}
		if err = d.Set("session_timeout", respSessionTimeout.ExpiryTimeMinutes); err != nil {
			return diag.FromErr(err)
		}
		// read login banner
		var respLoginBanner *models.V1LoginBannerSettings
		respLoginBanner, err = c.GetLoginBanner(tenantUID)
		if err != nil {
			return handleReadError(d, err, diags)
		}
		if respLoginBanner.Title != "" && respLoginBanner.Message != "" {
			bannerDetails := make([]interface{}, 0)
			bd := map[string]string{
				"title":   respLoginBanner.Title,
				"message": respLoginBanner.Message,
			}
			bannerDetails = append(bannerDetails, bd)
			if err = d.Set("login_banner", bannerDetails); err != nil {
				return diag.FromErr(err)
			}
		}
		// get cluster_auto_remediation tenant
		var respRemediation *models.V1TenantClusterSettings
		respRemediation, err = c.GetClusterAutoRemediationForTenant(tenantUID)
		if err != nil {
			return handleReadError(d, err, diags)
		}
		if err = d.Set("cluster_auto_remediation", respRemediation.NodesAutoRemediationSetting.DisableNodesAutoRemediation); err != nil {
			return diag.FromErr(err)
		}
		// get fips settings
		var fipsPreference *models.V1FipsSettings

		_, fpOk := d.GetOk("non_fips_addon_pack")
		_, ffOk := d.GetOk("non_fips_features")
		_, fiOk := d.GetOk("non_fips_cluster_import")

		if fiOk || ffOk || fpOk {
			fipsPreference, err = c.GetFIPSPreference(tenantUID)
			if err != nil {
				return handleReadError(d, err, diags)
			}
			if _, ok := d.GetOk("non_fips_addon_pack"); ok {
				err := d.Set("non_fips_addon_pack", convertFIPSString(*fipsPreference.FipsPackConfig.Mode))
				if err != nil {
					return diag.FromErr(err)
				}
			}
			if _, ok := d.GetOk("non_fips_features"); ok {
				err := d.Set("non_fips_features", convertFIPSString(*fipsPreference.FipsClusterFeatureConfig.Mode))
				if err != nil {
					return nil
				}
			}
			if _, ok := d.GetOk("non_fips_cluster_import"); ok {
				err := d.Set("non_fips_cluster_import", convertFIPSString(*fipsPreference.FipsClusterImportConfig.Mode))
				if err != nil {
					return nil
				}
			}
		}

	} else {
		// get cluster_auto_remediation project
		var respProjectRemediation *models.V1ProjectClusterSettings
		respProjectRemediation, err = c.GetClusterAutoRemediationForProject(ProviderInitProjectUid)
		if err != nil {
			return handleReadError(d, err, diags)
		}
		if err = d.Set("cluster_auto_remediation", respProjectRemediation.NodesAutoRemediationSetting.DisableNodesAutoRemediation); err != nil {
			return diag.FromErr(err)
		}
		if err = d.Set("enable_auto_remediation", respProjectRemediation.NodesAutoRemediationSetting.IsEnabled); err != nil {
			return diag.FromErr(err)
		}
	}
	// pause agent upgrade setting according to context
	var upgradeSetting *models.V1ClusterUpgradeSettingsEntity
	upgradeSetting, err = c.GetPlatformClustersUpgradeSetting()
	if err != nil {
		return handleReadError(d, err, diags)
	}
	if err = d.Set("pause_agent_upgrades", upgradeSetting.SpectroComponents); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourcePlatformSettingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	platformSettingContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, platformSettingContext)
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	var diags diag.Diagnostics

	remediationSettings := &models.V1NodesAutoRemediationSettings{
		DisableNodesAutoRemediation: d.Get("cluster_auto_remediation").(bool),
		IsEnabled:                   d.Get("enable_auto_remediation").(bool), // when ever we are setting `cluster_auto_remediation` we need enable it hence set same attribute
	}
	if platformSettingContext == tenantString {
		// session timeout
		if d.HasChange("session_timeout") {
			if sessionTime, ok := d.GetOk("session_timeout"); ok {
				err = c.UpdateSessionTimeout(tenantUID,
					&models.V1AuthTokenSettings{ExpiryTimeMinutes: int32(sessionTime.(int))})
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
		if d.HasChange("login_banner") {
			loginBanner := d.Get("login_banner").([]interface{})
			// login banner
			if len(loginBanner) == 1 {
				bannerData := loginBanner[0].(map[string]interface{})
				bannerSetting := &models.V1LoginBannerSettings{
					Message:   bannerData["message"].(string),
					IsEnabled: true,
					Title:     bannerData["title"].(string),
				}
				err = c.UpdateLoginBanner(tenantUID, bannerSetting)
				if err != nil {
					return diag.FromErr(err)
				}
			} else {
				bannerSetting := &models.V1LoginBannerSettings{
					Message:   "",
					IsEnabled: false,
					Title:     "",
				}
				err = c.UpdateLoginBanner(tenantUID, bannerSetting)
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
		if d.HasChanges("cluster_auto_remediation", "enable_auto_remediation") {
			// cluster node remediation for tenant
			err = c.UpdateClusterAutoRemediationForTenant(tenantUID, remediationSettings)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		// non fip related setting
		fipsAddonPack := "nonFipsDisabled"
		fipsFeatures := "nonFipsDisabled"
		fipsClusterImport := "nonFipsDisabled"
		if d.HasChanges("non_fips_addon_pack", "non_fips_features", "non_fips_cluster_import") {
			if v, ok := d.GetOk("non_fips_addon_pack"); ok {
				fipsAddonPack = convertFIPSBool(v.(bool))
			}
			if v, ok := d.GetOk("non_fips_features"); ok {
				fipsFeatures = convertFIPSBool(v.(bool))
			}
			if v, ok := d.GetOk("non_fips_cluster_import"); ok {
				fipsClusterImport = convertFIPSBool(v.(bool))
			}
			err = c.UpdateFIPSPreference(tenantUID, &models.V1FipsSettings{
				FipsClusterFeatureConfig: &models.V1NonFipsConfig{Mode: &fipsFeatures},
				FipsClusterImportConfig:  &models.V1NonFipsConfig{Mode: &fipsClusterImport},
				FipsPackConfig:           &models.V1NonFipsConfig{Mode: &fipsAddonPack},
			})
			if err != nil {
				return diag.FromErr(err)
			}
		}

	} else {
		// cluster node remediation for project
		if d.HasChanges("cluster_auto_remediation", "enable_auto_remediation") {
			err = c.UpdateClusterAutoRemediationForProject(ProviderInitProjectUid, remediationSettings)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	// pause agent upgrade setting according to context
	if d.HasChange("pause_agent_upgrades") {
		err = c.UpdatePlatformClusterUpgradeSetting(&models.V1ClusterUpgradeSettingsEntity{
			SpectroComponents: d.Get("pause_agent_upgrades").(string)})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func updatePlatformSettingsDefault(d *schema.ResourceData, m interface{}) diag.Diagnostics {
	platformSettingContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, platformSettingContext)
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	var diags diag.Diagnostics
	remediationSettings := &models.V1NodesAutoRemediationSettings{
		DisableNodesAutoRemediation: true,
		IsEnabled:                   true,
	}
	if platformSettingContext == tenantString {
		// session timeout
		err = c.UpdateSessionTimeout(tenantUID,
			&models.V1AuthTokenSettings{ExpiryTimeMinutes: int32(240)})
		if err != nil {
			return diag.FromErr(err)
		}

		bannerSetting := &models.V1LoginBannerSettings{
			Message:   "",
			IsEnabled: false,
			Title:     "",
		}
		err = c.UpdateLoginBanner(tenantUID, bannerSetting)
		if err != nil {
			return diag.FromErr(err)
		}
		// cluster node remediation for tenant
		err = c.UpdateClusterAutoRemediationForTenant(tenantUID, remediationSettings)
		if err != nil {
			return diag.FromErr(err)
		}
		// fips setting to default
		_, fpOk := d.GetOk("non_fips_addon_pack")
		_, ffOk := d.GetOk("non_fips_features")
		_, fiOk := d.GetOk("non_fips_cluster_import")

		fipsAddonPack := "nonFipsDisabled"
		fipsFeatures := "nonFipsDisabled"
		fipsClusterImport := "nonFipsDisabled"
		if fiOk || ffOk || fpOk {
			err = c.UpdateFIPSPreference(tenantUID, &models.V1FipsSettings{
				FipsClusterFeatureConfig: &models.V1NonFipsConfig{Mode: &fipsFeatures},
				FipsClusterImportConfig:  &models.V1NonFipsConfig{Mode: &fipsClusterImport},
				FipsPackConfig:           &models.V1NonFipsConfig{Mode: &fipsAddonPack},
			})
			if err != nil {
				return diag.FromErr(err)
			}
		}

	} else {
		// cluster node remediation for project
		err = c.UpdateClusterAutoRemediationForProject(ProviderInitProjectUid, remediationSettings)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	// pause agent upgrade setting according to context
	err = c.UpdatePlatformClusterUpgradeSetting(&models.V1ClusterUpgradeSettingsEntity{
		SpectroComponents: "unlock"})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourcePlatformSettingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return updatePlatformSettingsDefault(d, m)
}

func resourcePlatformSettingImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	platformContext, uid, err := ParseResourceID(d)
	if err != nil {
		return nil, err
	}
	err = ValidateContext(platformContext)
	if err != nil {
		return nil, err
	}
	c := getV1ClientWithResourceContext(m, platformContext)
	var diags diag.Diagnostics

	if platformContext == tenantString {
		givenTenantId := uid
		actualTenantId, err := c.GetTenantUID()
		if err != nil {
			return nil, err
		}
		if givenTenantId != actualTenantId {
			return nil, fmt.Errorf("tenant id is not valid with curent user or invalid tenant uid provided: %v", diags)
		}
		if err = d.Set("context", tenantString); err != nil {
			return nil, err
		}
	} else {
		givenProjectId := uid
		actualProjectId := ProviderInitProjectUid
		if givenProjectId != actualProjectId {
			return nil, fmt.Errorf("project id is not valid with provider initialization: %v", diags)
		}
		if err = d.Set("context", tenantString); err != nil {
			return nil, err
		}
	}
	diags = resourcePlatformSettingRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read developer settings for import: %v", diags)
	}
	return []*schema.ResourceData{d}, nil
}
