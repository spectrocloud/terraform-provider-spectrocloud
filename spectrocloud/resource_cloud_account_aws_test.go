package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-api-go/models"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func TestToAwsAccountCTXProjectSecret(t *testing.T) {
	rd := resourceCloudAccountAws().TestResourceData()
	rd.Set("name", "aws_unit_test_acc")
	rd.Set("aws_access_key", "ABCDEFGHIJKLMNOPQRST")
	rd.Set("aws_secret_key", "sasf1424aqsfsdf123423SDFs23412sadf@#$@#$")
	rd.Set("context", "project")
	rd.Set("type", "secret")
	acc, err := toAwsAccount(rd)
	assert.NoError(t, err)

	assert.Equal(t, rd.Get("name"), acc.Metadata.Name)
	assert.Equal(t, rd.Get("aws_access_key"), acc.Spec.AccessKey)
	assert.Equal(t, rd.Get("aws_secret_key"), acc.Spec.SecretKey)
	assert.Equal(t, "project", acc.Metadata.Annotations["scope"])
	assert.Equal(t, rd.Get("type"), string(acc.Spec.CredentialType))
}

func TestToAwsAccountCTXTenantSecret(t *testing.T) {
	rd := resourceCloudAccountAws().TestResourceData()
	rd.Set("name", "aws_unit_test_acc")
	rd.Set("aws_access_key", "ABCDEFGHIJKLMNOPQRST")
	rd.Set("aws_secret_key", "sasf1424aqsfsdf123423SDFs23412sadf@#$@#$")
	rd.Set("context", "tenant")
	rd.Set("type", "secret")
	rd.Set("partition", "test_partition")
	acc, err := toAwsAccount(rd)
	assert.NoError(t, err)

	assert.Equal(t, rd.Get("name"), acc.Metadata.Name)
	assert.Equal(t, rd.Get("aws_access_key"), acc.Spec.AccessKey)
	assert.Equal(t, rd.Get("aws_secret_key"), acc.Spec.SecretKey)
	assert.Equal(t, "tenant", acc.Metadata.Annotations["scope"])
	assert.Equal(t, rd.Get("type"), string(acc.Spec.CredentialType))
	assert.Equal(t, rd.Get("partition"), *acc.Spec.Partition)
}

func TestToAwsAccountCTXProjectSTS(t *testing.T) {
	rd := resourceCloudAccountAws().TestResourceData()
	rd.Set("name", "aws_unit_test_acc")
	rd.Set("type", "sts")
	rd.Set("arn", "ARN::AWSAD:12312sdTEd")
	rd.Set("external_id", "TEST-External23423ID")
	rd.Set("context", "project")
	acc, err := toAwsAccount(rd)
	assert.NoError(t, err)

	assert.Equal(t, rd.Get("name"), acc.Metadata.Name)
	assert.Equal(t, rd.Get("arn"), acc.Spec.Sts.Arn)
	assert.Equal(t, rd.Get("external_id"), acc.Spec.Sts.ExternalID)
	assert.Equal(t, "project", acc.Metadata.Annotations["scope"])
	assert.Equal(t, rd.Get("type"), string(acc.Spec.CredentialType))
}

func TestToAwsAccountCTXTenantSTS(t *testing.T) {
	rd := resourceCloudAccountAws().TestResourceData()
	rd.Set("name", "aws_unit_test_acc")
	rd.Set("type", "sts")
	rd.Set("arn", "ARN::AWSAD:12312sdTEd")
	rd.Set("external_id", "TEST-External23423ID")
	rd.Set("context", "tenant")
	acc, err := toAwsAccount(rd)
	assert.NoError(t, err)

	assert.Equal(t, rd.Get("name"), acc.Metadata.Name)
	assert.Equal(t, rd.Get("arn"), acc.Spec.Sts.Arn)
	assert.Equal(t, rd.Get("external_id"), acc.Spec.Sts.ExternalID)
	assert.Equal(t, "tenant", acc.Metadata.Annotations["scope"])
	assert.Equal(t, rd.Get("type"), string(acc.Spec.CredentialType))
}

func TestFlattenCloudAccountAwsSTS(t *testing.T) {
	// Create a mock ResourceData object
	rd := resourceCloudAccountAws().TestResourceData() // Assuming this method exists

	// Create a mock AWS account model
	account := &models.V1AwsAccount{
		Metadata: &models.V1ObjectMeta{
			Name: "aws_test_account",
			Annotations: map[string]string{
				"scope": "aws_scope_test",
			},
		},
		Spec: &models.V1AwsCloudAccount{
			CredentialType: models.V1AwsCloudAccountCredentialTypeSts,
			Sts:            &models.V1AwsStsCredentials{Arn: "test_arn"},
			Partition:      types.Ptr("test_partition"),
			PolicyARNs:     []string{"arn:aws:test_policy1", "arn:aws:test_policy2"},
		},
	}

	// Call the flatten function
	diags, hasError := flattenCloudAccountAws(rd, account)

	// Assertions
	assert.Nil(t, diags)
	assert.False(t, hasError)
	assert.Equal(t, "aws_test_account", rd.Get("name"))
	assert.Equal(t, "aws_scope_test", rd.Get("context"))
	assert.Equal(t, "test_arn", rd.Get("arn"))
	assert.Equal(t, "test_partition", rd.Get("partition"))
	assert.Equal(t, string(models.V1AwsCloudAccountCredentialTypeSts), rd.Get("type"))

	// Handle policy_arns as a *schema.Set
	policyARNs, ok := rd.Get("policy_arns").(*schema.Set)
	if !ok {
		t.Fatalf("Expected policy_arns to be a *schema.Set")
	}

	var actualARNs []string
	for _, v := range policyARNs.List() {
		actualARNs = append(actualARNs, v.(string))
	}

	expectedARNs := []string{"arn:aws:test_policy1", "arn:aws:test_policy2"}
	assert.ElementsMatch(t, expectedARNs, actualARNs)
}

func TestFlattenCloudAccountAws_NonStsType(t *testing.T) {
	rd := resourceCloudAccountAws().TestResourceData()

	account := &models.V1AwsAccount{
		Metadata: &models.V1ObjectMeta{
			Name: "aws_test_account_secret",
			Annotations: map[string]string{
				"scope": "aws_scope_test_secret",
			},
		},
		Spec: &models.V1AwsCloudAccount{
			CredentialType: models.V1AwsCloudAccountCredentialTypeSecret,
			AccessKey:      "test_access_key_secret",
			Partition:      types.Ptr("test_partition_secret"),
			PolicyARNs:     []string{"arn:aws:test_policy_secret1", "arn:aws:test_policy_secret2"},
		},
	}

	// Call the flatten function
	diags, hasError := flattenCloudAccountAws(rd, account)

	// Assertions
	assert.Nil(t, diags)
	assert.False(t, hasError)
	assert.Equal(t, "aws_test_account_secret", rd.Get("name"))
	assert.Equal(t, "aws_scope_test_secret", rd.Get("context"))
	assert.Equal(t, "test_access_key_secret", rd.Get("aws_access_key"))
	assert.Empty(t, rd.Get("arn")) // Asserting that arn is not set
	assert.Equal(t, "test_partition_secret", rd.Get("partition"))

	// Handle policy_arns as a *schema.Set
	policyARNs, ok := rd.Get("policy_arns").(*schema.Set)
	if !ok {
		t.Fatalf("Expected policy_arns to be a *schema.Set")
	}

	var actualARNs []string
	for _, v := range policyARNs.List() {
		actualARNs = append(actualARNs, v.(string))
	}

	expectedARNs := []string{"arn:aws:test_policy_secret1", "arn:aws:test_policy_secret2"}
	assert.ElementsMatch(t, expectedARNs, actualARNs)
}

//func TestFlattenCloudAccountAws(t *testing.T) {
//	testCases := []struct {
//		name         string
//		inputAccount map[string]interface{}
//		expectedData *models.V1AwsAccount
//	}{
//		{
//			name: "Successful flattening with AWS access key",
//			inputAccount: map[string]interface{}{
//				"name":           "test-account",
//				"context":        "test-context",
//				"type":           "secret",
//				"aws_access_key": "access-key",
//				"partition":      "aws-partition",
//				"policy_arns":    []interface{}{"policy-arn-1", "policy-arn-2"},
//			},
//			expectedData: &models.V1AwsAccount{
//				Metadata: &models.V1ObjectMeta{
//					Name: "test-account",
//					Annotations: map[string]string{
//						"scope": "test-context",
//					},
//				},
//				Spec: &models.V1AwsCloudAccount{
//					CredentialType: "secret",
//					AccessKey:      "access-key",
//					Partition:      types.Ptr("aws-partition"),
//					PolicyARNs:     []string{"policy-arn-1", "policy-arn-2"},
//				},
//			},
//		},
//		{
//			name: "Successful flattening with ARN",
//			inputAccount: map[string]interface{}{
//				"name":        "test-account",
//				"context":     "test-context",
//				"type":        "arn",
//				"arn":         "arn:aws:sts::123456789012:assumed-role/role-name",
//				"partition":   "aws-partition",
//				"policy_arns": []interface{}{"policy-arn-1", "policy-arn-2"},
//			},
//			expectedData: &models.V1AwsAccount{
//				Metadata: &models.V1ObjectMeta{
//					Name: "test-account",
//					Annotations: map[string]string{
//						"scope": "test-context",
//					},
//				},
//				Spec: &models.V1AwsCloudAccount{
//					CredentialType: "arn",
//					Sts: &models.V1AwsStsCredentials{
//						Arn: "arn:aws:sts::123456789012:assumed-role/role-name",
//					},
//					Partition:  types.Ptr("aws-partition"),
//					PolicyARNs: []string{"policy-arn-1", "policy-arn-2"},
//				},
//			},
//		},
//		{
//			name: "Flattening with empty fields",
//			inputAccount: map[string]interface{}{
//				"name":        "test-account",
//				"context":     "test-context",
//				"type":        "arn",
//				"arn":         "arn:aws:sts::123456789012:assumed-role/role-name",
//				"policy_arns": nil,
//			},
//			expectedData: &models.V1AwsAccount{
//				Metadata: &models.V1ObjectMeta{
//					Name: "test-account",
//					Annotations: map[string]string{
//						"scope": "test-context",
//					},
//				},
//				Spec: &models.V1AwsCloudAccount{
//					CredentialType: "arn",
//					Sts: &models.V1AwsStsCredentials{
//						Arn: "arn:aws:sts::123456789012:assumed-role/role-name",
//					},
//					PolicyARNs: nil,
//				},
//			},
//		},
//	}
//
//	for _, tc := range testCases {
//		t.Run(tc.name, func(t *testing.T) {
//			// Create ResourceData with test values
//			resourceData := schema.TestResourceDataRaw(t, resourceCloudAccountAws().Schema, tc.inputAccount)
//
//			diags, _ := flattenCloudAccountAws(resourceData, tc.expectedData)
//
//			assert.Nil(t, diags) // Expect no diagnostics (errors)
//
//			// Convert *schema.Set to []string for comparison
//			if policyARNs, ok := tc.inputAccount["policy_arns"].([]interface{}); ok {
//				expectedPolicyARNs := convertInterfaceSliceToStringSlice(policyARNs)
//				actualPolicyARNs := convertSchemaSetToStringSlice(resourceData.Get("policy_arns").(*schema.Set))
//
//				// Sort both slices before comparison
//				sort.Strings(expectedPolicyARNs)
//				sort.Strings(actualPolicyARNs)
//
//				assert.Equal(t, expectedPolicyARNs, actualPolicyARNs, "Mismatch in field: policy_arns")
//			}
//
//			// Compare other fields
//			for key, expectedValue := range tc.inputAccount {
//				if key == "policy_arns" {
//					continue
//				}
//				actualValue := resourceData.Get(key)
//				assert.Equal(t, expectedValue, actualValue, "Mismatch in field: "+key)
//			}
//		})
//	}
//}
//
//// Helper function to convert *schema.Set to []string
//func convertSchemaSetToStringSlice(set *schema.Set) []string {
//	var result []string
//	for _, v := range set.List() {
//		result = append(result, v.(string))
//	}
//	return result
//}
//
//// Helper function to convert []interface{} to []string
//func convertInterfaceSliceToStringSlice(slice []interface{}) []string {
//	var result []string
//	for _, v := range slice {
//		result = append(result, v.(string))
//	}
//	return result
//}
