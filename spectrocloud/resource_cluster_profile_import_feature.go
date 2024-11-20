package spectrocloud

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceClusterProfileImportFeature() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterProfileImportFeatureCreate,
		ReadContext:   resourceClusterProfileImportFeatureRead,
		UpdateContext: resourceClusterProfileImportFeatureUpdate,
		DeleteContext: resourceClusterProfileImportFeatureDelete,

		Schema: map[string]*schema.Schema{
			"import_file": {
				Type:     schema.TypeString,
				Required: true,
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant", "system"}, false),
				Description: "Allowed values are `project`, `tenant` or `system`. " +
					"Defaults to `project`. " + PROJECT_NAME_NUANCE,
			},
		},
	}
}

// implement the resource functions
func resourceClusterProfileImportFeatureCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ProfileContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, ProfileContext)

	importFile, err := toClusterProfileImportCreate(d)
	if err != nil {
		return diag.FromErr(err)
	}

	uid, err := c.CreateClusterProfileImport(importFile)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)
	return nil
}

func toClusterProfileImportCreate(d *schema.ResourceData) (*os.File, error) {
	importFilePath := d.Get("import_file").(string)

	importFile, err := os.Open(importFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening import file: %s", err)
	}
	/*defer func(importFile *os.File) {
		err := importFile.Close()
		if err != nil {
			fmt.Errorf("error closing import file: %s", err)
		}
	}(importFile)*/

	return importFile, nil
}

func resourceClusterProfileImportFeatureRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)
	_, err := c.ClusterProfileExport(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	//we don't want to set back the cluster profile, currently we're only supporting profile file name in schema not content.

	return nil
}

func resourceClusterProfileImportFeatureUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	importFile, err := toClusterProfileImportCreate(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Call the API endpoint to delete the cluster profile import resource
	err = c.DeleteClusterProfile(d.Id())
	if err != nil {
		// Return an error if the API call fails
		return diag.FromErr(err)
	}

	uid, err := c.CreateClusterProfileImport(importFile)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	return nil
}

func resourceClusterProfileImportFeatureDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Call the API endpoint to delete the cluster profile import resource
	if err := c.DeleteClusterProfile(d.Id()); err != nil {
		// Return an error if the API call fails
		return diag.FromErr(err)
	}

	// Set the ID to an empty string to indicate that the resource has been deleted
	d.SetId("")

	return nil
}
