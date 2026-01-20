package spectrocloud

import (
	"context"
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/assert"
)

func prepareOciEcrRegistryTestDataSTS() *schema.ResourceData {
	d := resourceRegistryOciEcr().TestResourceData()
	_ = d.Set("name", "testSTSRegistry")
	_ = d.Set("type", "ecr")
	_ = d.Set("endpoint", "123456.dkr.ecr.us-west-1.amazonaws.com")
	_ = d.Set("is_private", true)
	var credential []map[string]interface{}
	cred := map[string]interface{}{
		"credential_type": "sts",
		"arn":             "arn:aws:iam::123456:role/stage-demo-ecr",
		"external_id":     "sasdofiwhgowbsrgiornM=",
	}
	credential = append(credential, cred)
	_ = d.Set("credentials", credential)
	return d
}

func prepareOciEcrRegistryTestDataSecret() *schema.ResourceData {
	d := resourceRegistryOciEcr().TestResourceData()
	_ = d.Set("name", "testSecretRegistry")
	_ = d.Set("type", "ecr")
	_ = d.Set("endpoint", "123456.dkr.ecr.us-west-1.amazonaws.com")
	_ = d.Set("is_private", true)
	var credential []map[string]interface{}
	cred := map[string]interface{}{
		"credential_type": "secret",
		"secret_key":      "fasdfSADFsfasWQER23SADf23@",
		"access_key":      "ASFFSDFWEQDFVXRTGWDFV",
	}
	credential = append(credential, cred)
	d.Set("credentials", credential)
	return d
}

// Will enable back with adding support to validation
//func TestResourceRegistryEcrCreateSTS(t *testing.T) {
//	d := prepareOciEcrRegistryTestDataSTS()
//	ctx := context.Background()
//	diags := resourceRegistryEcrCreate(ctx, d, unitTestMockAPIClient)
//	assert.Empty(t, diags)
//}
//
//func TestResourceRegistryEcrCreateSecret(t *testing.T) {
//	d := prepareOciEcrRegistryTestDataSecret()
//	ctx := context.Background()
//	diags := resourceRegistryEcrCreate(ctx, d, unitTestMockAPIClient)
//	assert.Empty(t, diags)
//}

func TestResourceRegistryEcrRead(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSTS()
	ctx := context.Background()
	d.SetId("test-id")
	diags := resourceRegistryEcrRead(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}

//func TestResourceRegistryEcrUpdate(t *testing.T) {
//	d := prepareOciEcrRegistryTestDataSTS()
//	ctx := context.Background()
//	d.SetId("test-id")
//	diags := resourceRegistryEcrUpdate(ctx, d, unitTestMockAPIClient)
//	assert.Empty(t, diags)
//}

func TestResourceRegistryEcrDelete(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSTS()
	ctx := context.Background()
	d.SetId("test-id")
	diags := resourceRegistryEcrDelete(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}

func TestValidateRegistryCred(t *testing.T) {
	tests := []struct {
		name         string
		client       interface{}
		registryType string
		providerType string
		isSync       bool
		basicSpec    *models.V1BasicOciRegistrySpec
		ecrSpec      *models.V1EcrRegistrySpec
		expectError  bool
		description  string
	}{
		{
			name:         "Skip validation when isSync is false",
			client:       unitTestMockAPIClient,
			registryType: "basic",
			providerType: "helm",
			isSync:       false,
			basicSpec: &models.V1BasicOciRegistrySpec{
				Endpoint:     StringPtr("https://registry.example.com"),
				ProviderType: StringPtr("helm"),
			},
			ecrSpec:     nil,
			expectError: false,
			description: "Should skip validation when isSync is false, regardless of other parameters",
		},
		{
			name:         "Successfully validate basic registry with noAuth type",
			client:       unitTestMockAPIClient,
			registryType: "basic",
			providerType: "helm",
			isSync:       true,
			basicSpec: &models.V1BasicOciRegistrySpec{
				Endpoint:     StringPtr("https://public-registry.example.com"),
				ProviderType: StringPtr("helm"),
				Auth: &models.V1RegistryAuth{
					Type: "noAuth",
				},
			},
			ecrSpec:     nil,
			expectError: false,
			description: "Should successfully validate basic registry with noAuth authentication type",
		},
		{
			name:         "Successfully validate basic registry with zarf provider",
			client:       unitTestMockAPIClient,
			registryType: "basic",
			providerType: "zarf",
			isSync:       true,
			basicSpec: &models.V1BasicOciRegistrySpec{
				Endpoint:     StringPtr("https://registry.example.com"),
				ProviderType: StringPtr("zarf"),
				Auth: &models.V1RegistryAuth{
					Type:     "basic",
					Username: "test-user",
					Password: strfmt.Password("test-pass"),
				},
			},
			ecrSpec:     nil,
			expectError: false,
			description: "Should successfully validate basic registry with zarf provider when all conditions are met",
		},
		{
			name:         "Successfully validate basic registry with pack provider",
			client:       unitTestMockAPIClient,
			registryType: "basic",
			providerType: "pack",
			isSync:       true,
			basicSpec: &models.V1BasicOciRegistrySpec{
				Endpoint:     StringPtr("https://registry.example.com"),
				ProviderType: StringPtr("pack"),
				Auth: &models.V1RegistryAuth{
					Type:     "basic",
					Username: "test-user",
					Password: strfmt.Password("test-pass"),
				},
			},
			ecrSpec:     nil,
			expectError: false,
			description: "Should successfully validate basic registry with pack provider when all conditions are met",
		},
		{
			name:         "Successfully validate basic registry",
			client:       unitTestMockAPIClient,
			registryType: "basic",
			providerType: "helm",
			isSync:       true,
			basicSpec: &models.V1BasicOciRegistrySpec{
				Endpoint:     StringPtr("https://registry.example.com"),
				ProviderType: StringPtr("helm"),
				Auth: &models.V1RegistryAuth{
					Type:     "basic",
					Username: "test-user",
					Password: strfmt.Password("test-pass"),
				},
			},
			ecrSpec:     nil,
			expectError: false,
			description: "Should successfully validate basic registry when all conditions are met",
		},
		{
			name:         "Validate ECR registry - SDK error due to mock response",
			client:       unitTestMockAPIClient,
			registryType: "ecr",
			providerType: "helm",
			isSync:       true,
			basicSpec:    nil,
			ecrSpec: &models.V1EcrRegistrySpec{
				Endpoint:     StringPtr("123456.dkr.ecr.us-west-1.amazonaws.com"),
				IsPrivate:    BoolPtr(true),
				ProviderType: StringPtr("helm"),
				Credentials: &models.V1AwsCloudAccount{
					CredentialType: models.V1AwsCloudAccountCredentialTypeSecret.Pointer(),
					AccessKey:      "test-access-key",
					SecretKey:      "test-secret-key",
				},
			},
			expectError: true,
			description: "Should attempt ECR validation but get SDK error due to mock API returning 200 instead of 204",
		},
		{
			name:         "Error when basic registry validation fails",
			client:       unitTestMockAPINegativeClient,
			registryType: "basic",
			providerType: "helm",
			isSync:       true,
			basicSpec: &models.V1BasicOciRegistrySpec{
				Endpoint:     StringPtr("https://invalid-registry.example.com"),
				ProviderType: StringPtr("helm"),
				Auth: &models.V1RegistryAuth{
					Type:     "basic",
					Username: "invalid-user",
					Password: strfmt.Password("invalid-pass"),
				},
			},
			ecrSpec:     nil,
			expectError: true,
			description: "Should return error when basic registry validation fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert interface{} client to *client.V1Client
			c := getV1ClientWithResourceContext(tt.client, "tenant")

			err := validateRegistryCred(c, tt.registryType, tt.providerType, tt.isSync, tt.basicSpec, tt.ecrSpec)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

func TestToRegistryEcr(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *schema.ResourceData
		description string
		verify      func(t *testing.T, registry *models.V1EcrRegistry)
	}{
		{
			name: "Successfully convert with STS credentials",
			setup: func() *schema.ResourceData {
				d := resourceRegistryOciEcr().TestResourceData()
				d.Set("name", "test-ecr-registry")
				d.Set("endpoint", "123456.dkr.ecr.us-west-1.amazonaws.com")
				d.Set("is_private", true)
				d.Set("is_synchronization", true)
				d.Set("provider_type", "helm")
				d.Set("base_content_path", "/test/path")
				cred := []map[string]interface{}{
					{
						"credential_type": "sts",
						"arn":             "arn:aws:iam::123456:role/test-role",
						"external_id":     "test-external-id",
						"tls_config":      []interface{}{},
					},
				}
				d.Set("credentials", cred)
				return d
			},
			description: "Should successfully convert ResourceData to V1EcrRegistry with STS credentials",
			verify: func(t *testing.T, registry *models.V1EcrRegistry) {
				assert.NotNil(t, registry)
				assert.NotNil(t, registry.Metadata)
				assert.Equal(t, "test-ecr-registry", registry.Metadata.Name)
				assert.NotNil(t, registry.Spec)
				assert.NotNil(t, registry.Spec.Endpoint)
				assert.Equal(t, "123456.dkr.ecr.us-west-1.amazonaws.com", *registry.Spec.Endpoint)
				assert.NotNil(t, registry.Spec.IsPrivate)
				assert.True(t, *registry.Spec.IsPrivate)
				assert.True(t, registry.Spec.IsSyncSupported)
				assert.NotNil(t, registry.Spec.ProviderType)
				assert.Equal(t, "helm", *registry.Spec.ProviderType)
				assert.Equal(t, "/test/path", registry.Spec.BaseContentPath)
				assert.NotNil(t, registry.Spec.Credentials)
				assert.Equal(t, models.V1AwsCloudAccountCredentialTypeSts, *registry.Spec.Credentials.CredentialType)
				assert.NotNil(t, registry.Spec.Credentials.Sts)
				assert.Equal(t, "arn:aws:iam::123456:role/test-role", registry.Spec.Credentials.Sts.Arn)
				assert.Equal(t, "test-external-id", registry.Spec.Credentials.Sts.ExternalID)
				assert.NotNil(t, registry.Spec.TLS)
				assert.True(t, registry.Spec.TLS.Enabled)
				assert.False(t, registry.Spec.TLS.InsecureSkipVerify)
				assert.Empty(t, registry.Spec.TLS.Certificate)
			},
		},
		{
			name: "Successfully convert with secret credentials",
			setup: func() *schema.ResourceData {
				d := resourceRegistryOciEcr().TestResourceData()
				d.Set("name", "test-secret-registry")
				d.Set("endpoint", "789012.dkr.ecr.us-east-1.amazonaws.com")
				d.Set("is_private", false)
				d.Set("is_synchronization", false)
				d.Set("provider_type", "pack")
				d.Set("base_content_path", "")
				cred := []map[string]interface{}{
					{
						"credential_type": "secret",
						"access_key":      "AKIAIOSFODNN7EXAMPLE",
						"secret_key":      "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
						"tls_config":      []interface{}{},
					},
				}
				d.Set("credentials", cred)
				return d
			},
			description: "Should successfully convert ResourceData to V1EcrRegistry with secret credentials",
			verify: func(t *testing.T, registry *models.V1EcrRegistry) {
				assert.NotNil(t, registry)
				assert.Equal(t, "test-secret-registry", registry.Metadata.Name)
				assert.Equal(t, "789012.dkr.ecr.us-east-1.amazonaws.com", *registry.Spec.Endpoint)
				assert.False(t, *registry.Spec.IsPrivate)
				assert.False(t, registry.Spec.IsSyncSupported)
				assert.Equal(t, "pack", *registry.Spec.ProviderType)
				assert.Empty(t, registry.Spec.BaseContentPath)
				assert.Equal(t, models.V1AwsCloudAccountCredentialTypeSecret, *registry.Spec.Credentials.CredentialType)
				assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", registry.Spec.Credentials.AccessKey)
				assert.Equal(t, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", registry.Spec.Credentials.SecretKey)
			},
		},
		{
			name: "Successfully convert with TLS configuration",
			setup: func() *schema.ResourceData {
				d := resourceRegistryOciEcr().TestResourceData()
				d.Set("name", "test-tls-registry")
				d.Set("endpoint", "345678.dkr.ecr.eu-west-1.amazonaws.com")
				d.Set("is_private", true)
				d.Set("is_synchronization", true)
				d.Set("provider_type", "helm")
				d.Set("base_content_path", "/custom/path")
				cred := []map[string]interface{}{
					{
						"credential_type": "secret",
						"access_key":      "test-access-key",
						"secret_key":      "test-secret-key",
						"tls_config": []interface{}{
							map[string]interface{}{
								"certificate":          "-----BEGIN CERTIFICATE-----\nTEST_CERT\n-----END CERTIFICATE-----",
								"insecure_skip_verify": true,
							},
						},
					},
				}
				d.Set("credentials", cred)
				return d
			},
			description: "Should successfully convert ResourceData to V1EcrRegistry with TLS configuration",
			verify: func(t *testing.T, registry *models.V1EcrRegistry) {
				assert.NotNil(t, registry)
				assert.NotNil(t, registry.Spec.TLS)
				assert.True(t, registry.Spec.TLS.Enabled)
				assert.True(t, registry.Spec.TLS.InsecureSkipVerify)
				assert.Equal(t, "-----BEGIN CERTIFICATE-----\nTEST_CERT\n-----END CERTIFICATE-----", registry.Spec.TLS.Certificate)
			},
		},
		{
			name: "Successfully convert without TLS configuration",
			setup: func() *schema.ResourceData {
				d := resourceRegistryOciEcr().TestResourceData()
				d.Set("name", "test-no-tls-registry")
				d.Set("endpoint", "456789.dkr.ecr.ap-southeast-1.amazonaws.com")
				d.Set("is_private", true)
				d.Set("is_synchronization", false)
				d.Set("provider_type", "zarf")
				d.Set("base_content_path", "")
				cred := []map[string]interface{}{
					{
						"credential_type": "secret",
						"access_key":      "test-key",
						"secret_key":      "test-secret",
						"tls_config":      []interface{}{},
					},
				}
				d.Set("credentials", cred)
				return d
			},
			description: "Should successfully convert ResourceData to V1EcrRegistry without TLS configuration",
			verify: func(t *testing.T, registry *models.V1EcrRegistry) {
				assert.NotNil(t, registry)
				assert.NotNil(t, registry.Spec.TLS)
				assert.True(t, registry.Spec.TLS.Enabled)
				assert.False(t, registry.Spec.TLS.InsecureSkipVerify)
				assert.Empty(t, registry.Spec.TLS.Certificate)
			},
		},
		{
			name: "Successfully convert with empty credential_type defaults to secret",
			setup: func() *schema.ResourceData {
				d := resourceRegistryOciEcr().TestResourceData()
				d.Set("name", "test-default-cred-registry")
				d.Set("endpoint", "567890.dkr.ecr.ca-central-1.amazonaws.com")
				d.Set("is_private", true)
				d.Set("is_synchronization", true)
				d.Set("provider_type", "helm")
				d.Set("base_content_path", "/default/path")
				cred := []map[string]interface{}{
					{
						"credential_type": "",
						"access_key":      "default-access-key",
						"secret_key":      "default-secret-key",
						"tls_config":      []interface{}{},
					},
				}
				d.Set("credentials", cred)
				return d
			},
			description: "Should successfully convert with empty credential_type defaulting to secret",
			verify: func(t *testing.T, registry *models.V1EcrRegistry) {
				assert.NotNil(t, registry)
				assert.Equal(t, models.V1AwsCloudAccountCredentialTypeSecret, *registry.Spec.Credentials.CredentialType)
				assert.Equal(t, "default-access-key", registry.Spec.Credentials.AccessKey)
				assert.Equal(t, "default-secret-key", registry.Spec.Credentials.SecretKey)
			},
		},
		{
			name: "Successfully convert with all fields populated",
			setup: func() *schema.ResourceData {
				d := resourceRegistryOciEcr().TestResourceData()
				d.Set("name", "test-complete-registry")
				d.Set("endpoint", "678901.dkr.ecr.sa-east-1.amazonaws.com")
				d.Set("is_private", true)
				d.Set("is_synchronization", true)
				d.Set("provider_type", "pack")
				d.Set("base_content_path", "/complete/path")
				cred := []map[string]interface{}{
					{
						"credential_type": "sts",
						"arn":             "arn:aws:iam::678901:role/complete-role",
						"external_id":     "complete-external-id",
						"tls_config": []interface{}{
							map[string]interface{}{
								"certificate":          "complete-cert",
								"insecure_skip_verify": false,
							},
						},
					},
				}
				d.Set("credentials", cred)
				return d
			},
			description: "Should successfully convert with all fields populated including TLS",
			verify: func(t *testing.T, registry *models.V1EcrRegistry) {
				assert.NotNil(t, registry)
				assert.Equal(t, "test-complete-registry", registry.Metadata.Name)
				assert.Equal(t, "678901.dkr.ecr.sa-east-1.amazonaws.com", *registry.Spec.Endpoint)
				assert.True(t, *registry.Spec.IsPrivate)
				assert.True(t, registry.Spec.IsSyncSupported)
				assert.Equal(t, "pack", *registry.Spec.ProviderType)
				assert.Equal(t, "/complete/path", registry.Spec.BaseContentPath)
				assert.Equal(t, models.V1AwsCloudAccountCredentialTypeSts, *registry.Spec.Credentials.CredentialType)
				assert.Equal(t, "arn:aws:iam::678901:role/complete-role", registry.Spec.Credentials.Sts.Arn)
				assert.Equal(t, "complete-external-id", registry.Spec.Credentials.Sts.ExternalID)
				assert.Equal(t, "complete-cert", registry.Spec.TLS.Certificate)
				assert.False(t, registry.Spec.TLS.InsecureSkipVerify)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.setup()
			registry := toRegistryEcr(d)

			if tt.verify != nil {
				tt.verify(t, registry)
			}
		})
	}
}

func TestToRegistryBasic(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *schema.ResourceData
		description string
		expectPanic bool
		verify      func(t *testing.T, registry *models.V1BasicOciRegistry)
	}{
		{
			name: "Successfully convert with basic authentication",
			setup: func() *schema.ResourceData {
				d := resourceRegistryOciEcr().TestResourceData()
				d.Set("name", "test-basic-registry")
				d.Set("endpoint", "https://registry.example.com")
				d.Set("provider_type", "helm")
				d.Set("is_synchronization", true)
				d.Set("endpoint_suffix", "/v2")
				d.Set("base_content_path", "/test/path")
				cred := []map[string]interface{}{
					{
						"credential_type": "basic",
						"username":        "test-user",
						"password":        "test-password",
						"tls_config":      []interface{}{},
					},
				}
				d.Set("credentials", cred)
				return d
			},
			description: "Should successfully convert ResourceData to V1BasicOciRegistry with basic authentication",
			verify: func(t *testing.T, registry *models.V1BasicOciRegistry) {
				assert.NotNil(t, registry)
				assert.NotNil(t, registry.Metadata)
				assert.Equal(t, "test-basic-registry", registry.Metadata.Name)
				assert.NotNil(t, registry.Spec)
				assert.NotNil(t, registry.Spec.Endpoint)
				assert.Equal(t, "https://registry.example.com", *registry.Spec.Endpoint)
				assert.Equal(t, "/v2", registry.Spec.BasePath)
				assert.NotNil(t, registry.Spec.ProviderType)
				assert.Equal(t, "helm", *registry.Spec.ProviderType)
				assert.Equal(t, "/test/path", registry.Spec.BaseContentPath)
				assert.True(t, registry.Spec.IsSyncSupported)
				assert.NotNil(t, registry.Spec.Auth)
				assert.Equal(t, "basic", registry.Spec.Auth.Type)
				assert.Equal(t, "test-user", registry.Spec.Auth.Username)
				assert.Equal(t, "test-password", registry.Spec.Auth.Password.String())
				assert.NotNil(t, registry.Spec.Auth.TLS)
				assert.True(t, registry.Spec.Auth.TLS.Enabled)
				assert.False(t, registry.Spec.Auth.TLS.InsecureSkipVerify)
				assert.Empty(t, registry.Spec.Auth.TLS.Certificate)
			},
		},
		{
			name: "Successfully convert with noAuth authentication",
			setup: func() *schema.ResourceData {
				d := resourceRegistryOciEcr().TestResourceData()
				d.Set("name", "test-noauth-registry")
				d.Set("endpoint", "https://public-registry.example.com")
				d.Set("provider_type", "zarf")
				d.Set("is_synchronization", false)
				d.Set("endpoint_suffix", "")
				d.Set("base_content_path", "")
				cred := []map[string]interface{}{
					{
						"credential_type": "noAuth",
						"tls_config":      []interface{}{},
					},
				}
				d.Set("credentials", cred)
				return d
			},
			description: "Should successfully convert ResourceData to V1BasicOciRegistry with noAuth authentication",
			verify: func(t *testing.T, registry *models.V1BasicOciRegistry) {
				assert.NotNil(t, registry)
				assert.Equal(t, "test-noauth-registry", registry.Metadata.Name)
				assert.Equal(t, "https://public-registry.example.com", *registry.Spec.Endpoint)
				assert.Empty(t, registry.Spec.BasePath)
				assert.Equal(t, "zarf", *registry.Spec.ProviderType)
				assert.Empty(t, registry.Spec.BaseContentPath)
				assert.False(t, registry.Spec.IsSyncSupported)
				assert.Equal(t, "noAuth", registry.Spec.Auth.Type)
				assert.Empty(t, registry.Spec.Auth.Username)
				assert.Empty(t, registry.Spec.Auth.Password.String())
			},
		},
		{
			name: "Successfully convert with TLS configuration",
			setup: func() *schema.ResourceData {
				d := resourceRegistryOciEcr().TestResourceData()
				d.Set("name", "test-tls-registry")
				d.Set("endpoint", "https://secure-registry.example.com")
				d.Set("provider_type", "pack")
				d.Set("is_synchronization", true)
				d.Set("endpoint_suffix", "/secure")
				d.Set("base_content_path", "/secure/path")
				cred := []map[string]interface{}{
					{
						"credential_type": "basic",
						"username":        "secure-user",
						"password":        "secure-password",
						"tls_config": []interface{}{
							map[string]interface{}{
								"certificate":          "-----BEGIN CERTIFICATE-----\nSECURE_CERT\n-----END CERTIFICATE-----",
								"insecure_skip_verify": true,
							},
						},
					},
				}
				d.Set("credentials", cred)
				return d
			},
			description: "Should successfully convert ResourceData to V1BasicOciRegistry with TLS configuration",
			verify: func(t *testing.T, registry *models.V1BasicOciRegistry) {
				assert.NotNil(t, registry)
				assert.NotNil(t, registry.Spec.Auth.TLS)
				assert.True(t, registry.Spec.Auth.TLS.Enabled)
				assert.True(t, registry.Spec.Auth.TLS.InsecureSkipVerify)
				assert.Equal(t, "-----BEGIN CERTIFICATE-----\nSECURE_CERT\n-----END CERTIFICATE-----", registry.Spec.Auth.TLS.Certificate)
			},
		},
		{
			name: "Successfully convert without TLS configuration",
			setup: func() *schema.ResourceData {
				d := resourceRegistryOciEcr().TestResourceData()
				d.Set("name", "test-no-tls-registry")
				d.Set("endpoint", "https://registry.example.com")
				d.Set("provider_type", "helm")
				d.Set("is_synchronization", false)
				d.Set("endpoint_suffix", "")
				d.Set("base_content_path", "")
				cred := []map[string]interface{}{
					{
						"credential_type": "basic",
						"username":        "user",
						"password":        "pass",
						"tls_config":      []interface{}{},
					},
				}
				d.Set("credentials", cred)
				return d
			},
			description: "Should successfully convert ResourceData to V1BasicOciRegistry without TLS configuration",
			verify: func(t *testing.T, registry *models.V1BasicOciRegistry) {
				assert.NotNil(t, registry)
				assert.NotNil(t, registry.Spec.Auth.TLS)
				assert.True(t, registry.Spec.Auth.TLS.Enabled)
				assert.False(t, registry.Spec.Auth.TLS.InsecureSkipVerify)
				assert.Empty(t, registry.Spec.Auth.TLS.Certificate)
			},
		},
		{
			name: "Successfully convert with empty credential_type defaults to noAuth",
			setup: func() *schema.ResourceData {
				d := resourceRegistryOciEcr().TestResourceData()
				d.Set("name", "test-default-auth-registry")
				d.Set("endpoint", "https://registry.example.com")
				d.Set("provider_type", "helm")
				d.Set("is_synchronization", true)
				d.Set("endpoint_suffix", "/default")
				d.Set("base_content_path", "/default/path")
				cred := []map[string]interface{}{
					{
						"credential_type": "",
						"tls_config":      []interface{}{},
					},
				}
				d.Set("credentials", cred)
				return d
			},
			description: "Should successfully convert with empty credential_type defaulting to noAuth",
			verify: func(t *testing.T, registry *models.V1BasicOciRegistry) {
				assert.NotNil(t, registry)
				assert.Equal(t, "noAuth", registry.Spec.Auth.Type)
				assert.Empty(t, registry.Spec.Auth.Username)
				assert.Empty(t, registry.Spec.Auth.Password.String())
			},
		},
		{
			name: "Successfully convert with all fields populated",
			setup: func() *schema.ResourceData {
				d := resourceRegistryOciEcr().TestResourceData()
				d.Set("name", "test-complete-registry")
				d.Set("endpoint", "https://complete-registry.example.com")
				d.Set("provider_type", "zarf")
				d.Set("is_synchronization", true)
				d.Set("endpoint_suffix", "/complete/v2")
				d.Set("base_content_path", "/complete/path")
				cred := []map[string]interface{}{
					{
						"credential_type": "basic",
						"username":        "complete-user",
						"password":        "complete-password",
						"tls_config": []interface{}{
							map[string]interface{}{
								"certificate":          "complete-cert",
								"insecure_skip_verify": false,
							},
						},
					},
				}
				d.Set("credentials", cred)
				return d
			},
			description: "Should successfully convert with all fields populated including TLS",
			verify: func(t *testing.T, registry *models.V1BasicOciRegistry) {
				assert.NotNil(t, registry)
				assert.Equal(t, "test-complete-registry", registry.Metadata.Name)
				assert.Equal(t, "https://complete-registry.example.com", *registry.Spec.Endpoint)
				assert.Equal(t, "/complete/v2", registry.Spec.BasePath)
				assert.Equal(t, "zarf", *registry.Spec.ProviderType)
				assert.Equal(t, "/complete/path", registry.Spec.BaseContentPath)
				assert.True(t, registry.Spec.IsSyncSupported)
				assert.Equal(t, "basic", registry.Spec.Auth.Type)
				assert.Equal(t, "complete-user", registry.Spec.Auth.Username)
				assert.Equal(t, "complete-password", registry.Spec.Auth.Password.String())
				assert.Equal(t, "complete-cert", registry.Spec.Auth.TLS.Certificate)
				assert.False(t, registry.Spec.Auth.TLS.InsecureSkipVerify)
			},
		},
		{
			name: "Successfully convert with basic auth but missing username/password",
			setup: func() *schema.ResourceData {
				d := resourceRegistryOciEcr().TestResourceData()
				d.Set("name", "test-basic-no-creds-registry")
				d.Set("endpoint", "https://registry.example.com")
				d.Set("provider_type", "helm")
				d.Set("is_synchronization", false)
				d.Set("endpoint_suffix", "")
				d.Set("base_content_path", "")
				cred := []map[string]interface{}{
					{
						"credential_type": "basic",
						"tls_config":      []interface{}{},
					},
				}
				d.Set("credentials", cred)
				return d
			},
			description: "Should successfully convert with basic auth type but empty username/password",
			verify: func(t *testing.T, registry *models.V1BasicOciRegistry) {
				assert.NotNil(t, registry)
				assert.Equal(t, "basic", registry.Spec.Auth.Type)
				assert.Empty(t, registry.Spec.Auth.Username)
				assert.Empty(t, registry.Spec.Auth.Password.String())
			},
		},
		{
			name: "Panic when credentials is nil",
			setup: func() *schema.ResourceData {
				d := resourceRegistryOciEcr().TestResourceData()
				d.Set("name", "test-registry")
				d.Set("endpoint", "https://registry.example.com")
				d.Set("provider_type", "helm")
				d.Set("is_synchronization", false)
				d.Set("endpoint_suffix", "")
				d.Set("base_content_path", "")
				// Credentials not set - would cause panic on type assertion
				return d
			},
			expectPanic: true,
			description: "Should panic when credentials is nil due to type assertion failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.setup()

			if tt.expectPanic {
				assert.Panics(t, func() {
					toRegistryBasic(d)
				}, tt.description)
			} else {
				registry := toRegistryBasic(d)
				if tt.verify != nil {
					tt.verify(t, registry)
				}
			}
		})
	}
}

func TestToRegistryAwsAccountCredential(t *testing.T) {
	tests := []struct {
		name        string
		regCred     map[string]interface{}
		description string
		verify      func(t *testing.T, account *models.V1AwsCloudAccount)
	}{
		{
			name: "Successfully convert with secret credential type",
			regCred: map[string]interface{}{
				"credential_type": "secret",
				"access_key":      "AKIAIOSFODNN7EXAMPLE",
				"secret_key":      "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			},
			description: "Should successfully convert to V1AwsCloudAccount with secret credentials",
			verify: func(t *testing.T, account *models.V1AwsCloudAccount) {
				assert.NotNil(t, account)
				assert.NotNil(t, account.CredentialType)
				assert.Equal(t, models.V1AwsCloudAccountCredentialTypeSecret, *account.CredentialType)
				assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", account.AccessKey)
				assert.Equal(t, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", account.SecretKey)
				assert.Nil(t, account.Sts)
			},
		},
		{
			name: "Successfully convert with empty credential_type defaults to secret",
			regCred: map[string]interface{}{
				"credential_type": "",
				"access_key":      "DEFAULT_ACCESS_KEY",
				"secret_key":      "DEFAULT_SECRET_KEY",
			},
			description: "Should successfully convert with empty credential_type defaulting to secret",
			verify: func(t *testing.T, account *models.V1AwsCloudAccount) {
				assert.NotNil(t, account)
				assert.Equal(t, models.V1AwsCloudAccountCredentialTypeSecret, *account.CredentialType)
				assert.Equal(t, "DEFAULT_ACCESS_KEY", account.AccessKey)
				assert.Equal(t, "DEFAULT_SECRET_KEY", account.SecretKey)
				assert.Nil(t, account.Sts)
			},
		},
		{
			name: "Successfully convert with STS credential type",
			regCred: map[string]interface{}{
				"credential_type": "sts",
				"arn":             "arn:aws:iam::123456789012:role/test-role",
				"external_id":     "test-external-id-12345",
			},
			description: "Should successfully convert to V1AwsCloudAccount with STS credentials",
			verify: func(t *testing.T, account *models.V1AwsCloudAccount) {
				assert.NotNil(t, account)
				assert.NotNil(t, account.CredentialType)
				assert.Equal(t, models.V1AwsCloudAccountCredentialTypeSts, *account.CredentialType)
				assert.NotNil(t, account.Sts)
				assert.Equal(t, "arn:aws:iam::123456789012:role/test-role", account.Sts.Arn)
				assert.Equal(t, "test-external-id-12345", account.Sts.ExternalID)
				assert.Empty(t, account.AccessKey)
				assert.Empty(t, account.SecretKey)
			},
		},
		{
			name: "Successfully convert with STS credentials and empty external_id",
			regCred: map[string]interface{}{
				"credential_type": "sts",
				"arn":             "arn:aws:iam::987654321098:role/another-role",
				"external_id":     "",
			},
			description: "Should successfully convert with STS credentials even with empty external_id",
			verify: func(t *testing.T, account *models.V1AwsCloudAccount) {
				assert.NotNil(t, account)
				assert.Equal(t, models.V1AwsCloudAccountCredentialTypeSts, *account.CredentialType)
				assert.NotNil(t, account.Sts)
				assert.Equal(t, "arn:aws:iam::987654321098:role/another-role", account.Sts.Arn)
				assert.Empty(t, account.Sts.ExternalID)
			},
		},
		{
			name: "Successfully convert with secret credentials and empty keys",
			regCred: map[string]interface{}{
				"credential_type": "secret",
				"access_key":      "",
				"secret_key":      "",
			},
			description: "Should successfully convert with secret credentials even with empty keys",
			verify: func(t *testing.T, account *models.V1AwsCloudAccount) {
				assert.NotNil(t, account)
				assert.Equal(t, models.V1AwsCloudAccountCredentialTypeSecret, *account.CredentialType)
				assert.Empty(t, account.AccessKey)
				assert.Empty(t, account.SecretKey)
				assert.Nil(t, account.Sts)
			},
		},
		{
			name: "Successfully convert with long ARN and external ID",
			regCred: map[string]interface{}{
				"credential_type": "sts",
				"arn":             "arn:aws:iam::111222333444:role/very-long-role-name-with-many-segments",
				"external_id":     "very-long-external-id-with-special-chars-!@#$%^&*()",
			},
			description: "Should successfully convert with long STS credentials",
			verify: func(t *testing.T, account *models.V1AwsCloudAccount) {
				assert.NotNil(t, account)
				assert.Equal(t, models.V1AwsCloudAccountCredentialTypeSts, *account.CredentialType)
				assert.Equal(t, "arn:aws:iam::111222333444:role/very-long-role-name-with-many-segments", account.Sts.Arn)
				assert.Equal(t, "very-long-external-id-with-special-chars-!@#$%^&*()", account.Sts.ExternalID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := toRegistryAwsAccountCredential(tt.regCred)

			if tt.verify != nil {
				tt.verify(t, account)
			}
		})
	}
}

func TestResourceOciRegistrySyncRefreshFunc(t *testing.T) {
	uid := "test-registry-uid"

	tests := []struct {
		name        string
		setupClient func() *client.V1Client
		description string
		verify      func(t *testing.T, result interface{}, state string, err error)
	}{
		{
			name: "Success status returns Success state",
			setupClient: func() *client.V1Client {
				return getV1ClientWithResourceContext(unitTestMockAPIClient, "tenant")
			},
			description: "Should return Success state when status is Success",
			verify: func(t *testing.T, result interface{}, state string, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.Equal(t, "Success", state, "Should return Success state")
				assert.NotNil(t, result, "Result should not be nil")
			},
		},
		{
			name: "Completed status returns Success state",
			setupClient: func() *client.V1Client {
				return getV1ClientWithResourceContext(unitTestMockAPIClient, "tenant")
			},
			description: "Should return Success state when status is Completed (mock returns Success which maps to Completed)",
			verify: func(t *testing.T, result interface{}, state string, err error) {
				assert.NoError(t, err, "Should not have error")
				assert.Equal(t, "Success", state, "Should return Success state")
			},
		},
		{
			name: "API error returns error",
			setupClient: func() *client.V1Client {
				return getV1ClientWithResourceContext(unitTestMockAPINegativeClient, "tenant")
			},
			description: "Should return error when API call fails",
			verify: func(t *testing.T, result interface{}, state string, err error) {
				assert.Error(t, err, "Should have error when API fails")
				assert.Nil(t, result, "Result should be nil on error")
				assert.Empty(t, state, "State should be empty on error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setupClient()
			refreshFunc := resourceOciRegistrySyncRefreshFunc(c, uid)

			result, state, err := refreshFunc()

			if tt.verify != nil {
				tt.verify(t, result, state, err)
			}
		})
	}
}
