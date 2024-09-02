package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProvider(t *testing.T) {
	p := New("111.111.111")() // test version

	err := p.InternalValidate()

	if err != nil {
		t.Fatal(err)
	}
}

func prepareBaseProviderConfig() *schema.ResourceData {
	basSchema := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Spectro Cloud API host url. Can also be set with the `SPECTROCLOUD_HOST` environment variable. Defaults to https://api.spectrocloud.com",
				DefaultFunc: schema.EnvDefaultFunc("SPECTROCLOUD_HOST", "api.spectrocloud.com"),
			},
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The Spectro Cloud API key. Can also be set with the `SPECTROCLOUD_APIKEY` environment variable.",
				DefaultFunc: schema.EnvDefaultFunc("SPECTROCLOUD_APIKEY", nil),
			},
			"trace": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Enable HTTP request tracing. Can also be set with the `SPECTROCLOUD_TRACE` environment variable. To enable Terraform debug logging, set `TF_LOG=DEBUG`. Visit the Terraform documentation to learn more about Terraform [debugging](https://developer.hashicorp.com/terraform/plugin/log/managing).",
				DefaultFunc: schema.EnvDefaultFunc("SPECTROCLOUD_TRACE", nil),
			},
			"retry_attempts": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Number of retry attempts. Can also be set with the `SPECTROCLOUD_RETRY_ATTEMPTS` environment variable. Defaults to 10.",
				DefaultFunc: schema.EnvDefaultFunc("SPECTROCLOUD_RETRY_ATTEMPTS", 10),
			},
			"project_name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Default",
				// cannot be empty
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The Palette project the provider will target. If no value is provided, the `Default` Palette project is used. The default value is `Default`.",
			},
			"ignore_insecure_tls_error": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Ignore insecure TLS errors for Spectro Cloud API endpoints. Defaults to false.",
			},
		},
	}

	d := basSchema.TestResourceData()
	_ = d.Set("host", "127.0.0.1:8080")
	_ = d.Set("project_name", "Default")
	_ = d.Set("ignore_insecure_tls_error", true)
	_ = d.Set("api_key", "12345")
	_ = d.Set("trace", true)
	_ = d.Set("retry_attempts", 2)
	return d
}

func TestProviderConfig(t *testing.T) {
	d := prepareBaseProviderConfig()
	_, diags := providerConfigure(context.Background(), d)
	assert.Empty(t, diags)
}

func TestProviderConfigValidError(t *testing.T) {
	d := prepareBaseProviderConfig()
	// validating empty api key use case
	_ = d.Set("api_key", "")
	_, diags := providerConfigure(context.Background(), d)
	assertFirstDiagMessage(t, diags, "Unable to create Spectro Cloud client")
}
