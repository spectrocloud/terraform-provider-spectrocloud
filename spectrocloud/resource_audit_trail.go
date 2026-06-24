package spectrocloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

const (
	auditTrailTypeCloudWatch = "cloudwatch"
	auditTrailTypeSplunk     = "splunk"
)

func resourceAuditTrail() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAuditTrailCreate,
		ReadContext:   resourceAuditTrailRead,
		UpdateContext: resourceAuditTrailUpdate,
		DeleteContext: resourceAuditTrailDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAuditTrailImport,
		},
		Description: "Resource for managing tenant audit trail data sinks (CloudWatch or Splunk) in Spectro Cloud.",
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		SchemaVersion: 2,
		CustomizeDiff: resourceAuditTrailCustomizeDiff,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Human-readable name for the audit trail.",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{auditTrailTypeCloudWatch, auditTrailTypeSplunk}, false),
				Description:  "Audit trail sink type. Allowed values are `cloudwatch` or `splunk`.",
			},
			"cloudwatch": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "CloudWatch audit trail configuration. Required when `type` is `cloudwatch`.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "CloudWatch log group name.",
						},
						"region": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "AWS region for CloudWatch.",
						},
						"stream": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Optional CloudWatch log stream name.",
						},
						"credential_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "secret",
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice([]string{"secret", "sts"}, false),
							Description:  "AWS credential type. Allowed values are `secret` or `sts`. Default is `secret`.",
						},
						"access_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "AWS access key. Required when `credential_type` is `secret`.",
						},
						"secret_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "AWS secret key. Required when `credential_type` is `secret`.",
						},
						"arn": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "IAM role ARN. Required when `credential_type` is `sts`.",
						},
						"external_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "External ID for STS role assumption. Used with `credential_type` `sts`.",
						},
						"partition": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "aws",
							ValidateFunc: validation.StringInSlice([]string{"aws", "aws-us-gov"}, false),
							Description:  "AWS partition. Allowed values are `aws` or `aws-us-gov`. Default is `aws`.",
						},
					},
				},
			},
			"splunk": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Splunk HEC audit trail configuration. Required when `type` is `splunk`.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hec_url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Splunk HTTP Event Collector (HEC) URL.",
						},
						"token": {
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
							Description: "Splunk HEC token.",
						},
						"index": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Optional Splunk index. Uses the token default when empty.",
						},
						"source": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Optional Splunk source. Uses the token default when empty.",
						},
						"tls_config": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Optional TLS configuration for Splunk HEC.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ca_cert_base64": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Base64-encoded CA certificate for self-signed Splunk instances.",
									},
									"tls_verification": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether TLS certificate verification is enabled. Computed as the inverse of `insecure_skip_verify`.",
									},
									"insecure_skip_verify": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     true,
										Description: "Skip TLS certificate verification when set to `true`. Default is `false`.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceAuditTrailCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	auditType := diff.Get("type").(string)
	hasCloudWatch := len(diff.Get("cloudwatch").([]interface{})) > 0
	hasSplunk := len(diff.Get("splunk").([]interface{})) > 0

	switch auditType {
	case auditTrailTypeCloudWatch:
		if !hasCloudWatch {
			return fmt.Errorf("`cloudwatch` block is required when `type` is `cloudwatch`")
		}
		if hasSplunk {
			return fmt.Errorf("`splunk` block must not be set when `type` is `cloudwatch`")
		}
		cw := diff.Get("cloudwatch").([]interface{})[0].(map[string]interface{})
		credentialType := cw["credential_type"].(string)
		if credentialType == "" {
			credentialType = "secret"
		}
		switch credentialType {
		case "secret":
			if cw["access_key"].(string) == "" || cw["secret_key"].(string) == "" {
				return fmt.Errorf("`access_key` and `secret_key` are required when `credential_type` is `secret`")
			}
		case "sts":
			if cw["arn"].(string) == "" {
				return fmt.Errorf("`arn` is required when `credential_type` is `sts`")
			}
		}
	case auditTrailTypeSplunk:
		if !hasSplunk {
			return fmt.Errorf("`splunk` block is required when `type` is `splunk`")
		}
		if hasCloudWatch {
			return fmt.Errorf("`cloudwatch` block must not be set when `type` is `splunk`")
		}
		sp := diff.Get("splunk").([]interface{})[0].(map[string]interface{})
		if sp["hec_url"].(string) == "" || sp["token"].(string) == "" {
			return fmt.Errorf("`hec_url` and `token` are required in the `splunk` block")
		}
	}
	return nil
}

func resourceAuditTrailCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}

	auditType := d.Get("type").(string)
	switch auditType {
	case auditTrailTypeCloudWatch:
		if err := validateCloudWatchAuditTrail(d, c); err != nil {
			return diag.FromErr(err)
		}
		config, err := toCloudWatchDataSinkConfig(d, "")
		if err != nil {
			return diag.FromErr(err)
		}
		uid, err := c.CreateCloudWatchAuditTrail(tenantUID, config)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(uid)
	case auditTrailTypeSplunk:
		if err := validateSplunkAuditTrail(d, c, tenantUID); err != nil {
			return diag.FromErr(err)
		}
		entity, err := toSplunkSinkEntity(d)
		if err != nil {
			return diag.FromErr(err)
		}
		uid, err := c.CreateSplunkAuditTrail(tenantUID, entity)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(uid)
	default:
		return diag.Errorf("unsupported audit trail type: %s", auditType)
	}

	return resourceAuditTrailRead(ctx, d, m)
}

func resourceAuditTrailRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics

	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return handleReadError(d, err, diags)
	}

	auditType := d.Get("type").(string)
	switch auditType {
	case auditTrailTypeCloudWatch:
		config, err := c.GetCloudWatchAuditTrail(tenantUID)
		if err != nil {
			return handleReadError(d, err, diags)
		}
		if config == nil {
			d.SetId("")
			return diags
		}
		if err := flattenCloudWatchAuditTrail(d, config); err != nil {
			return diag.FromErr(err)
		}
	case auditTrailTypeSplunk:
		sink, err := c.GetSplunkAuditTrail(tenantUID, d.Id())
		if err != nil {
			return handleReadError(d, err, diags)
		}
		if sink == nil {
			d.SetId("")
			return diags
		}
		if err := flattenSplunkAuditTrail(d, sink); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceAuditTrailUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}

	auditType := d.Get("type").(string)
	switch auditType {
	case auditTrailTypeCloudWatch:
		if err := validateCloudWatchAuditTrail(d, c); err != nil {
			return diag.FromErr(err)
		}
		config, err := toCloudWatchDataSinkConfig(d, d.Id())
		if err != nil {
			return diag.FromErr(err)
		}
		if err := c.UpdateCloudWatchAuditTrail(tenantUID, config); err != nil {
			return diag.FromErr(err)
		}
	case auditTrailTypeSplunk:
		if err := validateSplunkAuditTrail(d, c, tenantUID); err != nil {
			return diag.FromErr(err)
		}
		updateEntity, err := toSplunkSinkEntity(d)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := c.UpdateSplunkAuditTrail(tenantUID, d.Id(), updateEntity); err != nil {
			return diag.FromErr(err)
		}
	default:
		return diag.Errorf("unsupported audit trail type: %s", auditType)
	}

	return resourceAuditTrailRead(ctx, d, m)
}

func resourceAuditTrailDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics

	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}

	auditType := d.Get("type").(string)
	switch auditType {
	case auditTrailTypeCloudWatch:
		if err := c.DeleteCloudWatchAuditTrail(tenantUID); err != nil {
			return diag.FromErr(err)
		}
	case auditTrailTypeSplunk:
		if err := c.DeleteSplunkAuditTrail(tenantUID, d.Id()); err != nil {
			return diag.FromErr(err)
		}
	default:
		return diag.Errorf("unsupported audit trail type: %s", auditType)
	}

	d.SetId("")
	return diags
}

func resourceAuditTrailImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, tenantString)
	uid := d.Id()
	if uid == "" {
		return nil, fmt.Errorf("audit trail import ID is required")
	}

	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return nil, err
	}

	sink, err := c.GetSplunkAuditTrail(tenantUID, uid)
	if err != nil {
		return nil, err
	}
	if sink != nil {
		if err := d.Set("type", auditTrailTypeSplunk); err != nil {
			return nil, err
		}
		diags := resourceAuditTrailRead(ctx, d, m)
		if diags.HasError() {
			return nil, fmt.Errorf("could not read splunk audit trail for import: %v", diags)
		}
		return []*schema.ResourceData{d}, nil
	}

	config, err := c.GetCloudWatchAuditTrail(tenantUID)
	if err != nil {
		return nil, err
	}
	if config == nil || config.Metadata == nil || config.Metadata.UID != uid {
		return nil, fmt.Errorf("audit trail with id '%s' not found", uid)
	}
	if err := d.Set("type", auditTrailTypeCloudWatch); err != nil {
		return nil, err
	}
	diags := resourceAuditTrailRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read cloudwatch audit trail for import: %v", diags)
	}
	return []*schema.ResourceData{d}, nil
}

func priorCloudWatchBlockString(d *schema.ResourceData, field string) string {
	raw := d.Get("cloudwatch")
	if raw == nil {
		return ""
	}
	lst, ok := raw.([]interface{})
	if !ok || len(lst) == 0 {
		return ""
	}
	m, ok := lst[0].(map[string]interface{})
	if !ok || m == nil {
		return ""
	}
	v, ok := m[field].(string)
	if !ok {
		return ""
	}
	return v
}

func priorSplunkBlockString(d *schema.ResourceData, field string) string {
	raw := d.Get("splunk")
	if raw == nil {
		return ""
	}
	lst, ok := raw.([]interface{})
	if !ok || len(lst) == 0 {
		return ""
	}
	m, ok := lst[0].(map[string]interface{})
	if !ok || m == nil {
		return ""
	}
	v, ok := m[field].(string)
	if !ok {
		return ""
	}
	return v
}

func cloudWatchSecretKeyForRead(d *schema.ResourceData, apiSecret string) string {
	if v := priorCloudWatchBlockString(d, "secret_key"); v != "" {
		return v
	}
	if apiSecret != "" && !strings.Contains(apiSecret, "*") {
		return apiSecret
	}
	return ""
}

func cloudWatchExternalIDForRead(d *schema.ResourceData, apiExternalID string) string {
	if v := priorCloudWatchBlockString(d, "external_id"); v != "" {
		return v
	}
	if apiExternalID != "" && !strings.Contains(apiExternalID, "*") {
		return apiExternalID
	}
	return ""
}

func splunkTokenForRead(d *schema.ResourceData) string {
	return priorSplunkBlockString(d, "token")
}

func toCloudWatchValidateConfig(d *schema.ResourceData) (*models.V1CloudWatchConfig, error) {
	cwList := d.Get("cloudwatch").([]interface{})
	if len(cwList) == 0 {
		return nil, fmt.Errorf("cloudwatch block is required")
	}
	cw := cwList[0].(map[string]interface{})

	stream := ""
	if v, ok := cw["stream"].(string); ok {
		stream = v
	}

	return &models.V1CloudWatchConfig{
		Group:       cw["group"].(string),
		Region:      cw["region"].(string),
		Stream:      stream,
		Credentials: toCloudWatchCredentials(d),
	}, nil
}

func validateCloudWatchAuditTrail(d *schema.ResourceData, c *client.V1Client) error {
	config, err := toCloudWatchValidateConfig(d)
	if err != nil {
		return err
	}
	return c.ValidateCloudWatchAuditTrail(config)
}

func validateSplunkAuditTrail(d *schema.ResourceData, c *client.V1Client, tenantUID string) error {
	entity, err := toSplunkSinkEntity(d)
	if err != nil {
		return err
	}
	return c.ValidateSplunkAuditTrail(tenantUID, entity.Spec)
}

func toCloudWatchCredentials(d *schema.ResourceData) *models.V1AwsCloudAccount {
	cwList := d.Get("cloudwatch").([]interface{})
	cw := cwList[0].(map[string]interface{})
	credentialType := cw["credential_type"].(string)
	if credentialType == "" {
		credentialType = "secret"
	}

	credentials := &models.V1AwsCloudAccount{
		Partition: types.Ptr(cw["partition"].(string)),
	}
	if credentials.Partition == nil || *credentials.Partition == "" {
		credentials.Partition = types.Ptr("aws")
	}

	switch credentialType {
	case "sts":
		credentials.CredentialType = models.V1AwsCloudAccountCredentialTypeSts.Pointer()
		credentials.Sts = &models.V1AwsStsCredentials{
			Arn:        cw["arn"].(string),
			ExternalID: cw["external_id"].(string),
		}
	default:
		credentials.CredentialType = models.V1AwsCloudAccountCredentialTypeSecret.Pointer()
		credentials.AccessKey = cw["access_key"].(string)
		credentials.SecretKey = cw["secret_key"].(string)
	}
	return credentials
}

func toCloudWatchDataSinkConfig(d *schema.ResourceData, uid string) (*models.V1DataSinkConfig, error) {
	cwList := d.Get("cloudwatch").([]interface{})
	if len(cwList) == 0 {
		return nil, fmt.Errorf("cloudwatch block is required")
	}
	cw := cwList[0].(map[string]interface{})

	stream := ""
	if v, ok := cw["stream"].(string); ok {
		stream = v
	}

	config := &models.V1DataSinkConfig{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
		},
		Spec: &models.V1DataSinkSpec{
			AuditDataSinks: []*models.V1DataSinkableSpec{
				{
					Type: models.V1DataSinkableSpecTypeCloudwatch,
					CloudWatch: &models.V1CloudWatch{
						Group:       cw["group"].(string),
						Region:      cw["region"].(string),
						Stream:      stream,
						Credentials: toCloudWatchCredentials(d),
					},
				},
			},
		},
	}
	if uid != "" {
		config.Metadata.UID = uid
	}
	return config, nil
}

func toSplunkSinkEntity(d *schema.ResourceData) (*models.V1SplunkSinkEntity, error) {
	spList := d.Get("splunk").([]interface{})
	if len(spList) == 0 {
		return nil, fmt.Errorf("splunk block is required")
	}
	sp := spList[0].(map[string]interface{})

	token := sp["token"].(string)
	hecURL := sp["hec_url"].(string)
	tokenVal := strfmt.Password(token)

	spec := &models.V1SplunkSinkSpec{
		HecURL: types.Ptr(hecURL),
		Token:  &tokenVal,
		Index:  sp["index"].(string),
		Source: sp["source"].(string),
	}

	if tlsList, ok := sp["tls_config"].([]interface{}); ok && len(tlsList) > 0 {
		tls := tlsList[0].(map[string]interface{})
		insecureSkipVerify := false
		if v, ok := tls["insecure_skip_verify"].(bool); ok {
			insecureSkipVerify = v
		}
		spec.TLSConfig = &models.V1TLSCA{
			CaCertBase64:       tls["ca_cert_base64"].(string),
			InsecureSkipVerify: insecureSkipVerify,
			Enabled:            true,
		}
	}

	name := d.Get("name").(string)
	return &models.V1SplunkSinkEntity{
		Name: types.Ptr(name),
		Spec: spec,
	}, nil
}

func flattenCloudWatchAuditTrail(d *schema.ResourceData, config *models.V1DataSinkConfig) error {
	if config.Metadata != nil {
		if err := d.Set("name", config.Metadata.Name); err != nil {
			return err
		}
	}

	var cw *models.V1CloudWatch
	if config.Spec != nil {
		for _, sink := range config.Spec.AuditDataSinks {
			if sink != nil && sink.CloudWatch != nil {
				cw = sink.CloudWatch
				break
			}
		}
	}
	if cw == nil {
		return fmt.Errorf("cloudwatch audit trail configuration not found in API response")
	}

	cwMap := map[string]interface{}{
		"group":  cw.Group,
		"region": cw.Region,
		"stream": cw.Stream,
	}

	if cw.Credentials != nil {
		partition := "aws"
		if cw.Credentials.Partition != nil && *cw.Credentials.Partition != "" {
			partition = *cw.Credentials.Partition
		}
		cwMap["partition"] = partition

		if cw.Credentials.CredentialType != nil && *cw.Credentials.CredentialType == models.V1AwsCloudAccountCredentialTypeSts {
			cwMap["credential_type"] = "sts"
			if cw.Credentials.Sts != nil {
				cwMap["arn"] = cw.Credentials.Sts.Arn
				if externalID := cloudWatchExternalIDForRead(d, cw.Credentials.Sts.ExternalID); externalID != "" {
					cwMap["external_id"] = externalID
				}
			}
		} else {
			cwMap["credential_type"] = "secret"
			cwMap["access_key"] = cw.Credentials.AccessKey
			if secretKey := cloudWatchSecretKeyForRead(d, cw.Credentials.SecretKey); secretKey != "" {
				cwMap["secret_key"] = secretKey
			}
		}
	}

	return d.Set("cloudwatch", []interface{}{cwMap})
}

func flattenSplunkAuditTrail(d *schema.ResourceData, sink *models.V1SplunkSink) error {
	if sink.Metadata != nil {
		if err := d.Set("name", sink.Metadata.Name); err != nil {
			return err
		}
	}
	if sink.Spec == nil {
		return fmt.Errorf("splunk audit trail spec not found in API response")
	}

	hecURL := ""
	if sink.Spec.HecURL != nil {
		hecURL = *sink.Spec.HecURL
	}
	spMap := map[string]interface{}{
		"hec_url": hecURL,
		"index":   sink.Spec.Index,
		"source":  sink.Spec.Source,
	}
	if token := splunkTokenForRead(d); token != "" {
		spMap["token"] = token
	}

	if sink.Spec.TLSConfig != nil {
		insecureSkipVerify := sink.Spec.TLSConfig.InsecureSkipVerify
		spMap["tls_config"] = []interface{}{
			map[string]interface{}{
				"ca_cert_base64":       sink.Spec.TLSConfig.CaCertBase64,
				"insecure_skip_verify": insecureSkipVerify,
				"tls_verification":     !insecureSkipVerify,
			},
		}
	}

	return d.Set("splunk", []interface{}{spMap})
}
