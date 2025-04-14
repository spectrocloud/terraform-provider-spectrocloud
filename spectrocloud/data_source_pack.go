package spectrocloud

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

var AllowedPackLayers = []string{
	string(models.V1PackLayerKernel),
	string(models.V1PackLayerOs),
	string(models.V1PackLayerK8s),
	string(models.V1PackLayerCni),
	string(models.V1PackLayerCsi),
	string(models.V1PackLayerAddon),
}

var AllowedAddonType = []string{"load balancer", "ingress", "logging", "monitoring", "security", "authentication",
	"servicemesh", "system app", "app services", "registry", "csi", "cni", "integration", ""}

var AllowedEnvs = []string{
	"all", "aws", "eks", "gcp", "gke", "vsphere",
	"maas", "openstack", "edge-native", "aks", "azure",
}

var AllowedPackType = []string{string(models.V1PackTypeSpectro), string(models.V1PackTypeHelm), string(models.V1PackTypeManifest), string(models.V1PackTypeOci)}

func dataSourcePack() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePackRead,
		Description: "This data resource provides the ability to search for a pack in the Palette registries. It supports more advanced search criteria than the `pack_simple` data source.",

		Schema: map[string]*schema.Schema{
			"filters": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Filters to apply when searching for a pack. This is a string of the form 'key1=value1' with 'AND', 'OR` operators. Refer to the Palette API [pack search API endpoint documentation](https://docs.spectrocloud.com/api/v1/v-1-packs-search/) for filter examples. The filter attribute will be deprecated soon; use `advance_filter` instead.",
				ConflictsWith: []string{"id", "cloud", "name", "version", "registry_uid"},
			},
			"advance_filters": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				Description:   "A set of advanced filters to refine the selection of packs. These filters allow users to specify criteria such as pack type, add-on type, pack layer, and environment.",
				ConflictsWith: []string{"id", "cloud", "filters"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pack_type": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Specify the type of pack. Allowed values are `helm`, `spectro`, `oci`, and `manifest`. If not specified, all options will be set by default.",
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(AllowedPackType, false),
							},
						},
						"addon_type": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Defines the type of add-on pack. Allowed values are `load balancer`, `ingress`, `logging`, `monitoring`, `security`, `authentication`, `servicemesh`, `system app`, `app services`, `registry` and `integration`. If not specified, all options will be set by default. For `storage` and `network` addon_type set `csi` or `cni` respectively in pack_layer",
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(AllowedAddonType, false),
							},
						},
						"pack_layer": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Indicates the pack layer, such as `kernel`, `os`, `k8s`, `cni`, `csi`, or `addon`. If not specified, all options will be set by default.",
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(AllowedPackLayers, false),
							},
						},
						"environment": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Defines the environment where the pack will be deployed. Options include `all`, `aws`, `eks`, `azure`, `aks`, `gcp`, `gke`, `vsphere`, `maas`, `openstack` and `edge-native`. If not specified, all options will be set by default.",
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(AllowedEnvs, false),
							},
						},
						"is_fips": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Indicates whether the pack is FIPS-compliant. If `true`, only FIPS-compliant components will be used.",
						},
						"pack_source": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Specify the source of the pack. Allowed values are `spectrocloud` and `community`. If not specified, all options will be set by default.",
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{"spectrocloud", "community"}, false),
							},
						},
					},
				},
			},
			"id": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				Description:   "The UID of the pack returned.",
				ConflictsWith: []string{"filters", "cloud", "name", "version", "registry_uid"},
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the pack to search for.",
				Computed:    true,
				Optional:    true,
			},
			"cloud": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Description: "Filter results by cloud type. If not provided, all cloud types are returned.",
				Set:         schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"version": {
				Type:        schema.TypeString,
				Description: "Specify the version of the pack to search for. If not set, the latest available version from the specified registry will be used.",
				Computed:    true,
				Optional:    true,
			},
			"registry_uid": {
				Type:        schema.TypeString,
				Description: "The unique identifier (UID) of the registry where the pack is located. Specify `registry_uid` to search within a specific registry.",
				Computed:    true,
				Optional:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "The type of pack to search for. Supported values are `helm`, `manifest`, `container`, `operator-instance`.",
				Computed:    true,
				Optional:    true,
			},
			"values": {
				Type:        schema.TypeString,
				Description: "The YAML values of the pack returned as string.",
				Computed:    true,
			},
		},
	}
}

func dataSourcePackRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var packName = ""
	var err error
	var advancePacksResult []*models.V1PackMetadata
	var packs []*models.V1PackSummary
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var advanceFilterSpec *models.V1PackFilterSpec
	if v, ok := d.GetOk("type"); ok {
		if v.(string) == "manifest" {
			return diags
		}
		if v.(string) == "helm" {
			if regUID, ok := d.GetOk("registry_uid"); ok {
				registry, err := c.GetHelmRegistry(regUID.(string))
				if err != nil {
					return diag.FromErr(err)
				}
				if registry.Spec.IsPrivate {
					return diags
				}
			}
		}
		if v.(string) == "oci" {
			if _, ok := d.GetOk("registry_uid"); ok {
				// we don't have provision to get all helm chart/packs from oci basic type registry, hence skipping validation
				// we will move registry validation in profile creation (TBU)
				return diags
			}
		}
	}

	filters := make([]string, 0)
	registryUID := ""
	if v, ok := d.GetOk("filters"); ok {
		filters = append(filters, v.(string))
	} else if v, ok := d.GetOk("id"); ok {
		filters = append(filters, fmt.Sprintf("metadata.uid=%s", v.(string)))
	} else if v, ok := d.GetOk("advance_filters"); ok {

		if v, ok := d.GetOk("name"); ok {
			packName = v.(string)
		}
		if v, ok := d.GetOk("registry_uid"); ok {
			registryUID = v.(string)
		}
		advanceFilter := v.([]interface{})[0].(map[string]interface{})
		var registryList []string
		if registryUID != "" {
			registryList = []string{registryUID}
		}
		packTypeValues := convertToV1PackType(advanceFilter["pack_type"].(*schema.Set)) // returns []models.V1PackType
		var packTypePtr []*models.V1PackType
		for i := range packTypeValues {
			packTypePtr = append(packTypePtr, &packTypeValues[i])
		}
		advanceFilterSpec = &models.V1PackFilterSpec{
			Name: &models.V1FilterString{
				Eq: StringPtr(packName),
			},
			Type:        packTypePtr,
			Layer:       convertToV1PackLayer(advanceFilter["pack_layer"].(*schema.Set)),
			Environment: convertToStringSlice(advanceFilter["environment"].(*schema.Set).List()),
			AddOnType:   convertToAddOnType(advanceFilter["addon_type"].(*schema.Set).List(), advanceFilter["pack_layer"].(*schema.Set)),
			RegistryUID: registryList,
			IsFips:      advanceFilter["is_fips"].(bool),
			Source:      convertToStringSlice(advanceFilter["pack_source"].(*schema.Set).List()),
		}
	} else {
		if v, ok := d.GetOk("name"); ok {

			/*
				Cluster profile now supports packs duplication, but pack name has to be unique and will be double dashed
				and first part would be any random name to make overall pack name unique and 2nd part is actual pack name.
				Thus, splitting pack name with '--' to get the correct pack name to find pack uuid
			*/

			if strings.Contains(v.(string), "--") {
				v = strings.Split(v.(string), "--")[1]
			}
			packName = v.(string)
			filters = append(filters, fmt.Sprintf("spec.name=%s", v.(string)))
		}
		if v, ok := d.GetOk("registry_uid"); ok {
			registryUID = v.(string)
		}
		if v, ok := d.GetOk("version"); ok {
			filters = append(filters, fmt.Sprintf("spec.version=%s", v.(string)))
		} else {
			latestVersion := setLatestPackVersionToFilters(packName, registryUID, c)
			if latestVersion != "" {
				filters = append(filters, fmt.Sprintf("spec.version=%s", latestVersion))
			}
		}
		if v, ok := d.GetOk("cloud"); ok {
			clouds := expandStringList(v.(*schema.Set).List())
			if !stringContains(clouds, "all") {
				clouds = append(clouds, "all")
			}
			filters = append(filters, fmt.Sprintf("spec.cloudTypes_in_%s", strings.Join(clouds, ",")))
		}
	}

	if _, ok := d.GetOk("advance_filters"); ok {
		advancePacksResult, err = c.SearchPacks(advanceFilterSpec, nil)
		if err != nil {
			return diag.FromErr(err)
		}

		resultCount := len(advancePacksResult)
		if resultCount == 0 {
			return diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "no matching packs for advance_filters",
				Detail:   "No packs matching criteria found",
			}}
		}

		if resultCount > 1 {
			return diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "multiple packs returned for specified advance_filter",
				Detail:   fmt.Sprintf("Found %d matching packs. Restrict packs criteria to a single match.", resultCount),
			}}
		}

		registries := advancePacksResult[0].Spec.Registries
		registryCount := len(registries)

		if registryCount == 0 {
			return diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "no matching packs for advance_filters",
				Detail:   "No packs matching criteria found",
			}}
		}

		if registryCount > 1 {
			return diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "packs available in multiple registries for given advance_filter",
				Detail:   fmt.Sprintf("Packs found in %d registries. Restrict packs criteria to a single match.", registryCount),
			}}
		}

		// Exactly one registry
		//if ver, ok := d.GetOk("version"); ok {
		supportedVersionList, err := c.GetPacksByNameAndRegistry(packName, registryUID)
		if err != nil {
			return diag.FromErr(err)
		}
		if ver, ok := d.GetOk("version"); ok {
			for _, v := range supportedVersionList.Tags {
				if ver == v.Version {
					filters = []string{fmt.Sprintf("metadata.uid=%s", v.PackUID)}
					break
				}
			}
			if len(filters) == 0 {
				return diag.Diagnostics{{
					Severity: diag.Error,
					Summary:  "no matching packs for advance_filters",
					Detail:   "No packs matching criteria found",
				}}
			}
		} else {
			if supportedVersionList != nil {
				filters = []string{fmt.Sprintf("metadata.uid=%s", supportedVersionList.Tags[len(supportedVersionList.Tags)-1].PackUID)}
			}

		}
	}

	packs, err = c.GetPacks(filters, registryUID)
	if err != nil {
		return diag.FromErr(err)
	}

	packName = "unknown"
	if v, ok := d.GetOk("name"); ok {
		packName = v.(string)
	}

	if len(packs) == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("%s: no matching packs", packName),
			Detail:   "No packs matching criteria found",
		})
		return diags
	} else if len(packs) > 1 {
		packs_map := make(map[string]string, 0)
		for _, pack := range packs {
			packs_map[pack.Spec.RegistryUID] = pack.Spec.Name
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("%s: Multiple packs returned", packName),
			Detail:   fmt.Sprintf("Found %d matching packs. Restrict packs criteria to a single match. %s", len(packs), packs_map),
		})
		return diags
	}

	pack := packs[0]

	clouds := make([]string, 0)
	for _, cloudType := range pack.Spec.CloudTypes {
		clouds = append(clouds, string(cloudType))
	}

	d.SetId(pack.Metadata.UID)
	err = d.Set("name", pack.Spec.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("cloud", clouds)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("version", pack.Spec.Version)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("registry_uid", pack.Spec.RegistryUID)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("values", pack.Spec.Values)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func setLatestPackVersionToFilters(packName string, registryUID string, c *client.V1Client) string {
	var packLayers = []models.V1PackLayer{models.V1PackLayerKernel, models.V1PackLayerOs, models.V1PackLayerK8s, models.V1PackLayerCni, models.V1PackLayerCsi, models.V1PackLayerAddon}

	var packTypes = []*models.V1PackType{
		types.Ptr(models.V1PackTypeSpectro),
		types.Ptr(models.V1PackTypeHelm),
		types.Ptr(models.V1PackTypeManifest),
		types.Ptr(models.V1PackTypeOci),
	}

	newFilter := &models.V1PackFilterSpec{
		Name: &models.V1FilterString{
			Eq: StringPtr(packName),
		},
		Type:        packTypes,
		Layer:       packLayers,
		Environment: []string{"all"},
		AddOnType:   AllowedAddonType,
	}
	if registryUID != "" {
		newFilter.RegistryUID = []string{registryUID}
	}
	var newSort []*models.V1PackSortSpec
	latestVersion := ""
	packsResults, _ := c.SearchPacks(newFilter, newSort)
	if len(packsResults) == 1 {
		latestVersion, _ = getLatestVersion(packsResults[0].Spec.Registries)
		return latestVersion
	}
	return ""
}

// getLatestVersion returns the latest version from a list of version strings.
func getLatestVersion(versions []*models.V1RegistryPackMetadata) (string, error) {
	if len(versions) == 0 {
		return "", fmt.Errorf("no versions provided")
	}
	semverVersions := make([]*semver.Version, len(versions))
	for i, v := range versions {
		ver, err := semver.NewVersion(v.LatestVersion)
		if err != nil {
			return "", fmt.Errorf("invalid version %q: %w", v, err)
		}
		semverVersions[i] = ver
	}
	sort.Sort(semver.Collection(semverVersions))

	return semverVersions[len(semverVersions)-1].Original(), nil
}

func convertToV1PackType(set *schema.Set) []models.V1PackType {
	var result []models.V1PackType
	for _, v := range set.List() {
		if str, ok := v.(string); ok {
			result = append(result, models.V1PackType(str))
		}
	}
	return result
}

func convertToV1PackLayer(set *schema.Set) []models.V1PackLayer {
	var result []models.V1PackLayer
	for _, v := range set.List() {
		if str, ok := v.(string); ok {
			result = append(result, models.V1PackLayer(str))
		}
	}
	return result
}

func convertToAddOnType(input []interface{}, packLayer *schema.Set) []string {
	result := make([]string, len(input))
	for i, v := range input {
		if str, ok := v.(string); ok {
			result[i] = str
		}
	}
	if len(result) == 0 {
		pl := convertToV1PackLayer(packLayer)
		for _, v := range pl {
			if v == "addon" {
				result = AllowedAddonType
			}
		}
	}
	return result
}

func convertToStringSlice(input []interface{}) []string {
	result := make([]string, len(input))
	for i, v := range input {
		if str, ok := v.(string); ok {
			result[i] = str
		}
	}
	return result
}
