package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
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
