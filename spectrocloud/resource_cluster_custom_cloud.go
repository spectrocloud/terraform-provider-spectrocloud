package spectrocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	"gopkg.in/yaml.v3"
)

func resourceClusterCustomCloud() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterCustomCloudCreate,
		ReadContext:   resourceClusterCustomCloudRead,
		UpdateContext: resourceClusterCustomCloudUpdate,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{

			StateContext: resourceClusterCustomImport,
		},
		Description: "Resource for managing custom cloud clusters in Spectro Cloud through Palette.",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the cluster.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the EKS cluster. Allowed values are `project` or `tenant`. " +
					"Default is `project`. " + PROJECT_NAME_NUANCE,
			},
			"cloud": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The cloud provider name.",
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A list of tags to be applied to the cluster. Tags must be in the form of `key:value`.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The description of the cluster. Default value is empty string.",
			},
			"cluster_profile": schemas.ClusterProfileSchema(),
			"apply_setting": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "DownloadAndInstall",
				ValidateFunc: validation.StringInSlice([]string{"DownloadAndInstall", "DownloadAndInstallLater"}, false),
				Description: "The setting to apply the cluster profile. `DownloadAndInstall` will download and install packs in one action. " +
					"`DownloadAndInstallLater` will only download artifact and postpone install for later. " +
					"Default value is `DownloadAndInstall`.",
			},
			"cloud_account_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The cloud account id to use for this cluster.",
			},
			"cloud_config_id": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "This field is deprecated and will be removed in the future. Use `cloud_config` instead.",
			},
			"cloud_config": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "The Cloud environment configuration settings such as network parameters and encryption parameters that apply to this cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"values": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The values of the cloud config. The values are specified in YAML format. ",
							StateFunc: func(val interface{}) string {
								// Normalize YAML content to handle formatting differences
								if yamlStr, ok := val.(string); ok {
									return NormalizeYamlContent(yamlStr)
								}
								return val.(string)
							},
						},
						"overrides": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Description: "Key-value pairs to override specific values in the YAML. Supports template variables, wildcard patterns, field pattern search, document-specific and global overrides.\n\n" +
								"Template variables: Simple identifiers that replace ${var}, {{var}}, or $var patterns in YAML (e.g., 'cluster_name' replaces ${cluster_name})\n" +
								"Wildcard patterns: Patterns starting with '*' that match field names containing the specified substring (e.g., '*cluster-api-autoscaler-node-group-max-size' matches any field containing 'cluster-api-autoscaler-node-group-max-size')\n" +
								"Field pattern search: Patterns that find and update ALL matching nested fields anywhere in YAML (e.g., 'replicas' updates any 'replicas' field, 'rootVolume.size' updates any 'rootVolume.size' pattern)\n" +
								"Document-specific syntax: 'Kind.path' (e.g., 'Cluster.metadata.labels', 'AWSCluster.spec.region')\n" +
								"Global path syntax: 'path' (e.g., 'metadata.name', 'spec.region')\n\n" +
								"Processing order: 1) Template substitution, 2) Wildcard patterns, 3) Field pattern search, 4) Path-based overrides. " +
								"Supports dot notation for nested paths and array indexing with [index]. " +
								"Values are strings but support JSON syntax for arrays/objects.",
						},
					},
				},
			},

			"machine_pool": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "The machine pool configuration for the cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the machine pool. This will be derived from the name value in the `node_pool_config`.",
						},
						"count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of nodes in the machine pool. This will be derived from the replica value in the 'node_pool_config'.",
						},
						"control_plane": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether this machine pool is a control plane. Defaults to `false`.",
						},
						"control_plane_as_worker": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether this machine pool is a control plane and a worker. Defaults to `false`.",
						},
						"node_pool_config": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The values of the node pool config. The values are specified in YAML format. ",
							StateFunc: func(val interface{}) string {
								// Normalize YAML content to handle formatting differences
								if yamlStr, ok := val.(string); ok {
									return NormalizeYamlContent(yamlStr)
								}
								return val.(string)
							},
						},
						"overrides": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Description: "Key-value pairs to override specific values in the node pool config YAML. Supports template variables, wildcard patterns, field pattern search, document-specific and global overrides.\n\n" +
								"Template variables: Simple identifiers that replace ${var}, {{var}}, or $var patterns in YAML (e.g., 'node_count' replaces ${node_count})\n" +
								"Wildcard patterns: Patterns starting with '*' that match field names containing the specified substring (e.g., '*cluster-api-autoscaler-node-group-max-size' matches any field containing 'cluster-api-autoscaler-node-group-max-size')\n" +
								"Field pattern search: Patterns that find and update ALL matching nested fields anywhere in YAML (e.g., 'replicas' updates any 'replicas' field, 'rootVolume.size' updates any 'rootVolume.size' pattern)\n" +
								"Document-specific syntax: 'Kind.path' (e.g., 'AWSMachineTemplate.spec.template.spec.instanceType')\n" +
								"Global path syntax: 'path' (e.g., 'metadata.name', 'spec.instanceType')\n\n" +
								"Processing order: 1) Template substitution, 2) Wildcard patterns, 3) Field pattern search, 4) Path-based overrides. " +
								"Supports dot notation for nested paths and array indexing with [index]. " +
								"Values are strings but support JSON syntax for arrays/objects.",
						},
						// Planned for support on future release's - "update_strategy", "node_repave_interval"
					},
				},
			},
			"pause_agent_upgrades": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "unlock",
				ValidateFunc: validation.StringInSlice([]string{"lock", "unlock"}, false),
				Description:  "The pause agent upgrades setting allows to control the automatic upgrade of the Palette component and agent for an individual cluster. The default value is `unlock`, meaning upgrades occur automatically. Setting it to `lock` pauses automatic agent upgrades for the cluster.",
			},
			"os_patch_on_boot": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to apply OS patch on boot. Default is `false`.",
			},
			"os_patch_schedule": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateOsPatchSchedule,
				Description:      "The cron schedule for OS patching. This must be in the form of cron syntax. Ex: `0 0 * * *`.",
			},
			"os_patch_after": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateOsPatchOnDemandAfter,
				Description:      "Date and time after which to patch cluster `RFC3339: 2006-01-02T15:04:05Z07:00`",
			},

			"kubeconfig": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Kubeconfig for the cluster. This can be used to connect to the cluster using `kubectl`.",
			},
			"admin_kube_config": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Admin Kube-config for the cluster. This can be used to connect to the cluster using `kubectl`, With admin privilege.",
			},
			"backup_policy":        schemas.BackupPolicySchema(),
			"scan_policy":          schemas.ScanPolicySchema(),
			"cluster_rbac_binding": schemas.ClusterRbacBindingSchema(),
			"namespaces":           schemas.ClusterNamespacesSchema(),
			"location_config":      schemas.ClusterLocationSchema(),

			"skip_completion": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If `true`, the cluster will be created asynchronously. Default value is `false`.",
			},
			"force_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If set to `true`, the cluster will be force deleted and user has to manually clean up the provisioned cloud resources.",
			},
			"force_delete_delay": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          20,
				Description:      "Delay duration in minutes to before invoking cluster force delete. Default and minimum is 20.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(20)),
			},
			// Planned for support on future release's - "review_repave_state",
		},
	}
}

func resourceClusterCustomCloudCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	cluster, err := toCustomCloudCluster(c, d)
	if err != nil {
		return diag.FromErr(err)
	}
	cloudType := d.Get("cloud").(string)

	err = c.ValidateCustomCloudType(cloudType)
	if err != nil {
		return diag.FromErr(err)
	}

	uid, err := c.CreateClusterCustomCloud(cluster, cloudType)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, isError := waitForClusterCreation(ctx, d, uid, diags, c, true)
	if isError && diagnostics != nil {
		return diagnostics
	}

	resourceClusterCustomCloudRead(ctx, d, m)

	return diags
}

func resourceClusterCustomCloudRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	cluster, err := resourceClusterRead(d, c, diags)
	if err != nil {
		return handleReadError(d, err, diags)
	} else if cluster == nil {
		d.SetId("")
		return diags
	}
	diagnostics, hasError := readCommonFields(c, d, cluster)
	if hasError {
		return diagnostics
	}
	diagnostics, hasError = flattenCloudConfigCustom(cluster.Spec.CloudConfigRef.UID, d, c)
	if hasError {
		return diagnostics
	}

	return diags
}

func resourceClusterCustomCloudUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cloudConfigId := d.Get("cloud_config_id").(string)
	//clusterContext := d.Get("context").(string)
	cloudType := d.Get("cloud").(string)

	_, err := c.GetCloudConfigCustomCloud(cloudConfigId, cloudType)
	if err != nil {
		return diag.FromErr(err)
	}
	if d.HasChange("cloud_config") {
		config := toCustomCloudConfig(d)
		configEntity := &models.V1CustomCloudClusterConfigEntity{
			ClusterConfig: config,
		}
		err = c.UpdateCloudConfigCustomCloud(configEntity, cloudConfigId, cloudType)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("machine_pool") {
		oraw, nraw := d.GetChange("machine_pool")
		if oraw == nil {
			oraw = new(schema.Set)
		}
		if nraw == nil {
			nraw = new(schema.Set)
		}

		os := oraw.([]interface{})
		ns := nraw.([]interface{})

		osMap := make(map[string]interface{})
		for _, mp := range os {
			machinePool := mp.(map[string]interface{})
			osMap[machinePool["name"].(string)] = machinePool
		}

		nsMap := make(map[string]interface{})
		for _, mp := range ns {
			machinePoolResource := mp.(map[string]interface{})
			nsMap[machinePoolResource["name"].(string)] = machinePoolResource
			if machinePoolResource["name"].(string) != "" {
				name := machinePoolResource["name"].(string)
				newHash := resourceMachinePoolCustomCloudHash(machinePoolResource)
				var err error
				machinePool := toMachinePoolCustomCloud(mp)
				if oldMachinePool, ok := osMap[name]; !ok {
					log.Printf("[DEBUG] Creating new machine pool %s", name)
					if err = c.CreateMachinePoolCustomCloud(machinePool, cloudConfigId, cloudType); err != nil {
						return diag.FromErr(err)
					}
				} else {
					oldHash := resourceMachinePoolCustomCloudHash(oldMachinePool)
					log.Printf("[DEBUG] Machine pool %s - Old hash: %d, New hash: %d", name, oldHash, newHash)
					if newHash != oldHash {
						log.Printf("[DEBUG] Change detected in machine pool %s - updating", name)
						if err = c.UpdateMachinePoolCustomCloud(machinePool, name, cloudConfigId, cloudType); err != nil {
							return diag.FromErr(err)
						}
					} else {
						log.Printf("[DEBUG] No changes detected in machine pool %s - skipping update", name)
					}
				}
				// Processed (if exists)
				delete(osMap, name)
			}
		}
		// Deleted old machine pools
		for _, mp := range osMap {
			machinePool := mp.(map[string]interface{})
			name := machinePool["name"].(string)
			log.Printf("Deleted machine pool %s", name)
			if err = c.DeleteMachinePoolCustomCloud(name, cloudConfigId, cloudType); err != nil {
				return diag.FromErr(err)
			}
		}

	}

	diagnostics, done := updateCommonFields(d, c)
	if done {
		return diagnostics
	}

	resourceClusterCustomCloudRead(ctx, d, m)

	return diags
}

func toCustomCloudCluster(c *client.V1Client, d *schema.ResourceData) (*models.V1SpectroCustomClusterEntity, error) {

	clusterContext := d.Get("context").(string)
	profiles, err := toProfiles(c, d, clusterContext)
	if err != nil {
		return nil, err
	}

	// policies in not supported for custom cluster during cluster creation UI also its same.
	// policies := toPolicies(d)

	customCloudConfig := toCustomCloudConfig(d)

	customClusterConfig := toCustomClusterConfig(d)

	machinePoolConfigs := make([]*models.V1CustomMachinePoolConfigEntity, 0)
	for _, machinePool := range d.Get("machine_pool").([]interface{}) {
		mp := toMachinePoolCustomCloud(machinePool)
		machinePoolConfigs = append(machinePoolConfigs, mp)
	}

	cluster := &models.V1SpectroCustomClusterEntity{
		Metadata: toClusterMetadataUpdate(d),
		Spec: &models.V1SpectroCustomClusterEntitySpec{
			CloudAccountUID:   types.Ptr(d.Get("cloud_account_id").(string)),
			CloudConfig:       customCloudConfig,
			ClusterConfig:     customClusterConfig,
			Machinepoolconfig: machinePoolConfigs,
			Profiles:          profiles,
		},
	}

	return cluster, nil
}

func toCustomCloudConfig(d *schema.ResourceData) *models.V1CustomClusterConfig {
	cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})
	valuesYamlStr := strings.TrimSpace(cloudConfig["values"].(string))

	// Apply overrides if they exist
	if overrides, ok := cloudConfig["overrides"].(map[string]interface{}); ok && len(overrides) > 0 {
		log.Printf("[DEBUG] Applying %d YAML overrides to cloud config", len(overrides))
		for path, value := range overrides {
			log.Printf("[DEBUG] Override: %s = %v", path, value)
		}

		processedYaml, err := applyYamlOverridesWithTemplates(valuesYamlStr, overrides)
		if err != nil {
			log.Printf("Warning: Failed to apply YAML overrides: %v", err)
		} else {
			log.Printf("[DEBUG] YAML transformation successful. Original length: %d, New length: %d", len(valuesYamlStr), len(processedYaml))
			// Show a snippet of the transformation for debugging
			if len(processedYaml) > 0 && processedYaml != valuesYamlStr {
				log.Printf("[DEBUG] YAML values updated with overrides")
			}
			valuesYamlStr = processedYaml
		}
	}

	// Normalize the final YAML content to ensure consistent formatting
	valuesYamlStr = NormalizeYamlContent(valuesYamlStr)

	customCloudConfig := &models.V1CustomClusterConfig{
		Values: StringPtr(valuesYamlStr),
	}

	return customCloudConfig
}

// applyYamlOverrides applies key-value overrides to multi-document YAML
func applyYamlOverrides(yamlContent string, overrides map[string]interface{}) (string, error) {
	// Split multi-document YAML
	documents := strings.Split(yamlContent, "---")
	var processedDocs []string

	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			processedDocs = append(processedDocs, "")
			continue
		}

		// Parse YAML document
		var yamlData interface{}
		if err := yaml.Unmarshal([]byte(doc), &yamlData); err != nil {
			// If parsing fails, keep original document
			processedDocs = append(processedDocs, doc)
			continue
		}

		// Extract document kind for document-specific overrides
		documentKind := extractDocumentKind(yamlData)

		// Apply overrides to this document
		modified := false
		for path, value := range overrides {
			if applyOverrideToDocumentWithKind(&yamlData, path, value, documentKind) {
				log.Printf("[DEBUG] Successfully applied override: %s = %v to %s document", path, value, documentKind)
				modified = true
			}
		}

		// Convert back to YAML
		if modified {
			processedYaml, err := yaml.Marshal(yamlData)
			if err != nil {
				processedDocs = append(processedDocs, doc)
				continue
			}
			processedDocs = append(processedDocs, strings.TrimSpace(string(processedYaml)))
		} else {
			processedDocs = append(processedDocs, doc)
		}
	}

	return strings.Join(processedDocs, "\n---\n"), nil
}

// applyYamlOverridesWithTemplates applies template substitution, wildcard patterns, field name overrides, and path-based overrides
func applyYamlOverridesWithTemplates(yamlContent string, overrides map[string]interface{}) (string, error) {
	// Step 1: Separate override types
	templateVars, wildcardPatterns, fieldPatternOverrides, pathOverrides := separateOverrideTypes(yamlContent, overrides)

	processedYaml := yamlContent

	// Step 2: Apply template variable substitution first
	if len(templateVars) > 0 {
		log.Printf("[DEBUG] Applying %d template variable substitutions", len(templateVars))
		processedYaml = applyTemplateSubstitution(yamlContent, templateVars)
	}

	// Step 3: Apply wildcard pattern overrides
	if len(wildcardPatterns) > 0 {
		log.Printf("[DEBUG] Applying %d wildcard pattern overrides", len(wildcardPatterns))
		var err error
		processedYaml, err = applyWildcardPatternOverrides(processedYaml, wildcardPatterns)
		if err != nil {
			return processedYaml, err
		}
	}

	// Step 4: Apply field pattern overrides
	if len(fieldPatternOverrides) > 0 {
		log.Printf("[DEBUG] Applying %d field pattern overrides", len(fieldPatternOverrides))
		var err error
		processedYaml, err = applyFieldPatternOverrides(processedYaml, fieldPatternOverrides)
		if err != nil {
			return processedYaml, err
		}
	}

	// Step 5: Apply path-based overrides on the processed YAML
	if len(pathOverrides) > 0 {
		log.Printf("[DEBUG] Applying %d path-based overrides", len(pathOverrides))
		return applyYamlOverrides(processedYaml, pathOverrides)
	}

	return processedYaml, nil
}

// applyFieldPatternOverrides applies field pattern-based overrides to multi-document YAML
func applyFieldPatternOverrides(yamlContent string, fieldOverrides map[string]interface{}) (string, error) {
	// Split multi-document YAML
	documents := strings.Split(yamlContent, "---")
	var processedDocs []string

	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			processedDocs = append(processedDocs, "")
			continue
		}

		// Parse YAML document
		var yamlData interface{}
		if err := yaml.Unmarshal([]byte(doc), &yamlData); err != nil {
			// If parsing fails, keep original document
			processedDocs = append(processedDocs, doc)
			continue
		}

		// Apply field pattern overrides to this document
		modified := false
		for fieldPattern, value := range fieldOverrides {
			convertedValue := convertStringToAppropriateType(value.(string))
			if applyFieldPatternOverrideToDocument(&yamlData, fieldPattern, convertedValue) {
				log.Printf("[DEBUG] Successfully applied field pattern override: %s = %v", fieldPattern, convertedValue)
				modified = true
			}
		}

		// Convert back to YAML
		if modified {
			processedYaml, err := yaml.Marshal(yamlData)
			if err != nil {
				processedDocs = append(processedDocs, doc)
				continue
			}
			processedDocs = append(processedDocs, strings.TrimSpace(string(processedYaml)))
		} else {
			processedDocs = append(processedDocs, doc)
		}
	}

	return strings.Join(processedDocs, "\n---\n"), nil
}

// applyFieldPatternOverrideToDocument applies a field pattern override to all matching patterns in a document
func applyFieldPatternOverrideToDocument(data *interface{}, fieldPattern string, value interface{}) bool {
	modified := false

	// Split the field pattern into parts (e.g., "rootVolume.size" -> ["rootVolume", "size"])
	patternParts := strings.Split(fieldPattern, ".")

	// Try to find and update the pattern starting from current level
	if findAndUpdatePattern(data, patternParts, value, "") {
		modified = true
	}

	return modified
}

// findAndUpdatePattern recursively searches for and updates field patterns in YAML data
func findAndUpdatePattern(data *interface{}, patternParts []string, value interface{}, currentPath string) bool {
	if len(patternParts) == 0 {
		return false
	}

	modified := false

	switch currentData := (*data).(type) {
	case map[string]interface{}:
		// Check if we can match the pattern starting from this level
		if canMatchPatternFromHere(currentData, patternParts) {
			if applyPatternToMap(currentData, patternParts, value) {
				log.Printf("[DEBUG] Found and updated field pattern '%s' at path '%s'", strings.Join(patternParts, "."), currentPath)
				modified = true
			}
		}

		// Recursively search nested structures
		for key, v := range currentData {
			newPath := currentPath
			if newPath != "" {
				newPath += "."
			}
			newPath += key

			if findAndUpdatePattern(&v, patternParts, value, newPath) {
				modified = true
			}
		}

	case map[interface{}]interface{}:
		// Convert to string keys and search
		stringMap := make(map[string]interface{})
		for k, v := range currentData {
			if keyStr, ok := k.(string); ok {
				stringMap[keyStr] = v
			}
		}

		// Check pattern match and recursively search
		if canMatchPatternFromHere(stringMap, patternParts) {
			if applyPatternToMap(stringMap, patternParts, value) {
				log.Printf("[DEBUG] Found and updated field pattern '%s' at path '%s'", strings.Join(patternParts, "."), currentPath)
				*data = stringMap
				modified = true
			}
		}

		for key, v := range stringMap {
			newPath := currentPath
			if newPath != "" {
				newPath += "."
			}
			newPath += key

			if findAndUpdatePattern(&v, patternParts, value, newPath) {
				modified = true
			}
		}

	case []interface{}:
		// Search in array elements
		for i := range currentData {
			arrayPath := fmt.Sprintf("%s[%d]", currentPath, i)
			if findAndUpdatePattern(&currentData[i], patternParts, value, arrayPath) {
				modified = true
			}
		}
	}

	return modified
}

// canMatchPatternFromHere checks if a field pattern can be matched starting from the given map
func canMatchPatternFromHere(data map[string]interface{}, patternParts []string) bool {
	current := data

	for i, part := range patternParts {
		if i == len(patternParts)-1 {
			// Last part - just check if key exists
			_, exists := current[part]
			return exists
		}

		// Intermediate part - must exist and be a map
		if val, exists := current[part]; exists {
			if nextMap, ok := val.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return false
			}
		} else {
			return false
		}
	}

	return false
}

// applyPatternToMap applies a pattern to a map by navigating the nested structure
func applyPatternToMap(data map[string]interface{}, patternParts []string, value interface{}) bool {
	current := data

	for i, part := range patternParts {
		if i == len(patternParts)-1 {
			// Last part - update the value
			current[part] = value
			return true
		}

		// Intermediate part - navigate deeper
		if val, exists := current[part]; exists {
			if nextMap, ok := val.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return false
			}
		} else {
			return false
		}
	}

	return false
}

// separateTemplateAndPathOverrides separates template variables from path-based overrides and field name searches
func separateTemplateAndPathOverrides(overrides map[string]interface{}) (map[string]interface{}, map[string]interface{}) {
	templateVars := make(map[string]interface{})
	pathOverrides := make(map[string]interface{})

	for key, value := range overrides {
		if isTemplateVariable(key) {
			templateVars[key] = value
		} else {
			pathOverrides[key] = value
		}
	}

	return templateVars, pathOverrides
}

// separateOverrideTypes separates overrides into template variables, wildcard patterns, field pattern searches, and path-based overrides
func separateOverrideTypes(yamlContent string, overrides map[string]interface{}) (map[string]interface{}, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	templateVars := make(map[string]interface{})
	wildcardPatterns := make(map[string]interface{})
	fieldPatternOverrides := make(map[string]interface{})
	pathOverrides := make(map[string]interface{})

	for key, value := range overrides {
		if isActualTemplateVariable(yamlContent, key) {
			templateVars[key] = value
		} else if isWildcardPattern(key) {
			wildcardPatterns[key] = value
		} else if isFieldPattern(key) {
			fieldPatternOverrides[key] = value
		} else {
			pathOverrides[key] = value
		}
	}

	return templateVars, wildcardPatterns, fieldPatternOverrides, pathOverrides
}

// isActualTemplateVariable checks if a key is actually used as a template variable in the YAML content
func isActualTemplateVariable(yamlContent string, key string) bool {
	// Template variables must be simple identifiers without dots
	if strings.Contains(key, ".") || strings.Contains(key, "[") || strings.Contains(key, "]") {
		return false
	}

	// Must be a valid identifier
	for _, char := range key {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return false
		}
	}

	if len(key) == 0 {
		return false
	}

	// Check if the key is actually used as a template variable in the YAML
	patterns := []string{
		"${" + key + "}",  // ${key}
		"{{" + key + "}}", // {{key}}
		"$" + key,         // $key (simple)
	}

	for _, pattern := range patterns {
		if strings.Contains(yamlContent, pattern) {
			log.Printf("[DEBUG] Found template variable pattern '%s' for key '%s'", pattern, key)
			return true
		}
	}

	return false
}

// isFieldPattern determines if a key represents a field pattern for field pattern search
func isFieldPattern(key string) bool {
	// Must not contain array notation
	if strings.Contains(key, "[") || strings.Contains(key, "]") {
		return false
	}

	// Allow field patterns with any case - the distinction between document-specific
	// paths and field patterns will be made based on dot notation presence

	// Split by dots to validate each part
	parts := strings.Split(key, ".")
	for _, part := range parts {
		if part == "" {
			return false // Empty parts not allowed
		}

		// Each part must be a valid identifier (letters, numbers, underscore, hyphen)
		for _, char := range part {
			if !((char >= 'a' && char <= 'z') ||
				(char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9') ||
				char == '_' || char == '-') {
				return false
			}
		}
	}

	return len(key) > 0
}

// isTemplateVariable determines if a key represents a template variable
// Template variables are simple identifiers without dots, colons, or uppercase prefixes
func isTemplateVariable(key string) bool {
	// Must not contain dots (path separators)
	if strings.Contains(key, ".") {
		return false
	}

	// Must not contain array notation
	if strings.Contains(key, "[") || strings.Contains(key, "]") {
		return false
	}

	// Must not start with uppercase (Kind prefixes)
	if len(key) > 0 && key[0] >= 'A' && key[0] <= 'Z' {
		return false
	}

	// Must be a valid identifier (letters, numbers, underscore, hyphen)
	for _, char := range key {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return false
		}
	}

	return len(key) > 0
}

// isWildcardPattern determines if a key represents a wildcard pattern (starts with *)
func isWildcardPattern(key string) bool {
	return strings.HasPrefix(key, "*") && len(key) > 1
}

// applyWildcardPatternOverrides applies wildcard pattern-based overrides to multi-document YAML
func applyWildcardPatternOverrides(yamlContent string, wildcardOverrides map[string]interface{}) (string, error) {
	// Split multi-document YAML
	documents := strings.Split(yamlContent, "---")
	var processedDocs []string

	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			processedDocs = append(processedDocs, "")
			continue
		}

		// Parse YAML document
		var yamlData interface{}
		if err := yaml.Unmarshal([]byte(doc), &yamlData); err != nil {
			// If parsing fails, keep original document
			processedDocs = append(processedDocs, doc)
			continue
		}

		// Apply wildcard pattern overrides to this document
		modified := false
		for wildcardPattern, value := range wildcardOverrides {
			// Remove the * prefix to get the actual pattern to match
			pattern := strings.TrimPrefix(wildcardPattern, "*")
			convertedValue := convertStringToAppropriateType(value.(string))

			if applyWildcardPatternToDocument(&yamlData, pattern, convertedValue) {
				log.Printf("[DEBUG] Successfully applied wildcard pattern override: %s = %v", wildcardPattern, convertedValue)
				modified = true
			}
		}

		// Convert back to YAML
		if modified {
			processedYaml, err := yaml.Marshal(yamlData)
			if err != nil {
				processedDocs = append(processedDocs, doc)
				continue
			}
			processedDocs = append(processedDocs, strings.TrimSpace(string(processedYaml)))
		} else {
			processedDocs = append(processedDocs, doc)
		}
	}

	return strings.Join(processedDocs, "\n---\n"), nil
}

// applyWildcardPatternToDocument applies a wildcard pattern override to all matching field names in a document
func applyWildcardPatternToDocument(data *interface{}, pattern string, value interface{}) bool {
	modified := false

	// Recursively search for field names that contain the pattern
	if findAndUpdateWildcardPattern(data, pattern, value, "") {
		modified = true
	}

	return modified
}

// findAndUpdateWildcardPattern recursively searches for field names containing the pattern and updates their values
func findAndUpdateWildcardPattern(data *interface{}, pattern string, value interface{}, currentPath string) bool {
	if pattern == "" {
		return false
	}

	modified := false

	switch currentData := (*data).(type) {
	case map[string]interface{}:
		// Check all keys in the current map for pattern matches
		for key, v := range currentData {
			// Check if this key contains the pattern
			if strings.Contains(key, pattern) {
				log.Printf("[DEBUG] Found wildcard pattern match: field '%s' contains pattern '%s' at path '%s'", key, pattern, currentPath)
				currentData[key] = value
				modified = true
			}

			// Recursively search nested structures
			newPath := currentPath
			if newPath != "" {
				newPath += "."
			}
			newPath += key

			if findAndUpdateWildcardPattern(&v, pattern, value, newPath) {
				modified = true
			}
		}

	case map[interface{}]interface{}:
		// Convert to string keys and search
		stringMap := make(map[string]interface{})
		for k, v := range currentData {
			if keyStr, ok := k.(string); ok {
				stringMap[keyStr] = v
			}
		}

		// Check all keys for pattern matches
		for key, v := range stringMap {
			if strings.Contains(key, pattern) {
				log.Printf("[DEBUG] Found wildcard pattern match: field '%s' contains pattern '%s' at path '%s'", key, pattern, currentPath)
				stringMap[key] = value
				*data = stringMap
				modified = true
			}

			// Recursively search nested structures
			newPath := currentPath
			if newPath != "" {
				newPath += "."
			}
			newPath += key

			if findAndUpdateWildcardPattern(&v, pattern, value, newPath) {
				modified = true
			}
		}

	case []interface{}:
		// Search in array elements
		for i := range currentData {
			arrayPath := fmt.Sprintf("%s[%d]", currentPath, i)
			if findAndUpdateWildcardPattern(&currentData[i], pattern, value, arrayPath) {
				modified = true
			}
		}
	}

	return modified
}

// applyTemplateSubstitution replaces template variables in the YAML content
func applyTemplateSubstitution(yamlContent string, templateVars map[string]interface{}) string {
	result := yamlContent

	for varName, value := range templateVars {
		valueStr := value.(string) // Type assertion safe since we control input

		// Support multiple template syntaxes
		patterns := []string{
			"${" + varName + "}",  // ${var_name}
			"{{" + varName + "}}", // {{var_name}}
			"$" + varName,         // $var_name (simple)
		}

		for _, pattern := range patterns {
			if strings.Contains(result, pattern) {
				log.Printf("[DEBUG] Replacing template variable: %s -> %s", pattern, valueStr)
				result = strings.ReplaceAll(result, pattern, valueStr)
			}
		}
	}

	return result
}

// extractDocumentKind extracts the 'kind' field from a YAML document
func extractDocumentKind(data interface{}) string {
	if dataMap, ok := data.(map[string]interface{}); ok {
		if kind, exists := dataMap["kind"]; exists {
			if kindStr, ok := kind.(string); ok {
				return kindStr
			}
		}
	}
	if dataMap, ok := data.(map[interface{}]interface{}); ok {
		if kind, exists := dataMap["kind"]; exists {
			if kindStr, ok := kind.(string); ok {
				return kindStr
			}
		}
	}
	return ""
}

// parseDocumentSpecificPath splits document-specific path into kind and path components
// Examples: "Cluster.metadata.labels" -> ("Cluster", "metadata.labels")
//
//	"metadata.labels" -> ("", "metadata.labels")
func parseDocumentSpecificPath(path string) (string, string) {
	parts := strings.SplitN(path, ".", 2)
	if len(parts) == 2 {
		// Check if first part looks like a Kubernetes kind (starts with uppercase)
		if len(parts[0]) > 0 && parts[0][0] >= 'A' && parts[0][0] <= 'Z' {
			return parts[0], parts[1]
		}
	}
	// Not document-specific, return empty kind and full path
	return "", path
}

// applyOverrideToDocumentWithKind applies a single override to a YAML document with kind matching
func applyOverrideToDocumentWithKind(data *interface{}, path string, value interface{}, documentKind string) bool {
	// Parse document-specific path
	targetKind, actualPath := parseDocumentSpecificPath(path)

	// If target kind is specified and doesn't match document kind, skip
	if targetKind != "" && targetKind != documentKind {
		return false
	}

	// Apply the override using the actual path (without kind prefix)
	pathParts := parsePath(actualPath)
	if len(pathParts) == 0 {
		return false
	}

	// Convert string value to appropriate type (TypeMap only supports strings)
	convertedValue := convertStringToAppropriateType(value.(string))

	return setValueAtPath(data, pathParts, convertedValue)
}

// applyOverrideToDocument applies a single override to a YAML document (legacy function for backward compatibility)
func applyOverrideToDocument(data *interface{}, path string, value interface{}) bool {
	return applyOverrideToDocumentWithKind(data, path, value, "")
}

// parsePath splits a dot-notation path into components, handling array indices
func parsePath(path string) []string {
	// Replace array notation [index] with .index for consistent splitting
	path = strings.ReplaceAll(path, "[", ".")
	path = strings.ReplaceAll(path, "]", "")

	parts := strings.Split(path, ".")
	var result []string
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// convertOverrideValueFromInterface handles values that come as interface{} from Terraform
func convertOverrideValueFromInterface(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		// If it's a string, try to parse it as JSON for complex types
		return convertStringToAppropriateType(v)
	case bool, int, int64, float64:
		// Native types are returned as-is
		return v
	case []interface{}, map[string]interface{}:
		// Complex types are returned as-is
		return v
	default:
		// Fallback to string representation
		return fmt.Sprintf("%v", v)
	}
}

// convertStringToAppropriateType tries to convert a string to a more appropriate type
func convertStringToAppropriateType(value string) interface{} {
	value = strings.TrimSpace(value)

	// Try to parse as JSON first (for arrays and objects)
	var jsonValue interface{}
	if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
		return jsonValue
	}

	// Try boolean
	if boolVal, err := strconv.ParseBool(value); err == nil {
		return boolVal
	}

	// Try integer
	if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
		return intVal
	}

	// Try float
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	// Return as string if nothing else works
	return value
}

// convertOverrideValue converts a string value to its appropriate type (kept for backward compatibility)
func convertOverrideValue(value string) (interface{}, error) {
	return convertStringToAppropriateType(value), nil
}

// setValueAtPath sets a value at the specified path in a nested structure
func setValueAtPath(data *interface{}, pathParts []string, value interface{}) bool {
	if len(pathParts) == 0 {
		return false
	}

	current := *data

	for i, part := range pathParts {
		isLast := i == len(pathParts)-1

		switch currentData := current.(type) {
		case map[string]interface{}:
			if isLast {
				currentData[part] = value
				return true
			}

			// Create nested structure if it doesn't exist
			if _, exists := currentData[part]; !exists {
				// Determine if next part is an array index
				if i+1 < len(pathParts) {
					if _, err := strconv.Atoi(pathParts[i+1]); err == nil {
						currentData[part] = make([]interface{}, 0)
					} else {
						currentData[part] = make(map[string]interface{})
					}
				} else {
					currentData[part] = make(map[string]interface{})
				}
			}
			current = currentData[part]

		case map[interface{}]interface{}:
			// Convert to map[string]interface{} for easier handling
			stringMap := make(map[string]interface{})
			for k, v := range currentData {
				if keyStr, ok := k.(string); ok {
					stringMap[keyStr] = v
				}
			}

			if isLast {
				stringMap[part] = value
				*data = stringMap
				return true
			}

			if _, exists := stringMap[part]; !exists {
				if i+1 < len(pathParts) {
					if _, err := strconv.Atoi(pathParts[i+1]); err == nil {
						stringMap[part] = make([]interface{}, 0)
					} else {
						stringMap[part] = make(map[string]interface{})
					}
				} else {
					stringMap[part] = make(map[string]interface{})
				}
			}
			current = stringMap[part]

		case []interface{}:
			index, err := strconv.Atoi(part)
			if err != nil {
				return false
			}

			// Extend array if necessary
			for len(currentData) <= index {
				currentData = append(currentData, nil)
			}

			if isLast {
				currentData[index] = value
				return true
			}

			// Create nested structure if it doesn't exist
			if currentData[index] == nil {
				if i+1 < len(pathParts) {
					if _, err := strconv.Atoi(pathParts[i+1]); err == nil {
						currentData[index] = make([]interface{}, 0)
					} else {
						currentData[index] = make(map[string]interface{})
					}
				}
			}
			current = currentData[index]

		default:
			// Cannot traverse further
			return false
		}
	}

	return false
}

func toCustomClusterConfig(d *schema.ResourceData) *models.V1CustomClusterConfigEntity {
	customClusterConfig := &models.V1CustomClusterConfigEntity{
		Location:                toClusterLocationConfigs(d),
		MachineManagementConfig: toMachineManagementConfig(d),
		Resources:               toClusterResourceConfig(d),
	}

	return customClusterConfig
}

func toMachinePoolCustomCloud(machinePool interface{}) *models.V1CustomMachinePoolConfigEntity {
	mp := &models.V1CustomMachinePoolConfigEntity{}
	node := machinePool.(map[string]interface{})
	controlPlane, _ := node["control_plane"].(bool)
	controlPlaneAsWorker, _ := node["control_plane_as_worker"].(bool)

	// Get node pool config YAML
	nodePoolConfigYaml := strings.TrimSpace(node["node_pool_config"].(string))

	log.Printf("[DEBUG] === MACHINE POOL OVERRIDE PROCESSING ===")
	log.Printf("[DEBUG] Original node pool config YAML length: %d", len(nodePoolConfigYaml))
	log.Printf("[DEBUG] Original YAML preview: %s", nodePoolConfigYaml[:min(300, len(nodePoolConfigYaml))])

	// Check if overrides exist in the node map
	log.Printf("[DEBUG] Checking for overrides in node map. Keys: %v", getMapKeys(node))
	if overrideValue, exists := node["overrides"]; exists {
		log.Printf("[DEBUG] Raw overrides value: %v (type: %T)", overrideValue, overrideValue)
	} else {
		log.Printf("[DEBUG] No 'overrides' key found in node map")
	}

	// Apply overrides if they exist
	if overrides, ok := node["overrides"].(map[string]interface{}); ok && len(overrides) > 0 {
		log.Printf("[DEBUG] Successfully cast overrides. Applying %d YAML overrides to node pool config", len(overrides))
		for path, value := range overrides {
			log.Printf("[DEBUG] Node pool override: %s = %v (type: %T)", path, value, value)
		}

		processedYaml, err := applyYamlOverridesWithTemplates(nodePoolConfigYaml, overrides)
		if err != nil {
			log.Printf("[ERROR] Failed to apply YAML overrides to node pool config: %v", err)
		} else {
			log.Printf("[DEBUG] Node pool YAML transformation successful. Original length: %d, New length: %d", len(nodePoolConfigYaml), len(processedYaml))
			if len(processedYaml) > 0 && processedYaml != nodePoolConfigYaml {
				log.Printf("[DEBUG] Node pool YAML values updated with overrides")
				log.Printf("[DEBUG] Processed YAML preview: %s", processedYaml[:min(300, len(processedYaml))])
			} else {
				log.Printf("[DEBUG] WARNING: Processed YAML is identical to original - overrides may not have applied")
			}
			nodePoolConfigYaml = processedYaml
		}
	} else {
		log.Printf("[DEBUG] No overrides found for node pool config or failed to cast to map[string]interface{}")
		if overrideValue, exists := node["overrides"]; exists {
			log.Printf("[DEBUG] Override value exists but type assertion failed. Type: %T, Value: %v", overrideValue, overrideValue)
		}
	}

	// Normalize the final YAML content to ensure consistent formatting
	nodePoolConfigYaml = NormalizeYamlContent(nodePoolConfigYaml)
	log.Printf("[DEBUG] Final normalized node pool YAML length: %d", len(nodePoolConfigYaml))

	mp.CloudConfig = &models.V1CustomMachinePoolCloudConfigEntity{
		Values: nodePoolConfigYaml,
	}
	mp.PoolConfig = &models.V1CustomMachinePoolBaseConfigEntity{
		IsControlPlane:          controlPlane,
		UseControlPlaneAsWorker: controlPlaneAsWorker,
	}
	return mp
}

func flattenMachinePoolConfigsCustomCloud(machinePools []*models.V1CustomMachinePoolConfig) []interface{} {
	if len(machinePools) == 0 {
		return make([]interface{}, 0)
	}
	mps := make([]interface{}, len(machinePools))

	for i, machinePool := range machinePools {
		mp := make(map[string]interface{})
		mp["control_plane_as_worker"] = machinePool.UseControlPlaneAsWorker
		mp["control_plane"] = machinePool.IsControlPlane
		mp["node_pool_config"] = machinePool.Values
		mp["name"] = machinePool.Name
		mp["count"] = machinePool.Size
		mps[i] = mp
	}

	return mps
}

func flattenMachinePoolConfigsCustomCloudWithOverrides(machinePools []*models.V1CustomMachinePoolConfig, d *schema.ResourceData) []interface{} {
	if len(machinePools) == 0 {
		return make([]interface{}, 0)
	}

	// Get current machine pool configuration from state
	currentMachinePools := d.Get("machine_pool").([]interface{})
	currentMPMap := make(map[string]map[string]interface{})

	for _, mp := range currentMachinePools {
		if mpMap, ok := mp.(map[string]interface{}); ok {
			if name, exists := mpMap["name"]; exists {
				currentMPMap[name.(string)] = mpMap
			}
		}
	}

	mps := make([]interface{}, len(machinePools))

	for i, machinePool := range machinePools {
		mp := make(map[string]interface{})
		mp["control_plane_as_worker"] = machinePool.UseControlPlaneAsWorker
		mp["control_plane"] = machinePool.IsControlPlane
		mp["name"] = machinePool.Name
		mp["count"] = machinePool.Size

		// Handle node_pool_config with override reconciliation
		apiNodePoolConfig := machinePool.Values

		// Get current configuration for this machine pool
		if currentMP, exists := currentMPMap[machinePool.Name]; exists {
			var currentNodePoolConfig string
			var currentOverrides map[string]interface{}

			if config, exists := currentMP["node_pool_config"]; exists {
				currentNodePoolConfig = config.(string)
			}
			if overrides, exists := currentMP["overrides"]; exists {
				if overridesMap, ok := overrides.(map[string]interface{}); ok {
					currentOverrides = overridesMap
				}
			}

			if currentNodePoolConfig != "" && len(currentOverrides) > 0 {
				// Apply current overrides to current config to get expected result
				expectedConfig, err := applyYamlOverridesWithTemplates(currentNodePoolConfig, currentOverrides)
				if err != nil {
					log.Printf("[DEBUG] Failed to apply overrides for machine pool state comparison: %v", err)
					nodeConfig := NormalizeYamlContent(apiNodePoolConfig)
					mp["node_pool_config"] = nodeConfig
					if isMultiLineYAML(nodeConfig) {
						log.Printf("[INFO] Machine pool '%s' contains multi-line YAML. Consider using heredoc syntax (<<EOT...EOT) for better readability after import.", machinePool.Name)
					}
				} else {
					// Normalize both for comparison
					expectedNormalized := NormalizeYamlContent(expectedConfig)
					apiNormalized := NormalizeYamlContent(apiNodePoolConfig)

					if expectedNormalized == apiNormalized {
						// No drift detected - use normalized original config (before overrides)
						log.Printf("[DEBUG] No drift detected for machine pool %s, preserving original configuration", machinePool.Name)
						nodeConfig := NormalizeYamlContent(currentNodePoolConfig)
						mp["node_pool_config"] = nodeConfig
						if isMultiLineYAML(nodeConfig) {
							log.Printf("[INFO] Machine pool '%s' contains multi-line YAML. Consider using heredoc syntax (<<EOT...EOT) for better readability after import.", machinePool.Name)
						}
					} else {
						// Drift detected - use normalized API config
						log.Printf("[DEBUG] Drift detected for machine pool %s, using API config", machinePool.Name)
						nodeConfig := NormalizeYamlContent(apiNodePoolConfig)
						mp["node_pool_config"] = nodeConfig
						if isMultiLineYAML(nodeConfig) {
							log.Printf("[INFO] Machine pool '%s' contains multi-line YAML. Consider using heredoc syntax (<<EOT...EOT) for better readability after import.", machinePool.Name)
						}
					}
				}

				// Always preserve overrides from current state
				mp["overrides"] = currentOverrides
			} else {
				// No overrides, use normalized API config directly
				nodeConfig := NormalizeYamlContent(apiNodePoolConfig)
				mp["node_pool_config"] = nodeConfig
				if isMultiLineYAML(nodeConfig) {
					log.Printf("[INFO] Machine pool '%s' contains multi-line YAML. Consider using heredoc syntax (<<EOT...EOT) for better readability after import.", machinePool.Name)
				}

				// Preserve any existing overrides from state
				if overrides, exists := currentMP["overrides"]; exists {
					mp["overrides"] = overrides
				}
			}
		} else {
			// New machine pool or not found in current state - use normalized API config
			nodeConfig := NormalizeYamlContent(apiNodePoolConfig)
			mp["node_pool_config"] = nodeConfig
			if isMultiLineYAML(nodeConfig) {
				log.Printf("[INFO] Machine pool '%s' contains multi-line YAML. Consider using heredoc syntax (<<EOT...EOT) for better readability after import.", machinePool.Name)
			}
		}

		mps[i] = mp
	}

	return mps
}

func flattenCloudConfigCustom(configUID string, d *schema.ResourceData, c *client.V1Client) (diag.Diagnostics, bool) {
	cloudType := d.Get("cloud").(string)
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err), true
	}

	if err := ReadCommonAttributes(d); err != nil {
		return diag.FromErr(err), true
	}
	if config, err := c.GetCloudConfigCustomCloud(configUID, cloudType); err != nil {
		return diag.FromErr(err), true
	} else {
		if config.Spec != nil && config.Spec.CloudAccountRef != nil {
			if err := d.Set("cloud_account_id", config.Spec.CloudAccountRef.UID); err != nil {
				return diag.FromErr(err), true
			}
		}
		if err := d.Set("cloud_config", flattenCloudConfigsValuesCustomCloudWithOverrides(config, d)); err != nil {
			return diag.FromErr(err), true
		}
		if err := d.Set("machine_pool", flattenMachinePoolConfigsCustomCloudWithOverrides(config.Spec.MachinePoolConfig, d)); err != nil {
			return diag.FromErr(err), true
		}
	}

	return nil, false
}

func flattenCloudConfigsValuesCustomCloud(config *models.V1CustomCloudConfig) []interface{} {
	if config == nil || config.Spec == nil || config.Spec.ClusterConfig == nil {
		return make([]interface{}, 0)
	}

	m := make(map[string]interface{})

	if String(config.Spec.ClusterConfig.Values) != "" {
		m["values"] = String(config.Spec.ClusterConfig.Values)
	}

	return []interface{}{m}
}

func flattenCloudConfigsValuesCustomCloudWithOverrides(config *models.V1CustomCloudConfig, d *schema.ResourceData) []interface{} {
	if config == nil || config.Spec == nil || config.Spec.ClusterConfig == nil {
		return make([]interface{}, 0)
	}

	m := make(map[string]interface{})

	// Get the current configuration values and overrides
	currentConfig := d.Get("cloud_config").([]interface{})
	var currentValues string
	var currentOverrides map[string]interface{}

	if len(currentConfig) > 0 {
		if currentConfigMap, ok := currentConfig[0].(map[string]interface{}); ok {
			if values, exists := currentConfigMap["values"]; exists {
				currentValues = values.(string)
			}
			if overrides, exists := currentConfigMap["overrides"]; exists {
				if overridesMap, ok := overrides.(map[string]interface{}); ok {
					currentOverrides = overridesMap
				}
			}
		}
	}

	// Get the actual values from API
	apiValues := String(config.Spec.ClusterConfig.Values)

	var finalValues string
	if currentValues != "" && len(currentOverrides) > 0 {
		// Apply current overrides to current config values to get expected result
		expectedValues, err := applyYamlOverridesWithTemplates(currentValues, currentOverrides)
		if err != nil {
			log.Printf("[DEBUG] Failed to apply overrides for state comparison: %v", err)
			// Fall back to using normalized API values
			finalValues = NormalizeYamlContent(apiValues)
		} else {
			// Normalize both for comparison
			expectedNormalized := NormalizeYamlContent(expectedValues)
			apiNormalized := NormalizeYamlContent(apiValues)

			if expectedNormalized == apiNormalized {
				// No drift detected - use normalized original values (before overrides)
				log.Printf("[DEBUG] No drift detected, preserving original configuration values")
				finalValues = NormalizeYamlContent(currentValues)
			} else {
				// Drift detected - use normalized API values
				log.Printf("[DEBUG] Drift detected, using API values")
				finalValues = NormalizeYamlContent(apiValues)
			}
		}

		// Always preserve overrides from current state
		m["overrides"] = currentOverrides
	} else {
		// No overrides, use normalized API values directly
		finalValues = NormalizeYamlContent(apiValues)

		// Preserve any existing overrides from state
		if len(currentConfig) > 0 {
			if currentConfigMap, ok := currentConfig[0].(map[string]interface{}); ok {
				if overrides, exists := currentConfigMap["overrides"]; exists {
					m["overrides"] = overrides
				}
			}
		}
	}

	// Check if this is multi-line YAML that would benefit from heredoc formatting
	if isMultiLineYAML(finalValues) {
		// Add the values with a hint that heredoc would be better for readability
		m["values"] = finalValues
		// Note: We could add metadata here but Terraform generate-config-out will still escape it
		log.Printf("[INFO] Generated configuration contains multi-line YAML. Consider using heredoc syntax (<<EOT...EOT) for better readability after import.")
	} else {
		m["values"] = finalValues
	}

	return []interface{}{m}
}

// isMultiLineYAML checks if the content is multi-line YAML that would benefit from heredoc formatting
func isMultiLineYAML(content string) bool {
	// Check for multiple lines and YAML document separators or typical YAML structure
	lines := strings.Split(strings.TrimSpace(content), "\n")
	return len(lines) > 5 && (strings.Contains(content, "---") || strings.Contains(content, "apiVersion:"))
}

// Helper function to get map keys for debugging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
