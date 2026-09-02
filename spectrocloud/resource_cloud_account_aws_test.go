package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func TestToAwsAccount(t *testing.T) {
	secretKey := "sasf1424aqsfsdf123423SDFs23412sadf@#$@#$"
	tests := []struct {
		name   string
		input  map[string]interface{}
		verify func(t *testing.T, acc *models.V1AwsAccount)
	}{
		{
			name: "CTX project secret",
			input: map[string]interface{}{
				"name": "aws_unit_test_acc", "aws_access_key": "ABCDEFGHIJKLMNOPQRST", "aws_secret_key": secretKey,
				"context": "project", "type": "secret",
			},
			verify: func(t *testing.T, acc *models.V1AwsAccount) {
				assert.Equal(t, "aws_unit_test_acc", acc.Metadata.Name)
				assert.Equal(t, "ABCDEFGHIJKLMNOPQRST", acc.Spec.AccessKey)
				assert.Equal(t, secretKey, acc.Spec.SecretKey)
				assert.Equal(t, "project", acc.Metadata.Annotations["scope"])
				assert.Equal(t, "secret", string(*acc.Spec.CredentialType))
			},
		},
		{
			name: "CTX tenant secret",
			input: map[string]interface{}{
				"name": "aws_unit_test_acc", "aws_access_key": "ABCDEFGHIJKLMNOPQRST", "aws_secret_key": secretKey,
				"context": "tenant", "type": "secret", "partition": "test_partition",
			},
			verify: func(t *testing.T, acc *models.V1AwsAccount) {
				assert.Equal(t, "aws_unit_test_acc", acc.Metadata.Name)
				assert.Equal(t, "ABCDEFGHIJKLMNOPQRST", acc.Spec.AccessKey)
				assert.Equal(t, secretKey, acc.Spec.SecretKey)
				assert.Equal(t, "tenant", acc.Metadata.Annotations["scope"])
				assert.Equal(t, "secret", string(*acc.Spec.CredentialType))
				assert.Equal(t, "test_partition", *acc.Spec.Partition)
			},
		},
		{
			name: "CTX project secured access key",
			input: map[string]interface{}{
				"name": "aws_unit_test_acc_secured", "aws_secured_access_key": "ABCDEFGHIJKLMNOPQRST", "aws_secret_key": secretKey,
				"context": "project", "type": "secret",
			},
			verify: func(t *testing.T, acc *models.V1AwsAccount) {
				assert.Equal(t, "aws_unit_test_acc_secured", acc.Metadata.Name)
				assert.Equal(t, "ABCDEFGHIJKLMNOPQRST", acc.Spec.AccessKey)
				assert.Equal(t, secretKey, acc.Spec.SecretKey)
				assert.Equal(t, "project", acc.Metadata.Annotations["scope"])
				assert.Equal(t, "secret", string(*acc.Spec.CredentialType))
			},
		},
		{
			name: "CTX tenant secured access key",
			input: map[string]interface{}{
				"name": "aws_unit_test_acc_secured", "aws_secured_access_key": "ABCDEFGHIJKLMNOPQRST", "aws_secret_key": secretKey,
				"context": "tenant", "type": "secret", "partition": "test_partition",
			},
			verify: func(t *testing.T, acc *models.V1AwsAccount) {
				assert.Equal(t, "aws_unit_test_acc_secured", acc.Metadata.Name)
				assert.Equal(t, "ABCDEFGHIJKLMNOPQRST", acc.Spec.AccessKey)
				assert.Equal(t, secretKey, acc.Spec.SecretKey)
				assert.Equal(t, "tenant", acc.Metadata.Annotations["scope"])
				assert.Equal(t, "secret", string(*acc.Spec.CredentialType))
				assert.Equal(t, "test_partition", *acc.Spec.Partition)
			},
		},
		{
			name: "secured access key priority",
			input: map[string]interface{}{
				"name": "aws_unit_test_acc_priority", "aws_secured_access_key": "SECURED_ACCESS_KEY_123", "aws_secret_key": secretKey,
				"context": "project", "type": "secret",
			},
			verify: func(t *testing.T, acc *models.V1AwsAccount) {
				assert.Equal(t, "SECURED_ACCESS_KEY_123", acc.Spec.AccessKey)
				assert.Equal(t, secretKey, acc.Spec.SecretKey)
			},
		},
		{
			name: "both access keys set",
			input: map[string]interface{}{
				"name": "aws_unit_test_acc_priority", "aws_access_key": "LEGACY_ACCESS_KEY_123", "aws_secured_access_key": "SECURED_ACCESS_KEY_123",
				"aws_secret_key": secretKey, "context": "project", "type": "secret",
			},
			verify: func(t *testing.T, acc *models.V1AwsAccount) {
				assert.NotNil(t, acc)
				assert.Equal(t, "SECURED_ACCESS_KEY_123", acc.Spec.AccessKey)
				assert.Equal(t, secretKey, acc.Spec.SecretKey)
			},
		},
		{
			name: "CTX project STS",
			input: map[string]interface{}{
				"name": "aws_unit_test_acc", "type": "sts", "arn": "ARN::AWSAD:12312sdTEd",
				"external_id": "TEST-External23423ID", "context": "project",
			},
			verify: func(t *testing.T, acc *models.V1AwsAccount) {
				assert.Equal(t, "aws_unit_test_acc", acc.Metadata.Name)
				assert.Equal(t, "ARN::AWSAD:12312sdTEd", acc.Spec.Sts.Arn)
				assert.Equal(t, "TEST-External23423ID", acc.Spec.Sts.ExternalID)
				assert.Equal(t, "project", acc.Metadata.Annotations["scope"])
				assert.Equal(t, "sts", string(*acc.Spec.CredentialType))
			},
		},
		{
			name: "CTX tenant STS",
			input: map[string]interface{}{
				"name": "aws_unit_test_acc", "type": "sts", "arn": "ARN::AWSAD:12312sdTEd",
				"external_id": "TEST-External23423ID", "context": "tenant",
			},
			verify: func(t *testing.T, acc *models.V1AwsAccount) {
				assert.Equal(t, "aws_unit_test_acc", acc.Metadata.Name)
				assert.Equal(t, "ARN::AWSAD:12312sdTEd", acc.Spec.Sts.Arn)
				assert.Equal(t, "TEST-External23423ID", acc.Spec.Sts.ExternalID)
				assert.Equal(t, "tenant", acc.Metadata.Annotations["scope"])
				assert.Equal(t, "sts", string(*acc.Spec.CredentialType))
			},
		},
		{
			name: "CTX project pod identity",
			input: map[string]interface{}{
				"name": "aws_unit_test_acc_pod_identity", "type": "pod-identity",
				"role_arn":                "arn:aws:iam::123456789012:role/EKSPodIdentityRole",
				"permission_boundary_arn": "arn:aws:iam::123456789012:policy/PermissionBoundary", "context": "project",
			},
			verify: func(t *testing.T, acc *models.V1AwsAccount) {
				assert.Equal(t, "aws_unit_test_acc_pod_identity", acc.Metadata.Name)
				assert.Equal(t, "arn:aws:iam::123456789012:role/EKSPodIdentityRole", acc.Spec.PodIdentity.RoleArn)
				assert.Equal(t, "arn:aws:iam::123456789012:policy/PermissionBoundary", acc.Spec.PodIdentity.PermissionBoundaryArn)
				assert.Equal(t, "project", acc.Metadata.Annotations["scope"])
				assert.Equal(t, "pod-identity", string(*acc.Spec.CredentialType))
			},
		},
		{
			name: "CTX tenant pod identity",
			input: map[string]interface{}{
				"name": "aws_unit_test_acc_pod_identity_tenant", "type": "pod-identity",
				"role_arn":                "arn:aws:iam::123456789012:role/EKSPodIdentityRole",
				"permission_boundary_arn": "arn:aws:iam::123456789012:policy/PermissionBoundary",
				"context":                 "tenant", "partition": "aws",
			},
			verify: func(t *testing.T, acc *models.V1AwsAccount) {
				assert.Equal(t, "aws_unit_test_acc_pod_identity_tenant", acc.Metadata.Name)
				assert.Equal(t, "arn:aws:iam::123456789012:role/EKSPodIdentityRole", acc.Spec.PodIdentity.RoleArn)
				assert.Equal(t, "arn:aws:iam::123456789012:policy/PermissionBoundary", acc.Spec.PodIdentity.PermissionBoundaryArn)
				assert.Equal(t, "tenant", acc.Metadata.Annotations["scope"])
				assert.Equal(t, "pod-identity", string(*acc.Spec.CredentialType))
				assert.Equal(t, "aws", *acc.Spec.Partition)
			},
		},
		{
			name: "pod identity without permission boundary",
			input: map[string]interface{}{
				"name": "aws_unit_test_acc_pod_identity_no_boundary", "type": "pod-identity",
				"role_arn": "arn:aws:iam::123456789012:role/EKSPodIdentityRole", "context": "project",
			},
			verify: func(t *testing.T, acc *models.V1AwsAccount) {
				assert.Equal(t, "aws_unit_test_acc_pod_identity_no_boundary", acc.Metadata.Name)
				assert.Equal(t, "arn:aws:iam::123456789012:role/EKSPodIdentityRole", acc.Spec.PodIdentity.RoleArn)
				assert.Empty(t, acc.Spec.PodIdentity.PermissionBoundaryArn)
				assert.Equal(t, "project", acc.Metadata.Annotations["scope"])
				assert.Equal(t, "pod-identity", string(*acc.Spec.CredentialType))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := resourceCloudAccountAws().TestResourceData()
			for k, v := range tt.input {
				rd.Set(k, v)
			}
			acc, err := toAwsAccount(rd)
			assert.NoError(t, err)
			tt.verify(t, acc)
		})
	}
}

// assertFlattenResult checks expected fields on rd after flatten; expect keys with nil are skipped, *string "" means assert empty.
func assertFlattenResult(t *testing.T, rd *schema.ResourceData, expect map[string]*string, policyARNs []string) {
	t.Helper()
	for k, v := range expect {
		if v == nil {
			continue
		}
		if *v == "" {
			assert.Empty(t, rd.Get(k), "field %s", k)
		} else {
			assert.Equal(t, *v, rd.Get(k), "field %s", k)
		}
	}
	if policyARNs != nil {
		set, ok := rd.Get("policy_arns").(*schema.Set)
		assert.True(t, ok, "policy_arns should be *schema.Set")
		var actual []string
		for _, x := range set.List() {
			actual = append(actual, x.(string))
		}
		assert.ElementsMatch(t, policyARNs, actual)
	}
}

func TestFlattenCloudAccountAws_TableDriven(t *testing.T) {
	scenarios := []struct {
		name             string
		account          *models.V1AwsAccount
		rdPreSet         map[string]interface{}
		expect           map[string]*string
		expectPolicyARNs []string
	}{
		{
			name: "STS",
			account: &models.V1AwsAccount{
				Metadata: &models.V1ObjectMeta{
					Name:        "aws_test_account",
					Annotations: map[string]string{"scope": "aws_scope_test"},
				},
				Spec: &models.V1AwsCloudAccount{
					CredentialType: models.V1AwsCloudAccountCredentialTypeSts.Pointer(),
					Sts:            &models.V1AwsStsCredentials{Arn: "test_arn"},
					Partition:      types.Ptr("test_partition"),
					PolicyARNs:     []string{"arn:aws:test_policy1", "arn:aws:test_policy2"},
				},
			},
			expect: map[string]*string{
				"name": types.Ptr("aws_test_account"), "context": types.Ptr("aws_scope_test"),
				"arn": types.Ptr("test_arn"), "partition": types.Ptr("test_partition"),
				"type": types.Ptr(string(models.V1AwsCloudAccountCredentialTypeSts)),
			},
			expectPolicyARNs: []string{"arn:aws:test_policy1", "arn:aws:test_policy2"},
		},
		{
			name: "secret (non-STS)",
			account: &models.V1AwsAccount{
				Metadata: &models.V1ObjectMeta{
					Name:        "aws_test_account_secret",
					Annotations: map[string]string{"scope": "aws_scope_test_secret"},
				},
				Spec: &models.V1AwsCloudAccount{
					CredentialType: models.V1AwsCloudAccountCredentialTypeSecret.Pointer(),
					AccessKey:      "test_access_key_secret",
					Partition:      types.Ptr("test_partition_secret"),
					PolicyARNs:     []string{"arn:aws:test_policy_secret1", "arn:aws:test_policy_secret2"},
				},
			},
			expect: map[string]*string{
				"name": types.Ptr("aws_test_account_secret"), "context": types.Ptr("aws_scope_test_secret"),
				"aws_access_key": types.Ptr("test_access_key_secret"), "arn": types.Ptr(""),
				"partition": types.Ptr("test_partition_secret"), "type": types.Ptr(string(models.V1AwsCloudAccountCredentialTypeSecret)),
			},
			expectPolicyARNs: []string{"arn:aws:test_policy_secret1", "arn:aws:test_policy_secret2"},
		},
		{
			name: "secret with secured access key (rd pre-set)",
			account: &models.V1AwsAccount{
				Metadata: &models.V1ObjectMeta{
					Name:        "aws_test_account_secured",
					Annotations: map[string]string{"scope": "aws_scope_test_secured"},
				},
				Spec: &models.V1AwsCloudAccount{
					CredentialType: models.V1AwsCloudAccountCredentialTypeSecret.Pointer(),
					AccessKey:      "test_secured_access_key",
					Partition:      types.Ptr("test_partition_secured"),
					PolicyARNs:     []string{"arn:aws:test_policy_secured1"},
				},
			},
			rdPreSet: map[string]interface{}{"aws_secured_access_key": "existing_secured_key"},
			expect: map[string]*string{
				"name": types.Ptr("aws_test_account_secured"), "context": types.Ptr("aws_scope_test_secured"),
				"aws_secured_access_key": types.Ptr("test_secured_access_key"), "aws_access_key": types.Ptr(""),
				"partition": types.Ptr("test_partition_secured"),
			},
			expectPolicyARNs: []string{"arn:aws:test_policy_secured1"},
		},
		{
			name: "legacy access key",
			account: &models.V1AwsAccount{
				Metadata: &models.V1ObjectMeta{
					Name:        "aws_test_account_legacy",
					Annotations: map[string]string{"scope": "project"},
				},
				Spec: &models.V1AwsCloudAccount{
					CredentialType: models.V1AwsCloudAccountCredentialTypeSecret.Pointer(),
					AccessKey:      "test_legacy_access_key",
					Partition:      types.Ptr("aws"),
				},
			},
			expect: map[string]*string{
				"name": types.Ptr("aws_test_account_legacy"), "context": types.Ptr("project"),
				"aws_access_key": types.Ptr("test_legacy_access_key"), "aws_secured_access_key": types.Ptr(""),
				"partition": types.Ptr("aws"),
			},
		},
		{
			name: "switch from secured to legacy (keeps secured in state)",
			account: &models.V1AwsAccount{
				Metadata: &models.V1ObjectMeta{
					Name:        "aws_test_account_switch",
					Annotations: map[string]string{"scope": "project"},
				},
				Spec: &models.V1AwsCloudAccount{
					CredentialType: models.V1AwsCloudAccountCredentialTypeSecret.Pointer(),
					AccessKey:      "new_access_key",
					Partition:      types.Ptr("aws"),
				},
			},
			rdPreSet: map[string]interface{}{"aws_secured_access_key": "old_secured_key"},
			expect: map[string]*string{
				"name": types.Ptr("aws_test_account_switch"), "context": types.Ptr("project"),
				"aws_secured_access_key": types.Ptr("new_access_key"), "aws_access_key": types.Ptr(""),
				"partition": types.Ptr("aws"),
			},
		},
		{
			name: "clear conflicting field legacy",
			account: &models.V1AwsAccount{
				Metadata: &models.V1ObjectMeta{
					Name:        "aws_test_account_clear",
					Annotations: map[string]string{"scope": "project"},
				},
				Spec: &models.V1AwsCloudAccount{
					CredentialType: models.V1AwsCloudAccountCredentialTypeSecret.Pointer(),
					AccessKey:      "legacy_access_key",
					Partition:      types.Ptr("aws"),
				},
			},
			expect: map[string]*string{
				"name": types.Ptr("aws_test_account_clear"), "context": types.Ptr("project"),
				"aws_access_key": types.Ptr("legacy_access_key"), "aws_secured_access_key": types.Ptr(""),
				"partition": types.Ptr("aws"),
			},
		},
		{
			name: "pod identity with permission boundary",
			account: &models.V1AwsAccount{
				Metadata: &models.V1ObjectMeta{
					Name:        "aws_test_account_pod_identity",
					Annotations: map[string]string{"scope": "project"},
				},
				Spec: &models.V1AwsCloudAccount{
					CredentialType: models.V1AwsCloudAccountCredentialTypePodDashIdentity.Pointer(),
					PodIdentity: &models.V1AwsPodIdentityCredentials{
						RoleArn:               "arn:aws:iam::123456789012:role/EKSPodIdentityRole",
						PermissionBoundaryArn: "arn:aws:iam::123456789012:policy/PermissionBoundary",
					},
					Partition: types.Ptr("aws"),
				},
			},
			expect: map[string]*string{
				"name": types.Ptr("aws_test_account_pod_identity"), "context": types.Ptr("project"),
				"role_arn":                types.Ptr("arn:aws:iam::123456789012:role/EKSPodIdentityRole"),
				"permission_boundary_arn": types.Ptr("arn:aws:iam::123456789012:policy/PermissionBoundary"),
				"partition":               types.Ptr("aws"), "type": types.Ptr(string(models.V1AwsCloudAccountCredentialTypePodDashIdentity)),
			},
		},
		{
			name: "pod identity without permission boundary",
			account: &models.V1AwsAccount{
				Metadata: &models.V1ObjectMeta{
					Name:        "aws_test_account_pod_identity_no_boundary",
					Annotations: map[string]string{"scope": "tenant"},
				},
				Spec: &models.V1AwsCloudAccount{
					CredentialType: models.V1AwsCloudAccountCredentialTypePodDashIdentity.Pointer(),
					PodIdentity: &models.V1AwsPodIdentityCredentials{
						RoleArn: "arn:aws:iam::123456789012:role/EKSPodIdentityRole",
					},
					Partition:  types.Ptr("aws-us-gov"),
					PolicyARNs: []string{"arn:aws:iam::123456789012:policy/CustomPolicy"},
				},
			},
			expect: map[string]*string{
				"name": types.Ptr("aws_test_account_pod_identity_no_boundary"), "context": types.Ptr("tenant"),
				"role_arn":                types.Ptr("arn:aws:iam::123456789012:role/EKSPodIdentityRole"),
				"permission_boundary_arn": types.Ptr(""), "partition": types.Ptr("aws-us-gov"),
				"type": types.Ptr(string(models.V1AwsCloudAccountCredentialTypePodDashIdentity)),
			},
			expectPolicyARNs: []string{"arn:aws:iam::123456789012:policy/CustomPolicy"},
		},
	}
	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			rd := resourceCloudAccountAws().TestResourceData()
			for k, v := range s.rdPreSet {
				rd.Set(k, v)
			}
			diags, hasError := flattenCloudAccountAws(rd, s.account)
			assert.Nil(t, diags)
			assert.False(t, hasError)
			assertFlattenResult(t, rd, s.expect, s.expectPolicyARNs)
		})
	}
}

func prepareBaseAwsAccountTestData() *schema.ResourceData {
	d := resourceCloudAccountAws().TestResourceData()
	d.SetId("test-aws-account-1")
	_ = d.Set("name", "test-aws-account")
	_ = d.Set("context", "project")
	_ = d.Set("aws_access_key", "test-access-key")
	_ = d.Set("aws_secret_key", "test-secret-key")
	_ = d.Set("type", "secret")
	_ = d.Set("arn", "test-arn")
	_ = d.Set("external_id", "test-external-id")
	_ = d.Set("partition", "aws")
	_ = d.Set("policy_arns", []string{"test-policy-arn"})
	return d
}

func prepareSecuredAwsAccountTestData() *schema.ResourceData {
	d := resourceCloudAccountAws().TestResourceData()
	d.SetId("test-aws-account-1")
	_ = d.Set("name", "test-aws-account-secured")
	_ = d.Set("context", "project")
	_ = d.Set("aws_secured_access_key", "test-secured-access-key")
	_ = d.Set("aws_secret_key", "test-secret-key")
	_ = d.Set("type", "secret")
	_ = d.Set("partition", "aws")
	_ = d.Set("policy_arns", []string{"test-policy-arn"})
	return d
}

func TestResourceAwsAccountCRUD(t *testing.T) {
	testResourceCRUD(t, prepareSecuredAwsAccountTestData, unitTestMockAPIClient,
		resourceCloudAccountAwsCreate, resourceCloudAccountAwsRead, resourceCloudAccountAwsUpdate, resourceCloudAccountAwsDelete)
}
func TestResourceCloudAccountAwsImport(t *testing.T) {
	ctx := context.Background()
	d := prepareBaseAwsAccountTestData()
	d.SetId("test-import-acc-id:project")
	_, err := resourceAccountAwsImport(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, err)
	assert.Equal(t, "test-import-acc-id", d.Id())
}

func TestResourceCloudAccountAwsCreateWithSecuredAccessKey(t *testing.T) {
	ctx := context.Background()
	d := prepareSecuredAwsAccountTestData()
	diags := resourceCloudAccountAwsCreate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-aws-account-1", d.Id())
}

func TestResourceCloudAccountAwsReadWithSecuredAccessKey(t *testing.T) {
	ctx := context.Background()
	d := prepareSecuredAwsAccountTestData()
	diags := resourceCloudAccountAwsRead(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-aws-account-1", d.Id())
}

func TestResourceCloudAccountAwsUpdateWithSecuredAccessKey(t *testing.T) {
	ctx := context.Background()
	d := prepareSecuredAwsAccountTestData()
	diags := resourceCloudAccountAwsUpdate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-aws-account-1", d.Id())
}

func TestResourceCloudAccountAwsDeleteWithSecuredAccessKey(t *testing.T) {
	ctx := context.Background()
	d := prepareSecuredAwsAccountTestData()
	diags := resourceCloudAccountAwsDelete(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
}

// ==================== Pod Identity Tests ====================

func preparePodIdentityAwsAccountTestData() *schema.ResourceData {
	d := resourceCloudAccountAws().TestResourceData()
	d.SetId("test-aws-account-1")
	_ = d.Set("name", "test-aws-account-pod-identity")
	_ = d.Set("context", "project")
	_ = d.Set("type", "pod-identity")
	_ = d.Set("role_arn", "arn:aws:iam::123456789012:role/EKSPodIdentityRole")
	_ = d.Set("permission_boundary_arn", "arn:aws:iam::123456789012:policy/PermissionBoundary")
	_ = d.Set("partition", "aws")
	_ = d.Set("policy_arns", []string{"arn:aws:iam::123456789012:policy/TestPolicy"})
	return d
}

func TestResourceCloudAccountAwsCreateWithPodIdentity(t *testing.T) {
	ctx := context.Background()
	d := preparePodIdentityAwsAccountTestData()
	diags := resourceCloudAccountAwsCreate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-aws-account-1", d.Id())
}

func TestResourceCloudAccountAwsReadWithPodIdentity(t *testing.T) {
	ctx := context.Background()
	d := preparePodIdentityAwsAccountTestData()
	diags := resourceCloudAccountAwsRead(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-aws-account-1", d.Id())
}

func TestResourceCloudAccountAwsUpdateWithPodIdentity(t *testing.T) {
	ctx := context.Background()
	d := preparePodIdentityAwsAccountTestData()
	diags := resourceCloudAccountAwsUpdate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-aws-account-1", d.Id())
}

func TestResourceCloudAccountAwsDeleteWithPodIdentity(t *testing.T) {
	ctx := context.Background()
	d := preparePodIdentityAwsAccountTestData()
	diags := resourceCloudAccountAwsDelete(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
}
