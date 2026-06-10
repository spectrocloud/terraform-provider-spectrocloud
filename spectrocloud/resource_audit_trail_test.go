package spectrocloud

import (
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToCloudWatchDataSinkConfig(t *testing.T) {
	t.Parallel()

	d := schema.TestResourceDataRaw(t, resourceAuditTrail().Schema, map[string]interface{}{
		"name": "test-cw",
		"type": auditTrailTypeCloudWatch,
		"cloudwatch": []interface{}{
			map[string]interface{}{
				"group":           "logs",
				"region":          "us-east-1",
				"stream":          "audit",
				"credential_type": "secret",
				"access_key":      "AKIATEST",
				"secret_key":      "secret",
				"partition":       "aws",
			},
		},
	})

	config, err := toCloudWatchDataSinkConfig(d, "uid-123")
	require.NoError(t, err)
	require.NotNil(t, config.Metadata)
	assert.Equal(t, "test-cw", config.Metadata.Name)
	assert.Equal(t, "uid-123", config.Metadata.UID)
	require.Len(t, config.Spec.AuditDataSinks, 1)
	assert.Equal(t, models.V1DataSinkableSpecTypeCloudwatch, config.Spec.AuditDataSinks[0].Type)
	assert.Equal(t, "logs", config.Spec.AuditDataSinks[0].CloudWatch.Group)
	assert.Equal(t, "AKIATEST", config.Spec.AuditDataSinks[0].CloudWatch.Credentials.AccessKey)
}

func TestToCloudWatchDataSinkConfigDevHubbleAudits(t *testing.T) {
	t.Parallel()

	d := schema.TestResourceDataRaw(t, resourceAuditTrail().Schema, map[string]interface{}{
		"name": "rag",
		"type": auditTrailTypeCloudWatch,
		"cloudwatch": []interface{}{
			map[string]interface{}{
				"group":           "dev-hubble-audits",
				"region":          "us-east-1",
				"credential_type": "secret",
				"access_key":      "AKIATEST",
				"secret_key":      "secret",
				"partition":       "aws",
			},
		},
	})

	config, err := toCloudWatchDataSinkConfig(d, "")
	require.NoError(t, err)
	require.NotNil(t, config.Metadata)
	assert.Equal(t, "rag", config.Metadata.Name)
	require.Len(t, config.Spec.AuditDataSinks, 1)

	sink := config.Spec.AuditDataSinks[0]
	assert.Equal(t, models.V1DataSinkableSpecTypeCloudwatch, sink.Type)
	assert.Equal(t, "dev-hubble-audits", sink.CloudWatch.Group)
	assert.Equal(t, "us-east-1", sink.CloudWatch.Region)
	assert.Equal(t, "", sink.CloudWatch.Stream)
	require.NotNil(t, sink.CloudWatch.Credentials)
	assert.Equal(t, "AKIATEST", sink.CloudWatch.Credentials.AccessKey)
	assert.Equal(t, models.V1AwsCloudAccountCredentialTypeSecret, *sink.CloudWatch.Credentials.CredentialType)
}

func TestToCloudWatchValidateConfig(t *testing.T) {
	t.Parallel()

	d := schema.TestResourceDataRaw(t, resourceAuditTrail().Schema, map[string]interface{}{
		"name": "rag",
		"type": auditTrailTypeCloudWatch,
		"cloudwatch": []interface{}{
			map[string]interface{}{
				"group":           "dev-hubble-audits",
				"region":          "us-east-1",
				"credential_type": "secret",
				"access_key":      "AKIATEST",
				"secret_key":      "secret",
				"partition":       "aws",
			},
		},
	})

	config, err := toCloudWatchValidateConfig(d)
	require.NoError(t, err)
	assert.Equal(t, "dev-hubble-audits", config.Group)
	assert.Equal(t, "us-east-1", config.Region)
	assert.Equal(t, "", config.Stream)
	require.NotNil(t, config.Credentials)
	assert.Equal(t, "AKIATEST", config.Credentials.AccessKey)
	assert.Equal(t, "secret", config.Credentials.SecretKey)
}

func TestToCloudWatchDataSinkConfigSts(t *testing.T) {
	t.Parallel()

	d := schema.TestResourceDataRaw(t, resourceAuditTrail().Schema, map[string]interface{}{
		"name": "test-cw-sts",
		"type": auditTrailTypeCloudWatch,
		"cloudwatch": []interface{}{
			map[string]interface{}{
				"group":           "logs",
				"region":          "us-west-2",
				"credential_type": "sts",
				"arn":             "arn:aws:iam::123456789012:role/SpectroCloudRole",
				"external_id":     "external-id",
				"partition":       "aws",
			},
		},
	})

	config, err := toCloudWatchDataSinkConfig(d, "")
	require.NoError(t, err)
	credentials := config.Spec.AuditDataSinks[0].CloudWatch.Credentials
	require.NotNil(t, credentials.Sts)
	assert.Equal(t, "arn:aws:iam::123456789012:role/SpectroCloudRole", credentials.Sts.Arn)
	assert.Equal(t, "external-id", credentials.Sts.ExternalID)
}

func TestToSplunkSinkEntity(t *testing.T) {
	t.Parallel()

	d := schema.TestResourceDataRaw(t, resourceAuditTrail().Schema, map[string]interface{}{
		"name": "test-splunk",
		"type": auditTrailTypeSplunk,
		"splunk": []interface{}{
			map[string]interface{}{
				"hec_url": "https://http-inputs-example.splunkcloud.com:443",
				"token":   "hec-token",
				"index":   "main",
				"source":  "palette",
				"tls_config": []interface{}{
					map[string]interface{}{
						"ca_cert_base64":   "Y2VydA==",
						"tls_verification": true,
					},
				},
			},
		},
	})

	entity, err := toSplunkSinkEntity(d, false)
	require.NoError(t, err)
	require.NotNil(t, entity.Name)
	assert.Equal(t, "test-splunk", *entity.Name)
	require.NotNil(t, entity.Spec.HecURL)
	assert.Equal(t, "https://http-inputs-example.splunkcloud.com:443", *entity.Spec.HecURL)
	assert.Equal(t, strfmt.Password("hec-token"), *entity.Spec.Token)
	require.NotNil(t, entity.Spec.TLSConfig)
	assert.True(t, entity.Spec.TLSConfig.Enabled)
}

func TestToSplunkSinkEntityPreserveToken(t *testing.T) {
	t.Parallel()

	d := schema.TestResourceDataRaw(t, resourceAuditTrail().Schema, map[string]interface{}{
		"name": "test-splunk",
		"type": auditTrailTypeSplunk,
		"splunk": []interface{}{
			map[string]interface{}{
				"hec_url": "https://http-inputs-example.splunkcloud.com:443",
				"token":   "hec-token",
			},
		},
	})

	entity, err := toSplunkSinkEntity(d, true)
	require.NoError(t, err)
	assert.Equal(t, strfmt.Password(splunkTokenPreserve), *entity.Spec.Token)
}

func TestFlattenCloudWatchAuditTrail(t *testing.T) {
	t.Parallel()

	d := schema.TestResourceDataRaw(t, resourceAuditTrail().Schema, map[string]interface{}{
		"name": "test-cw",
		"type": auditTrailTypeCloudWatch,
	})

	config := &models.V1DataSinkConfig{
		Metadata: &models.V1ObjectMeta{Name: "test-cw"},
		Spec: &models.V1DataSinkSpec{
			AuditDataSinks: []*models.V1DataSinkableSpec{
				{
					Type: models.V1DataSinkableSpecTypeCloudwatch,
					CloudWatch: &models.V1CloudWatch{
						Group:  "logs",
						Region: "us-east-1",
						Stream: "audit",
						Credentials: &models.V1AwsCloudAccount{
							CredentialType: models.V1AwsCloudAccountCredentialTypeSecret.Pointer(),
							AccessKey:      "AKIATEST",
							Partition:      types.Ptr("aws"),
						},
					},
				},
			},
		},
	}

	require.NoError(t, flattenCloudWatchAuditTrail(d, config))
	cwList := d.Get("cloudwatch").([]interface{})
	require.Len(t, cwList, 1)
	cw := cwList[0].(map[string]interface{})
	assert.Equal(t, "logs", cw["group"])
	assert.Equal(t, "secret", cw["credential_type"])
	assert.Equal(t, "AKIATEST", cw["access_key"])
}

func TestFlattenCloudWatchAuditTrailPreservesSecretKeyFromState(t *testing.T) {
	t.Parallel()

	d := schema.TestResourceDataRaw(t, resourceAuditTrail().Schema, map[string]interface{}{
		"name": "test-cw",
		"type": auditTrailTypeCloudWatch,
		"cloudwatch": []interface{}{
			map[string]interface{}{
				"group":           "logs",
				"region":          "us-east-1",
				"credential_type": "secret",
				"access_key":      "AKIATEST",
				"secret_key":      "configured-secret",
				"partition":       "aws",
			},
		},
	})

	config := &models.V1DataSinkConfig{
		Metadata: &models.V1ObjectMeta{Name: "test-cw"},
		Spec: &models.V1DataSinkSpec{
			AuditDataSinks: []*models.V1DataSinkableSpec{
				{
					Type: models.V1DataSinkableSpecTypeCloudwatch,
					CloudWatch: &models.V1CloudWatch{
						Group:  "logs",
						Region: "us-east-1",
						Credentials: &models.V1AwsCloudAccount{
							CredentialType: models.V1AwsCloudAccountCredentialTypeSecret.Pointer(),
							AccessKey:      "AKIATEST",
						},
					},
				},
			},
		},
	}

	require.NoError(t, flattenCloudWatchAuditTrail(d, config))
	cw := d.Get("cloudwatch").([]interface{})[0].(map[string]interface{})
	assert.Equal(t, "configured-secret", cw["secret_key"])
}

func TestFlattenCloudWatchAuditTrailUsesAPISecretWhenStateEmpty(t *testing.T) {
	t.Parallel()

	d := schema.TestResourceDataRaw(t, resourceAuditTrail().Schema, map[string]interface{}{
		"name": "test-cw",
		"type": auditTrailTypeCloudWatch,
	})

	config := &models.V1DataSinkConfig{
		Metadata: &models.V1ObjectMeta{Name: "test-cw"},
		Spec: &models.V1DataSinkSpec{
			AuditDataSinks: []*models.V1DataSinkableSpec{
				{
					Type: models.V1DataSinkableSpecTypeCloudwatch,
					CloudWatch: &models.V1CloudWatch{
						Group:  "logs",
						Region: "us-east-1",
						Credentials: &models.V1AwsCloudAccount{
							CredentialType: models.V1AwsCloudAccountCredentialTypeSecret.Pointer(),
							AccessKey:      "AKIATEST",
							SecretKey:      "api-secret",
						},
					},
				},
			},
		},
	}

	require.NoError(t, flattenCloudWatchAuditTrail(d, config))
	cw := d.Get("cloudwatch").([]interface{})[0].(map[string]interface{})
	assert.Equal(t, "api-secret", cw["secret_key"])
}

func TestFlattenSplunkAuditTrail(t *testing.T) {
	t.Parallel()

	d := schema.TestResourceDataRaw(t, resourceAuditTrail().Schema, map[string]interface{}{
		"name": "test-splunk",
		"type": auditTrailTypeSplunk,
	})

	hecURL := "https://http-inputs-example.splunkcloud.com:443"
	sink := &models.V1SplunkSink{
		Metadata: &models.V1ObjectMeta{Name: "test-splunk"},
		Spec: &models.V1SplunkSinkSpec{
			HecURL: types.Ptr(hecURL),
			Index:  "main",
			Source: "palette",
			TLSConfig: &models.V1TLSCA{
				CaCertBase64:       "Y2VydA==",
				InsecureSkipVerify: false,
			},
		},
	}

	require.NoError(t, flattenSplunkAuditTrail(d, sink))
	spList := d.Get("splunk").([]interface{})
	require.Len(t, spList, 1)
	sp := spList[0].(map[string]interface{})
	assert.Equal(t, hecURL, sp["hec_url"])
	assert.Equal(t, "main", sp["index"])
}

func TestFlattenSplunkAuditTrailPreservesTokenFromState(t *testing.T) {
	t.Parallel()

	d := schema.TestResourceDataRaw(t, resourceAuditTrail().Schema, map[string]interface{}{
		"name": "test-splunk",
		"type": auditTrailTypeSplunk,
		"splunk": []interface{}{
			map[string]interface{}{
				"hec_url": "https://http-inputs-example.splunkcloud.com:443",
				"token":   "configured-token",
			},
		},
	})

	hecURL := "https://http-inputs-example.splunkcloud.com:443"
	sink := &models.V1SplunkSink{
		Metadata: &models.V1ObjectMeta{Name: "test-splunk"},
		Spec: &models.V1SplunkSinkSpec{
			HecURL: types.Ptr(hecURL),
			Index:  "main",
			Source: "palette",
		},
	}

	require.NoError(t, flattenSplunkAuditTrail(d, sink))
	sp := d.Get("splunk").([]interface{})[0].(map[string]interface{})
	assert.Equal(t, "configured-token", sp["token"])
}
