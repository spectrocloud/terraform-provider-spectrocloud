package spectrocloud

import (
	"context"
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
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
