package spectrocloud

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client/herr"
)

func resourceRegistrationToken() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRegistrationTokenCreate,
		ReadContext:   resourceRegistrationTokenRead,
		UpdateContext: resourceRegistrationTokenUpdate,
		DeleteContext: resourceRegistrationTokenDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRegistrationTokenImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the registration token.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "A brief description of the registration token.",
			},
			"project_uid": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The unique identifier of the project associated with the registration token.",
			},
			"expiry_date": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`),
					"expiry_date must be in YYYY-MM-DD format",
				),
				Description: "The expiration date of the registration token in `YYYY-MM-DD` format.",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice([]string{"active", "inactive"}, false),
				Description:  "The status of the registration token. Allowed values are `active` or `inactive`. Default is `active`.",
			},
			"token": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func flattenRegistrationToken(d *schema.ResourceData, tokenEntity *models.V1EdgeToken) error {
	if err := d.Set("name", tokenEntity.Metadata.Name); err != nil {
		return err
	}
	if desc, exists := tokenEntity.Metadata.Annotations["description"]; exists {
		if err := d.Set("description", desc); err != nil {
			return err
		}
	}
	if tokenEntity.Spec.DefaultProject != nil {
		if err := d.Set("project_uid", tokenEntity.Spec.DefaultProject.UID); err != nil {
			return err
		}
	}

	dt := strfmt.DateTime(tokenEntity.Spec.Expiry)
	expDate := time.Time(dt).Format("2006-01-02")
	if err := d.Set("expiry_date", expDate); err != nil {
		return err
	}
	if err := d.Set("token", tokenEntity.Spec.Token); err != nil {
		return err
	}
	if err := d.Set("status", StateConvertBool(tokenEntity.Status.IsActive)); err != nil {
		return err
	}

	return nil
}

func StateConvertBool(isActive bool) string {
	if isActive {
		return "active"
	}
	return "inactive"
}

func stateConvertString(state string) bool {
	return state == "active"
}

func toRegistrationTokenCreate(d *schema.ResourceData) (*models.V1EdgeTokenEntity, error) {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	defaultProjectUID := d.Get("project_uid").(string)
	//expiry := d.Get("expiry_date").(string)

	//Parse string to time.Time
	parsedTime, err := time.Parse("2006-01-02", d.Get("expiry_date").(string))
	if err != nil {
		fmt.Println("Error parsing date:", err)
		return nil, err
	}
	// Convert to strfmt.DateTime
	expiry := strfmt.DateTime(parsedTime)

	return &models.V1EdgeTokenEntity{
		Metadata: &models.V1ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				"description": description,
			},
		},
		Spec: &models.V1EdgeTokenSpecEntity{
			DefaultProjectUID: defaultProjectUID,
			Expiry:            models.V1Time(expiry),
		},
	}, nil
}

func toRegistrationTokenUpdate(d *schema.ResourceData) (*models.V1EdgeTokenUpdate, error) {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	defaultProjectUID := d.Get("project_uid").(string)
	//expiry := d.Get("expiry_date").(string)

	//Parse string to time.Time
	parsedTime, err := time.Parse("2006-01-02", d.Get("expiry_date").(string))
	if err != nil {
		fmt.Println("Error parsing date:", err)
		return nil, err
	}
	// Convert to strfmt.DateTime
	expiry := strfmt.DateTime(parsedTime)

	return &models.V1EdgeTokenUpdate{
		Metadata: &models.V1ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				"description": description,
			},
			UID: d.Id(),
		},
		Spec: &models.V1EdgeTokenSpecUpdate{
			DefaultProjectUID: defaultProjectUID,
			Expiry:            models.V1Time(expiry),
		},
	}, nil
}

func resourceRegistrationTokenCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	tokenEntity, err := toRegistrationTokenCreate(d)
	if err != nil {
		return diag.FromErr(err)
	}
	uid, token, err := c.CreateRegistrationToken(tokenEntity.Metadata.Name, tokenEntity)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)
	err = d.Set("token", token)
	if err != nil {
		return diag.FromErr(err)
	}
	state := stateConvertString(d.Get("status").(string))
	err = c.UpdateRegistrationTokenState(d.Id(), &models.V1EdgeTokenActiveState{IsActive: state})
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceRegistrationTokenRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	regToken, err := c.GetRegistrationTokenByUID(d.Id())
	if err != nil {
		return handleReadError(d, err, diags)
	}
	err = flattenRegistrationToken(d, regToken)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceRegistrationTokenUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	if d.HasChange("status") {
		state := stateConvertString(d.Get("status").(string))
		err := c.UpdateRegistrationTokenState(d.Id(), &models.V1EdgeTokenActiveState{IsActive: state})
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChanges("name", "description", "expiry_date", "project_uid") {
		regUpdateEntity, err := toRegistrationTokenUpdate(d)
		if err != nil {
			return diag.FromErr(err)
		}
		err = c.UpdateRegistrationTokenByUID(d.Id(), regUpdateEntity)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceRegistrationTokenDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	err := c.DeleteRegistrationToken(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceRegistrationTokenImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, tenantString)

	idOrName := d.Id()
	if idOrName == "" {
		return nil, fmt.Errorf("registration token import ID or name is required")
	}

	// Try lookup by UID first
	tokenEntity, err := c.GetRegistrationTokenByUID(idOrName)
	if err != nil && !herr.IsNotFound(err) {
		return nil, fmt.Errorf("unable to retrieve registration token '%s': %w", idOrName, err)
	}

	// Fall back to lookup by name
	if err != nil || tokenEntity == nil {
		tokenEntity, err = c.GetRegistrationTokenByName(idOrName)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve registration token by name '%s': %w", idOrName, err)
		}
		if tokenEntity == nil {
			return nil, fmt.Errorf("registration token with id or name '%s' not found", idOrName)
		}
	}

	// Always set the ID to the resolved UID for stable resource identity
	d.SetId(tokenEntity.Metadata.UID)

	diags := resourceRegistrationTokenRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read registration token for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}
