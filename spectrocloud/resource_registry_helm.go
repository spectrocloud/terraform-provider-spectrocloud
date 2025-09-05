package spectrocloud

import (
	"context"
	"time"

	"github.com/go-openapi/strfmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func resourceRegistryHelm() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRegistryHelmCreate,
		ReadContext:   resourceRegistryHelmRead,
		UpdateContext: resourceRegistryHelmUpdate,
		DeleteContext: resourceRegistryHelmDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRegistryHelmImport,
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
				ForceNew:    true,
				Description: "The name of the Helm registry. This must be unique",
			},
			"is_private": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Specifies whether the Helm registry is private or public.",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The URL endpoint of the Helm registry where the charts are hosted.",
			},
			"credentials": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Authentication credentials for accessing the Helm registry.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"credential_type": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The type of authentication used for the Helm registry. Supported values are 'noAuth' for no authentication, 'basic' for username/password, and 'token' for token-based authentication.",
							ValidateFunc: validation.StringInSlice([]string{"noAuth", "basic", "token"}, false),
						},
						"username": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The username for basic authentication. Required if 'credential_type' is set to 'basic'.",
						},
						"password": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The password for basic authentication. Required if 'credential_type' is set to 'basic'.",
						},
						"token": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The authentication token. Required if 'credential_type' is set to 'token'.",
						},
					},
				},
			},
		},
	}
}

func resourceRegistryHelmCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	registry := toRegistryEntityHelm(d)
	uid, err := c.CreateHelmRegistry(registry)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)
	return diags
}

func resourceRegistryHelmRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	registry, err := c.GetHelmRegistry(d.Id())
	if err != nil {
		return handleReadError(d, err, diags)
	} else if registry == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	if err := d.Set("name", registry.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_private", registry.Spec.IsPrivate); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("endpoint", registry.Spec.Endpoint); err != nil {
		return diag.FromErr(err)
	}

	if registry.Spec.Auth.Type == "noAuth" {
		credentials := make([]interface{}, 0, 1)
		acc := make(map[string]interface{})
		acc["credential_type"] = "noAuth"
		credentials = append(credentials, acc)
		if err := d.Set("credentials", credentials); err != nil {
			return diag.FromErr(err)
		}
	} else if registry.Spec.Auth.Type == "basic" {
		credentials := make([]interface{}, 0, 1)
		acc := make(map[string]interface{})
		acc["credential_type"] = "basic"
		acc["username"] = registry.Spec.Auth.Username
		acc["password"] = registry.Spec.Auth.Password.String()
		credentials = append(credentials, acc)
		if err := d.Set("credentials", credentials); err != nil {
			return diag.FromErr(err)
		}
	} else if registry.Spec.Auth.Type == "token" {
		credentials := make([]interface{}, 0, 1)
		acc := make(map[string]interface{})
		acc["credential_type"] = "token"
		acc["username"] = registry.Spec.Auth.Username
		acc["token"] = registry.Spec.Auth.Token.String()
		credentials = append(credentials, acc)
		if err := d.Set("credentials", credentials); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceRegistryHelmUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	registry := toRegistryHelm(d)
	err := c.UpdateHelmRegistry(d.Id(), registry)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceRegistryHelmDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics
	err := c.DeleteHelmRegistry(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toRegistryEntityHelm(d *schema.ResourceData) *models.V1HelmRegistryEntity {
	endpoint := d.Get("endpoint").(string)
	isPrivate := d.Get("is_private").(bool)
	config := d.Get("credentials").([]interface{})[0].(map[string]interface{})
	return &models.V1HelmRegistryEntity{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
		},
		Spec: &models.V1HelmRegistrySpecEntity{
			Name:      d.Get("name").(string),
			Auth:      toRegistryHelmCredential(config),
			Endpoint:  &endpoint,
			IsPrivate: isPrivate,
		},
	}
}

func toRegistryHelm(d *schema.ResourceData) *models.V1HelmRegistry {
	endpoint := d.Get("endpoint").(string)
	isPrivate := d.Get("is_private").(bool)
	config := d.Get("credentials").([]interface{})[0].(map[string]interface{})
	return &models.V1HelmRegistry{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
		},
		Spec: &models.V1HelmRegistrySpec{
			Name:      d.Get("name").(string),
			Auth:      toRegistryHelmCredential(config),
			Endpoint:  &endpoint,
			IsPrivate: isPrivate,
		},
	}
}

func toRegistryHelmCredential(regCred map[string]interface{}) *models.V1RegistryAuth {
	auth := &models.V1RegistryAuth{
		Type: "noAuth",
	}

	if regCred["credential_type"].(string) == "basic" {
		auth.Type = "basic"
		auth.Username = regCred["username"].(string)
		auth.Password = strfmt.Password(regCred["password"].(string))
	} else if regCred["credential_type"].(string) == "token" {
		auth.Type = "token"
		auth.Username = regCred["username"].(string)
		auth.Token = strfmt.Password(regCred["token"].(string))
	}
	return auth
}
