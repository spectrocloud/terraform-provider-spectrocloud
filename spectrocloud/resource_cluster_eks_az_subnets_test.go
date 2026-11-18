package spectrocloud

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

// PLT-2468: import of a static EKS cluster must read the control-plane subnets into cloud_config.az_subnets,
// whatever the control-plane pool is called, without making a dynamic cluster read back as static.

func eksTestCloudConfig(vpcID string, pools ...*models.V1EksMachinePoolConfig) *models.V1EksCloudConfig {
	return &models.V1EksCloudConfig{
		Spec: &models.V1EksCloudConfigSpec{
			ClusterConfig: &models.V1EksClusterConfig{
				Region:         types.Ptr("us-east-2"),
				EndpointAccess: &models.V1EksClusterConfigEndpointAccess{Public: true},
				VpcID:          vpcID,
			},
			MachinePoolConfig: pools,
		},
	}
}

func eksTestWorkerPool(subnets map[string]string) *models.V1EksMachinePoolConfig {
	return &models.V1EksMachinePoolConfig{
		Name:           "worker-pool",
		IsControlPlane: types.Ptr(false),
		Azs:            []string{"us-east-2a", "us-east-2c"},
		SubnetIds:      subnets,
	}
}

func eksFlattenedAzSubnets(t *testing.T, cfg *models.V1EksCloudConfig) (interface{}, bool) {
	t.Helper()
	flattened := flattenClusterConfigsEKS(cfg).([]interface{})
	require.Len(t, flattened, 1)
	v, ok := flattened[0].(map[string]interface{})["az_subnets"]
	return v, ok
}

func TestFlattenClusterConfigsEKS_ControlPlaneAzSubnets(t *testing.T) {
	// Shape returned by GET /v1/cloudconfigs/eks/{uid} for a UI-created static cluster.
	cpSubnets := map[string]string{
		"us-east-2a": "subnet-0ba60dc55d1bd5756,subnet-05ad5487490bc874f",
		"us-east-2c": "subnet-033a12c33a54af7a6,subnet-0967d4215d2ddf43f",
	}
	workerSubnets := map[string]string{"us-east-2a": "subnet-0ba60dc55d1bd5756", "us-east-2c": "subnet-0967d4215d2ddf43f"}

	tests := []struct {
		name    string
		cfg     *models.V1EksCloudConfig
		want    map[string]string
		wantSet bool
	}{
		{
			name: "static cluster created in the UI (control-plane-pool)",
			cfg: eksTestCloudConfig("vpc-0b69e5c29603de04e",
				&models.V1EksMachinePoolConfig{Name: "control-plane-pool", IsControlPlane: types.Ptr(true), Azs: []string{"us-east-2a", "us-east-2c"}, SubnetIds: cpSubnets},
				eksTestWorkerPool(workerSubnets)),
			want: cpSubnets, wantSet: true,
		},
		{
			name: "static cluster created by Terraform (cp-pool)",
			cfg: eksTestCloudConfig("vpc-0b69e5c29603de04e",
				eksTestWorkerPool(workerSubnets),
				&models.V1EksMachinePoolConfig{Name: "cp-pool", IsControlPlane: types.Ptr(true), SubnetIds: cpSubnets}),
			want: cpSubnets, wantSet: true,
		},
		{
			name: "cp-pool without isControlPlane still matches by name",
			cfg: eksTestCloudConfig("vpc-0b69e5c29603de04e",
				&models.V1EksMachinePoolConfig{Name: "cp-pool", SubnetIds: cpSubnets}),
			want: cpSubnets, wantSet: true,
		},
		{
			name: "dynamic cluster has no control-plane pool",
			cfg:  eksTestCloudConfig("", eksTestWorkerPool(nil)),
		},
		{
			name: "control-plane pool without subnets stays dynamic",
			cfg: eksTestCloudConfig("",
				&models.V1EksMachinePoolConfig{Name: "control-plane-pool", IsControlPlane: types.Ptr(true), Azs: []string{"us-east-2a"}}),
		},
		{
			name: "worker subnets are never used for the control plane",
			cfg:  eksTestCloudConfig("vpc-0b69e5c29603de04e", eksTestWorkerPool(workerSubnets)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := eksFlattenedAzSubnets(t, tt.cfg)
			assert.Equal(t, tt.wantSet, ok)
			if tt.wantSet {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestEksControlPlanePool(t *testing.T) {
	cp := &models.V1EksMachinePoolConfig{Name: "control-plane-pool", IsControlPlane: types.Ptr(true)}
	named := &models.V1EksMachinePoolConfig{Name: "cp-pool"}
	worker := &models.V1EksMachinePoolConfig{Name: "worker-pool", IsControlPlane: types.Ptr(false)}

	assert.Same(t, cp, eksControlPlanePool([]*models.V1EksMachinePoolConfig{worker, nil, cp}))
	assert.Same(t, cp, eksControlPlanePool([]*models.V1EksMachinePoolConfig{named, cp}), "isControlPlane wins over the name")
	assert.Same(t, named, eksControlPlanePool([]*models.V1EksMachinePoolConfig{worker, named}))
	assert.Nil(t, eksControlPlanePool([]*models.V1EksMachinePoolConfig{worker}))
	assert.Nil(t, eksControlPlanePool(nil))
}

// eksAzSubnetsDiff diffs a static EKS cluster whose state holds the control-plane subnets against a configuration
// whose cloud_config.az_subnets is configAzSubnets (cty.NilVal leaves the attribute out). The configuration is built
// the way Terraform's plan sends it: a cty value shimmed into a ResourceConfig and set as the prior state's RawConfig.
func eksAzSubnetsDiff(t *testing.T, configAzSubnets cty.Value) *terraform.InstanceDiff {
	t.Helper()
	res := resourceClusterEks()

	stateRD := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":             "tf-si-static-eks",
		"cloud_account_id": "account-uid",
		"cloud_config": []interface{}{map[string]interface{}{
			"region": "us-east-2",
			"vpc_id": "vpc-0b69e5c29603de04e",
			"az_subnets": map[string]interface{}{
				"us-east-2a": "subnet-0ba60dc55d1bd5756,subnet-05ad5487490bc874f",
				"us-east-2c": "subnet-033a12c33a54af7a6,subnet-0967d4215d2ddf43f",
			},
		}},
		"machine_pool": []interface{}{map[string]interface{}{
			"name":          "worker-pool",
			"count":         1,
			"instance_type": "t3.xlarge",
			"disk_size_gb":  60,
		}},
	})
	stateRD.SetId("cluster-uid")
	state := stateRD.State()
	require.NotNil(t, state)

	cloudConfig := map[string]cty.Value{
		"region": cty.StringVal("us-east-2"),
		"vpc_id": cty.StringVal("vpc-0b69e5c29603de04e"),
	}
	if configAzSubnets != cty.NilVal {
		cloudConfig["az_subnets"] = configAzSubnets
	}
	block := res.CoreConfigSchema()
	configVal, err := block.CoerceValue(cty.ObjectVal(map[string]cty.Value{
		"name":             cty.StringVal("tf-si-static-eks"),
		"cloud_account_id": cty.StringVal("account-uid"),
		"cloud_config":     cty.ListVal([]cty.Value{cty.ObjectVal(cloudConfig)}),
		"machine_pool": cty.SetVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
			"name":          cty.StringVal("worker-pool"),
			"count":         cty.NumberIntVal(1),
			"instance_type": cty.StringVal("t3.xlarge"),
			"disk_size_gb":  cty.NumberIntVal(60),
		})}),
	}))
	require.NoError(t, err)
	// PlanResourceChange sets the prior state's RawConfig to the configuration before diffing.
	state.RawConfig = configVal

	diff, err := res.Diff(context.Background(), state, terraform.NewResourceConfigShimmed(configVal, block), nil)
	require.NoError(t, err)
	return diff
}

func eksDiffHasAzSubnetsChange(diff *terraform.InstanceDiff) bool {
	if diff == nil {
		return false
	}
	for k, attr := range diff.Attributes {
		if strings.HasPrefix(k, "cloud_config.0.az_subnets") && attr != nil && attr.Old != attr.New {
			return true
		}
	}
	return false
}

func TestEksAzSubnetsDiff(t *testing.T) {
	t.Run("config left empty after an older import does not force replacement", func(t *testing.T) {
		diff := eksAzSubnetsDiff(t, cty.MapValEmpty(cty.String))
		assert.False(t, eksDiffHasAzSubnetsChange(diff))
		assert.False(t, diff != nil && diff.RequiresNew(), "must not replace the cluster")
	})
	t.Run("config omitting az_subnets does not force replacement", func(t *testing.T) {
		diff := eksAzSubnetsDiff(t, cty.NilVal)
		assert.False(t, eksDiffHasAzSubnetsChange(diff))
		assert.False(t, diff != nil && diff.RequiresNew(), "must not replace the cluster")
	})
	t.Run("config matching Palette has no diff", func(t *testing.T) {
		diff := eksAzSubnetsDiff(t, cty.MapVal(map[string]cty.Value{
			"us-east-2a": cty.StringVal("subnet-0ba60dc55d1bd5756,subnet-05ad5487490bc874f"),
			"us-east-2c": cty.StringVal("subnet-033a12c33a54af7a6,subnet-0967d4215d2ddf43f"),
		}))
		assert.False(t, eksDiffHasAzSubnetsChange(diff))
	})
	t.Run("a real subnet change still forces replacement", func(t *testing.T) {
		diff := eksAzSubnetsDiff(t, cty.MapVal(map[string]cty.Value{
			"us-east-2a": cty.StringVal("subnet-0ba60dc55d1bd5756,subnet-05ad5487490bc874f"),
			"us-east-2b": cty.StringVal("subnet-0aaaaaaaaaaaaaaaa"),
		}))
		require.NotNil(t, diff)
		assert.True(t, eksDiffHasAzSubnetsChange(diff))
		assert.True(t, diff.RequiresNew())
	})
}
