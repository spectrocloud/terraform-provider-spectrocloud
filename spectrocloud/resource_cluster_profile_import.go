package spectrocloud

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceClusterProfileImport() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterProfileImportCreate,
		ReadContext:   resourceClusterProfileImportRead,
		UpdateContext: resourceClusterProfileImportUpdate,
		DeleteContext: resourceClusterProfileImportDelete,

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
			},
		},
	}
}

// implement the resource functions
func resourceClusterProfileImportCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	importFile, err := toClusterProfileImportCreate(d)
	if err != nil {
		return diag.FromErr(err)
	}

	ProfileContext := d.Get("context").(string)
	uid, err := c.CreateClusterProfileImport(importFile, ProfileContext)
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

func resourceClusterProfileImportRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	clusterProfile, err := c.ClusterProfileExport(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("import_file", clusterProfile)

	return nil
}

func resourceClusterProfileImportUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

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

	ProfileContext := d.Get("context").(string)
	uid, err := c.CreateClusterProfileImport(importFile, ProfileContext)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	return nil
}

func resourceClusterProfileImportDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Call the API endpoint to delete the cluster profile import resource
	err := c.DeleteClusterProfile(d.Id())
	if err != nil {
		// Return an error if the API call fails
		return diag.FromErr(err)
	}

	// Set the ID to an empty string to indicate that the resource has been deleted
	d.SetId("")

	return nil
}
