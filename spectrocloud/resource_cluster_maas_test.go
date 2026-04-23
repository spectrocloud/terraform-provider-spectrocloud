package spectrocloud

import (
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func prepareClusterMaasTestResourceData(cloudConfigExtras map[string]interface{}) *schema.ResourceData {
	d := resourceClusterMaas().TestResourceData()
	d.SetId("")
	d.Set("name", "maas-test-cluster")
	d.Set("cloud_account_id", "maas-test-account-id")

	con := map[string]interface{}{
		"domain": "maas.test.local",
	}
	for k, v := range cloudConfigExtras {
		con[k] = v
	}
	d.Set("cloud_config", []map[string]interface{}{con})
	return d
}

func sortedCopy(in []string) []string {
	out := make([]string, len(in))
	copy(out, in)
	sort.Strings(out)
	return out
}

func TestGetMaasSSHKeys_FromSet(t *testing.T) {
	assert := assert.New(t)
	set := schema.NewSet(schema.HashString, []interface{}{"ssh-rsa AAAA1 ", " ssh-rsa AAAA2"})
	cloudConfig := map[string]interface{}{
		"ssh_keys": set,
	}
	keys := getMaasSSHKeys(cloudConfig)
	assert.Equal(2, len(keys))
	sorted := sortedCopy(keys)
	assert.Equal([]string{"ssh-rsa AAAA1", "ssh-rsa AAAA2"}, sorted)
}

func TestGetMaasSSHKeys_FromSingleKey(t *testing.T) {
	assert := assert.New(t)
	cloudConfig := map[string]interface{}{
		"ssh_key": "  ssh-rsa AAAAsingle  ",
	}
	keys := getMaasSSHKeys(cloudConfig)
	assert.Equal([]string{"ssh-rsa AAAAsingle"}, keys)
}

func TestGetMaasSSHKeys_Neither(t *testing.T) {
	assert := assert.New(t)
	cloudConfig := map[string]interface{}{}
	keys := getMaasSSHKeys(cloudConfig)
	assert.Nil(keys, "neither ssh_key nor ssh_keys set should return nil (backward compat)")
}

func TestToMaasCloudConfigUpdate_WithSSHKeys(t *testing.T) {
	assert := assert.New(t)
	set := schema.NewSet(schema.HashString, []interface{}{"ssh-rsa AAAA1", "ssh-rsa AAAA2"})
	cloudConfig := map[string]interface{}{
		"domain":        "maas.test.local",
		"enable_lxd_vm": false,
		"ssh_keys":      set,
	}
	entity := toMaasCloudConfigUpdate(cloudConfig)
	assert.NotNil(entity)
	assert.NotNil(entity.ClusterConfig)
	assert.Equal(2, len(entity.ClusterConfig.SSHKeys))
	sorted := sortedCopy(entity.ClusterConfig.SSHKeys)
	assert.Equal([]string{"ssh-rsa AAAA1", "ssh-rsa AAAA2"}, sorted)
}

func TestToMaasCloudConfigUpdate_WithSingleSSHKey(t *testing.T) {
	assert := assert.New(t)
	cloudConfig := map[string]interface{}{
		"domain":        "maas.test.local",
		"enable_lxd_vm": false,
		"ssh_key":       "ssh-rsa AAAAsingle",
	}
	entity := toMaasCloudConfigUpdate(cloudConfig)
	assert.NotNil(entity)
	assert.NotNil(entity.ClusterConfig)
	assert.Equal([]string{"ssh-rsa AAAAsingle"}, entity.ClusterConfig.SSHKeys)
}

func TestToMaasCloudConfigUpdate_NoSSHKeys(t *testing.T) {
	assert := assert.New(t)
	cloudConfig := map[string]interface{}{
		"domain":        "maas.test.local",
		"enable_lxd_vm": false,
	}
	entity := toMaasCloudConfigUpdate(cloudConfig)
	assert.NotNil(entity)
	assert.NotNil(entity.ClusterConfig)
	assert.Nil(entity.ClusterConfig.SSHKeys, "no SSH keys should serialise as nil (backward compat)")
}

func TestFlattenClusterConfigsMaas_SSHKeys_Preferred(t *testing.T) {
	assert := assert.New(t)
	keys := []string{"ssh-rsa AAAA1", "ssh-rsa AAAA2"}
	inputCloudConfig := &models.V1MaasCloudConfig{
		Spec: &models.V1MaasCloudConfigSpec{
			ClusterConfig: &models.V1MaasClusterConfig{
				Domain:  strPtr("maas.test.local"),
				SSHKeys: keys,
			},
		},
	}
	set := schema.NewSet(schema.HashString, []interface{}{"placeholder"})
	d := prepareClusterMaasTestResourceData(map[string]interface{}{
		"ssh_keys": set,
	})
	flat := flattenClusterConfigsMaas(d, inputCloudConfig)
	assert.Equal(1, len(flat))
	m := flat[0].(map[string]interface{})
	assert.Equal("maas.test.local", m["domain"])
	got, ok := m["ssh_keys"].([]string)
	assert.True(ok, "ssh_keys should be set via the preferred branch")
	assert.Equal(keys, got)
	_, hasKey := m["ssh_key"]
	assert.False(hasKey, "ssh_key should not be populated when user used ssh_keys")
}

func TestFlattenClusterConfigsMaas_SSHKey_Deprecated(t *testing.T) {
	assert := assert.New(t)
	inputCloudConfig := &models.V1MaasCloudConfig{
		Spec: &models.V1MaasCloudConfigSpec{
			ClusterConfig: &models.V1MaasClusterConfig{
				Domain:  strPtr("maas.test.local"),
				SSHKeys: []string{"  ssh-rsa AAAAsingle  "},
			},
		},
	}
	d := prepareClusterMaasTestResourceData(map[string]interface{}{
		"ssh_key": "ssh-rsa AAAAsingle",
	})
	flat := flattenClusterConfigsMaas(d, inputCloudConfig)
	assert.Equal(1, len(flat))
	m := flat[0].(map[string]interface{})
	assert.Equal("ssh-rsa AAAAsingle", m["ssh_key"], "ssh_key should be trimmed and populated")
}

func TestFlattenClusterConfigsMaas_Import(t *testing.T) {
	assert := assert.New(t)
	keys := []string{"ssh-rsa AAAAimport1", "ssh-rsa AAAAimport2"}
	inputCloudConfig := &models.V1MaasCloudConfig{
		Spec: &models.V1MaasCloudConfigSpec{
			ClusterConfig: &models.V1MaasClusterConfig{
				Domain:  strPtr("maas.test.local"),
				SSHKeys: keys,
			},
		},
	}
	// bare ResourceData — no cloud_config preset to simulate import
	d := resourceClusterMaas().TestResourceData()
	flat := flattenClusterConfigsMaas(d, inputCloudConfig)
	assert.Equal(1, len(flat))
	m := flat[0].(map[string]interface{})
	got, ok := m["ssh_keys"].([]string)
	assert.True(ok, "ssh_keys should default on import when neither GetOk branch trips")
	if !reflect.DeepEqual(got, keys) {
		t.Errorf("expected %#v, got %#v", keys, got)
	}
}

func TestFlattenClusterConfigsMaas_Nil(t *testing.T) {
	d := prepareClusterMaasTestResourceData(nil)
	flat := flattenClusterConfigsMaas(d, nil)
	if flat == nil {
		t.Errorf("flattenClusterConfigsMaas should not return nil for a nil input config")
	}
	if len(flat) != 0 {
		t.Errorf("flattenClusterConfigsMaas should return empty slice for nil config, got len=%d", len(flat))
	}
}

func TestFlattenClusterConfigsMaas_NoSSHKeys(t *testing.T) {
	assert := assert.New(t)
	inputCloudConfig := &models.V1MaasCloudConfig{
		Spec: &models.V1MaasCloudConfigSpec{
			ClusterConfig: &models.V1MaasClusterConfig{
				Domain: strPtr("maas.test.local"),
			},
		},
	}
	d := prepareClusterMaasTestResourceData(nil)
	flat := flattenClusterConfigsMaas(d, inputCloudConfig)
	assert.Equal(1, len(flat))
	m := flat[0].(map[string]interface{})
	_, hasKey := m["ssh_key"]
	_, hasKeys := m["ssh_keys"]
	assert.False(hasKey, "ssh_key should be absent when server has no SSH keys")
	assert.False(hasKeys, "ssh_keys should be absent when server has no SSH keys")
}

func strPtr(s string) *string {
	return &s
}
