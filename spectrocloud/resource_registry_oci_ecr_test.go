package spectrocloud

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-api-go/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mock"
)

func prepareOciEcrRegistryTestDataSTS() *schema.ResourceData {
	d := resourceRegistryOciEcr().TestResourceData()
	d.Set("name", "testSTSRegistry")
	d.Set("type", "ecr")
	d.Set("endpoint", "123456.dkr.ecr.us-west-1.amazonaws.com")
	d.Set("is_private", true)
	var credential []map[string]interface{}
	cred := map[string]interface{}{
		"credential_type": "sts",
		"arn":             "arn:aws:iam::123456:role/stage-demo-ecr",
		"external_id":     "sasdofiwhgowbsrgiornM=",
	}
	credential = append(credential, cred)
	d.Set("credentials", credential)
	return d
}

func prepareOciEcrRegistryTestDataSecret() *schema.ResourceData {
	d := resourceRegistryOciEcr().TestResourceData()
	d.Set("name", "testSecretRegistry")
	d.Set("type", "ecr")
	d.Set("endpoint", "123456.dkr.ecr.us-west-1.amazonaws.com")
	d.Set("is_private", true)
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

func TestResourceRegistryEcrCreateSTS(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSTS()
	m := &client.V1Client{
		CreateOciEcrRegistryFn: func(registry *models.V1EcrRegistry) (string, error) {
			return "test-sts-oci-reg-ecr-uid", nil
		},
	}
	ctx := context.Background()
	diags := resourceRegistryEcrCreate(ctx, d, m)
	if len(diags) > 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
	if d.Id() != "test-sts-oci-reg-ecr-uid" {
		t.Errorf("Expected ID to be 'test-sts-oci-reg-ecr-uid', got %s", d.Id())
	}
}

func TestResourceRegistryEcrCreateSecret(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSecret()
	m := &client.V1Client{
		CreateOciEcrRegistryFn: func(registry *models.V1EcrRegistry) (string, error) {
			return "test-secret-oci-reg-ecr-uid", nil
		},
	}
	ctx := context.Background()
	diags := resourceRegistryEcrCreate(ctx, d, m)
	if len(diags) > 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
	if d.Id() != "test-secret-oci-reg-ecr-uid" {
		t.Errorf("Expected ID to be 'test-secret-oci-reg-ecr-uid', got %s", d.Id())
	}
}

func TestResourceRegistryEcrCreateErr(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSecret()
	m := &client.V1Client{
		CreateOciEcrRegistryFn: func(registry *models.V1EcrRegistry) (string, error) {
			return "", errors.New("covering error case")
		},
	}
	ctx := context.Background()
	diags := resourceRegistryEcrCreate(ctx, d, m)
	if diags[0].Summary != "covering error case" {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
}

func TestResourceRegistryEcrReadSecret(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSTS()
	d.SetId("test-reg-oci")
	endpoint := "214575254960.dkr.ecr.us-east-2.amazonaws.com"
	isPriv := true
	m := &client.V1Client{
		GetOciRegistryFn: func(uid string) (*models.V1EcrRegistry, error) {
			registryOcr := &models.V1EcrRegistry{
				Kind: "",
				Metadata: &models.V1ObjectMeta{
					Annotations:           nil,
					CreationTimestamp:     models.V1Time{},
					DeletionTimestamp:     models.V1Time{},
					Labels:                nil,
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "test-registry-oci",
					Namespace:             "",
					ResourceVersion:       "",
					SelfLink:              "",
					UID:                   "test-reg-oci",
				},
				Spec: &models.V1EcrRegistrySpec{
					Credentials: &models.V1AwsCloudAccount{
						AccessKey:      "ASDSDFRVDSVXCVSGDFGfd",
						CredentialType: "secret",
						Partition:      nil,
						PolicyARNs:     nil,
						SecretKey:      "ASDFXCVvzxcvasfea1234zcxzcZXCV",
						Sts:            nil,
					},
					DefaultRegion: "",
					Endpoint:      &endpoint,
					IsPrivate:     &isPriv,
					Scope:         "cluster",
				},
			}
			return registryOcr, nil
		},
	}
	ctx := context.Background()
	diags := resourceRegistryEcrRead(ctx, d, m)
	cre := d.Get("credentials")
	assert.Equal(t, "secret", cre.([]interface{})[0].(map[string]interface{})["credential_type"])
	assert.Equal(t, "ASDSDFRVDSVXCVSGDFGfd", cre.([]interface{})[0].(map[string]interface{})["access_key"])
	if len(diags) > 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
	if d.Id() != "test-reg-oci" {
		t.Errorf("Expected ID to be 'test-reg-oci', got %s", d.Id())
	}
}

func TestResourceRegistryEcrReadSTS(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSTS()
	d.SetId("test-reg-oci")
	endpoint := "214575254960.dkr.ecr.us-east-2.amazonaws.com"
	isPriv := true
	m := &client.V1Client{
		GetOciRegistryFn: func(uid string) (*models.V1EcrRegistry, error) {
			registryOcr := &models.V1EcrRegistry{
				Kind: "",
				Metadata: &models.V1ObjectMeta{
					Annotations:           nil,
					CreationTimestamp:     models.V1Time{},
					DeletionTimestamp:     models.V1Time{},
					Labels:                nil,
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "test-registry-oci",
					Namespace:             "",
					ResourceVersion:       "",
					SelfLink:              "",
					UID:                   "test-reg-oci",
				},
				Spec: &models.V1EcrRegistrySpec{
					Credentials: &models.V1AwsCloudAccount{
						AccessKey:      "",
						CredentialType: "sts",
						Partition:      nil,
						PolicyARNs:     nil,
						SecretKey:      "",
						Sts: &models.V1AwsStsCredentials{
							Arn:        "testARN",
							ExternalID: "testExternalID",
						},
					},
					DefaultRegion: "",
					Endpoint:      &endpoint,
					IsPrivate:     &isPriv,
					Scope:         "cluster",
				},
			}
			return registryOcr, nil
		},
	}
	ctx := context.Background()
	diags := resourceRegistryEcrRead(ctx, d, m)
	cre := d.Get("credentials")
	assert.Equal(t, "sts", cre.([]interface{})[0].(map[string]interface{})["credential_type"])
	assert.Equal(t, "testARN", cre.([]interface{})[0].(map[string]interface{})["arn"])
	assert.Equal(t, "testExternalID", cre.([]interface{})[0].(map[string]interface{})["external_id"])
	if len(diags) > 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
	if d.Id() != "test-reg-oci" {
		t.Errorf("Expected ID to be 'test-reg-oci', got %s", d.Id())
	}
}

func TestResourceRegistryEcrReadErr(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSTS()
	endpoint := "214575254960.dkr.ecr.us-east-2.amazonaws.com"
	isPriv := true
	m := &client.V1Client{
		GetOciRegistryFn: func(uid string) (*models.V1EcrRegistry, error) {
			registryOcr := &models.V1EcrRegistry{
				Kind: "",
				Metadata: &models.V1ObjectMeta{
					Annotations:           nil,
					CreationTimestamp:     models.V1Time{},
					DeletionTimestamp:     models.V1Time{},
					Labels:                nil,
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "test-registry-oci",
					Namespace:             "",
					ResourceVersion:       "",
					SelfLink:              "",
					UID:                   "test-reg-oci",
				},
				Spec: &models.V1EcrRegistrySpec{
					Credentials: &models.V1AwsCloudAccount{
						AccessKey:      "",
						CredentialType: "sts-wrong-type",
						Partition:      nil,
						PolicyARNs:     nil,
						SecretKey:      "",
						Sts: &models.V1AwsStsCredentials{
							Arn:        "testARN",
							ExternalID: "testExternalID",
						},
					},
					DefaultRegion: "",
					Endpoint:      &endpoint,
					IsPrivate:     &isPriv,
					Scope:         "cluster",
				},
			}
			return registryOcr, nil
		},
	}
	ctx := context.Background()
	diags := resourceRegistryEcrRead(ctx, d, m)
	if diags[0].Summary != "Registry type sts-wrong-type not implemented." {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
}

func TestResourceRegistryEcrReadNil(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSTS()
	m := &client.V1Client{
		GetOciRegistryFn: func(uid string) (*models.V1EcrRegistry, error) {
			return nil, errors.New("covering error case")
		},
	}
	ctx := context.Background()
	diags := resourceRegistryEcrRead(ctx, d, m)
	if diags[0].Summary != "covering error case" {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
}
func TestResourceRegistryEcrReadRegistryNil(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSTS()
	m := &client.V1Client{
		GetOciRegistryFn: func(uid string) (*models.V1EcrRegistry, error) {
			return nil, nil
		},
	}
	ctx := context.Background()
	resourceRegistryEcrRead(ctx, d, m)
	assert.Equal(t, "", d.Id())
}

func TestResourceRegistryEcrUpdate(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSTS()
	m := &client.V1Client{
		UpdateEcrRegistryFn: func(uid string, registry *models.V1EcrRegistry) error {
			return nil
		},
	}
	ctx := context.Background()
	diags := resourceRegistryEcrUpdate(ctx, d, m)
	assert.Equal(t, "", d.Id())
	if len(diags) > 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
}

func TestResourceRegistryEcrDelete(t *testing.T) {
	testCases := []struct {
		name                  string
		expectedReturnedUID   string
		expectedReturnedDiags diag.Diagnostics
		expectedError         error
		mock                  *mock.ClusterClientMock
	}{
		{
			name:                  "EcrDelete",
			expectedReturnedUID:   "",
			expectedReturnedDiags: diag.Diagnostics{},
			expectedError:         nil,
			mock: &mock.ClusterClientMock{
				DeleteEcrRegistryErr: nil,
			},
		},
		{
			name:                  "EcrDeleteErr",
			expectedReturnedUID:   "",
			expectedReturnedDiags: diag.FromErr(errors.New("covering error case")),
			expectedError:         errors.New("covering error case"),
			mock: &mock.ClusterClientMock{
				DeleteEcrRegistryErr: errors.New("covering error case"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			d := prepareOciEcrRegistryTestDataSTS()

			h := &client.V1Client{
				ClusterC: tc.mock,
				DeleteOciEcrRegistryFn: func(uid string) error {
					if t.Name() == "TestResourceRegistryEcrDelete/EcrDeleteErr" {
						return errors.New("covering error case")
					}
					return nil
				},
			}

			ctx := context.Background()
			diags := resourceRegistryEcrDelete(ctx, d, h)
			assert.Equal(t, "", d.Id())

			if len(diags) != len(tc.expectedReturnedDiags) {
				t.Fail()
				t.Logf("Expected diags count: %v", len(tc.expectedReturnedDiags))
				t.Logf("Actual diags count: %v", len(diags))
			} else {
				for i := range diags {
					if diags[i].Severity != tc.expectedReturnedDiags[i].Severity {
						t.Fail()
						t.Logf("Expected severity: %v", tc.expectedReturnedDiags[i].Severity)
						t.Logf("Actual severity: %v", diags[i].Severity)
					}
					if diags[i].Summary != tc.expectedReturnedDiags[i].Summary {
						t.Fail()
						t.Logf("Expected summary: %v", tc.expectedReturnedDiags[i].Summary)
						t.Logf("Actual summary: %v", diags[i].Summary)
					}
					if diags[i].Detail != tc.expectedReturnedDiags[i].Detail {
						t.Fail()
						t.Logf("Expected detail: %v", tc.expectedReturnedDiags[i].Detail)
						t.Logf("Actual detail: %v", diags[i].Detail)
					}
				}
			}
		})
	}

}

func TestResourceRegistryEcrUpdateErr(t *testing.T) {
	d := prepareOciEcrRegistryTestDataSTS()
	m := &client.V1Client{
		UpdateEcrRegistryFn: func(uid string, registry *models.V1EcrRegistry) error {
			return errors.New("covering error case")
		},
	}
	ctx := context.Background()
	diags := resourceRegistryEcrUpdate(ctx, d, m)
	assert.Equal(t, "", d.Id())
	if diags[0].Summary != "covering error case" {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
}
