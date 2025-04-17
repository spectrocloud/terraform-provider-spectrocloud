package spectrocloud

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func getBaseResourceData() *schema.ResourceData {
	rd := resourceApplicationProfile().TestResourceData()
	rd.Set("name", "Profile-unit-test")
	rd.Set("version", "1.0.0")
	rd.Set("context", "project")
	rd.Set("description", "Test profile creation for unit test")
	rd.Set("cloud", "all")
	return rd

}

func TestToApplicationProfileBasic(t *testing.T) {
	appProfileBasicRd := getBaseResourceData()
	profileEntity := toApplicationProfileBasic(appProfileBasicRd)
	assert.Equal(t, appProfileBasicRd.Get("description"), profileEntity.Metadata.Annotations["description"])
	assert.Equal(t, appProfileBasicRd.Get("name"), profileEntity.Metadata.Name)
	assert.Equal(t, appProfileBasicRd.Get("version"), profileEntity.Spec.Version)
}

func TestToAppTiers(t *testing.T) {
	appProfileBasicRd := getBaseResourceData()
	profileEntity := toApplicationProfileBasic(appProfileBasicRd)
	if profileEntity.Spec.Template.AppTiers == nil {
		assert.Fail(t, "After convert toApplicationProfileBasic tier is set to nil")
	}
}

func TestToApplicationProfilePackCreateWithPack(t *testing.T) {

	packOne := make(map[string]interface{})
	prop := make(map[string]interface{})
	prop["dbRootPassword"] = "test123!wewe!"
	prop["dbVolumeSize"] = "20"
	prop["dbVersion"] = "5.7"
	packOne["type"] = "operator-instance"
	packOne["name"] = "mysql-3"
	packOne["source_app_tier"] = "636c0714c196e565df7a7b37"
	packOne["properties"] = prop
	packOne["values"] = ""
	packOne["manifest"] = make([]interface{}, 0)

	profileEntity, err := toApplicationProfilePackCreate(packOne)
	if err != nil {
		assert.Fail(t, "toApplicationProfilePackCreate - Not able to convert the interface")
	}

	assert.Equal(t, packOne["name"].(string), *profileEntity.Name)
	assert.Equal(t, packOne["source_app_tier"].(string), profileEntity.SourceAppTierUID)
	assert.Equal(t, packOne["values"].(string), profileEntity.Values)
	assert.Equal(t, packOne["type"], string(*profileEntity.Type))
	for _, v := range profileEntity.Properties {
		assert.Equal(t, prop[v.Name], v.Value)
	}

}

func TestToApplicationProfilePackCreateWithManifest(t *testing.T) {

	packOne := make(map[string]interface{})
	packOne["type"] = "operator-instance"
	packOne["name"] = "nginx"
	packOne["source_app_tier"] = "636c0714c196e565df7a7b37"
	packOne["values"] = ""
	manifest := make([]interface{}, 0)
	manifest = append(manifest, map[string]interface{}{
		"content": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: nginx-deployment\n  labels:\n    app: nginx\nspec:\n  replicas: 3\n  selector:\n    matchLabels:\n      app: nginx\n  template:\n    metadata:\n      labels:\n        app: nginx\n    spec:\n      containers:\n      - name: nginx\n        image: nginx:1.14.2\n        ports:\n        - containerPort: 80",
		"name":    "nginx",
	})

	packOne["manifest"] = manifest

	profileEntity, err := toApplicationProfilePackCreate(packOne)
	if err != nil {
		assert.Fail(t, "toApplicationProfilePackCreate - Not able to convert the interface")
	}

	assert.Equal(t, packOne["name"].(string), *profileEntity.Name)
	assert.Equal(t, packOne["source_app_tier"].(string), profileEntity.SourceAppTierUID)
	assert.Equal(t, packOne["values"].(string), profileEntity.Values)
	assert.Equal(t, packOne["type"], string(*profileEntity.Type))
	for _, v := range profileEntity.Manifests {
		assert.Equal(t, v.Content, strings.TrimSpace(manifest[0].(map[string]interface{})["content"].(string)))
		assert.Equal(t, v.Name, manifest[0].(map[string]interface{})["name"].(string))
	}

}

func TestToApplicationProfilePackCreateWithPackManifest(t *testing.T) {

	packOne := make(map[string]interface{})
	prop := make(map[string]interface{})
	prop["dbRootPassword"] = "test123!wewe!"
	prop["dbVolumeSize"] = "20"
	prop["dbVersion"] = "5.7"
	packOne["type"] = "operator-instance"
	packOne["name"] = "mysql-3"
	packOne["source_app_tier"] = "636c0714c196e565df7a7b37"
	packOne["properties"] = prop
	packOne["values"] = ""
	manifest := make([]interface{}, 0)
	manifest = append(manifest, map[string]interface{}{
		"content": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: nginx-deployment\n  labels:\n    app: nginx\nspec:\n  replicas: 3\n  selector:\n    matchLabels:\n      app: nginx\n  template:\n    metadata:\n      labels:\n        app: nginx\n    spec:\n      containers:\n      - name: nginx\n        image: nginx:1.14.2\n        ports:\n        - containerPort: 80",
		"name":    "nginx",
	})
	packOne["manifest"] = manifest

	profileEntity, err := toApplicationProfilePackCreate(packOne)
	if err != nil {
		assert.Fail(t, "toApplicationProfilePackCreate - Not able to convert the interface")
	}

	assert.Equal(t, packOne["name"].(string), *profileEntity.Name)
	assert.Equal(t, packOne["source_app_tier"].(string), profileEntity.SourceAppTierUID)
	assert.Equal(t, packOne["values"].(string), profileEntity.Values)
	assert.Equal(t, packOne["type"], string(*profileEntity.Type))
	for _, v := range profileEntity.Properties {
		assert.Equal(t, prop[v.Name], v.Value)
	}
	for _, v := range profileEntity.Manifests {
		assert.Equal(t, v.Content, strings.TrimSpace(manifest[0].(map[string]interface{})["content"].(string)))
		assert.Equal(t, v.Name, manifest[0].(map[string]interface{})["name"].(string))
	}

}

func TestToTags(t *testing.T) {
	tagRD := getBaseResourceData()
	tagMap := []string{"owner:sivaa", "unittest"}
	err := tagRD.Set("tags", tagMap)
	if err != nil {
		assert.Fail(t, "Error setting tags.")
	}
	tags := toTags(tagRD)
	assert.Equal(t, strings.Split(tagMap[0], ":")[1], tags["owner"])
	assert.Equal(t, "spectro__tag", tags["unittest"])

}

func TestFlattenTags(t *testing.T) {
	tagMap := make(map[string]string)
	tagMap["unittest"] = "spectro__tag"
	tagMap["owner"] = "siva"
	tags := flattenTags(tagMap)

	// Check that the tags slice contains the expected tags, regardless of order
	assert.Contains(t, tags, "unittest")
	assert.Contains(t, tags, "owner:"+tagMap["owner"])
}

func TestFlattenTagsEmpty(t *testing.T) {
	tagMap := make(map[string]string)
	tags := flattenTags(tagMap)
	// should be nil if empty
	assert.Equal(t, []interface{}(nil), tags)
}

func TestToApplicationProfilePatch(t *testing.T) {
	profilePatchRD := getBaseResourceData()
	tagMap := []string{"owner:sivaa", "unittest"}
	profilePatchRD.Set("tags", tagMap)
	profileMetaEntity, err := toApplicationProfilePatch(profilePatchRD)
	if err != nil {
		assert.Fail(t, "toApplicationProfilePatch - Not able to convert the resource data")
	}
	assert.Equal(t, profilePatchRD.Get("description"), profileMetaEntity.Metadata.Annotations["description"])
	assert.Equal(t, strings.Split(tagMap[0], ":")[1], profileMetaEntity.Metadata.Labels["owner"])
	assert.Equal(t, "spectro__tag", profileMetaEntity.Metadata.Labels["unittest"])
}

func TestToApplicationProfilePackUpdate(t *testing.T) {
	packOne := make(map[string]interface{})
	prop := make(map[string]interface{})
	prop["dbRootPassword"] = "test123!wewe!"
	prop["dbVolumeSize"] = "20"
	prop["dbVersion"] = "5.7"
	packOne["type"] = "operator-instance"
	packOne["name"] = "mysql-3"
	packOne["tag"] = "10.5"
	packOne["properties"] = prop
	packOne["values"] = ""
	manifest := make([]interface{}, 0)
	m := map[string]interface{}{
		"content": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: nginx-deployment\n  labels:\n    app: nginx\nspec:\n  replicas: 3\n  selector:\n    matchLabels:\n      app: nginx\n  template:\n    metadata:\n      labels:\n        app: nginx\n    spec:\n      containers:\n      - name: nginx\n        image: nginx:1.14.2\n        ports:\n        - containerPort: 80",
		"name":    "nginx",
	}
	manifest = append(manifest, m)
	packOne["manifest"] = manifest

	profileEntity := toApplicationProfilePackUpdate(packOne)
	assert.Equal(t, packOne["name"].(string), profileEntity.Name)
	assert.Equal(t, packOne["values"].(string), profileEntity.Values)

	for _, v := range profileEntity.Properties {
		assert.Equal(t, prop[v.Name], v.Value)
	}
	for _, v := range profileEntity.Manifests {
		assert.Equal(t, v.Content, strings.TrimSpace(m["content"].(string)))
		assert.Equal(t, *v.Name, m["name"].(string))
	}
}

func TestGetValueInProperties(t *testing.T) {
	prop := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	result := getValueInProperties(prop, "key2")
	assert.Equal(t, prop["key2"], result)
	result = getValueInProperties(prop, "key3")
	assert.Equal(t, "", result)
}

func TestToPropertiesTier(t *testing.T) {
	props := map[string]interface{}{
		"properties": map[string]interface{}{
			"aa": "value1",
			"bb": "value2",
		},
	}
	p := toPropertiesTier(props)

	assertProperties := func(name, value string) bool {
		for _, prop := range p {
			if prop.Name == name && prop.Value == value {
				return true
			}
		}
		return false
	}
	assert.True(t, assertProperties("aa", "value1"))
	assert.True(t, assertProperties("bb", "value2"))

	// assert there are no any other properties
	assert.Equal(t, 2, len(p))
}

func TestToApplicationProfileCreate(t *testing.T) {
	d := getBaseResourceData()
	var p []map[string]interface{}
	p = append(p, map[string]interface{}{
		"type":            "operator-instance",
		"source_app_tier": "testSUID",
		"registry_uid":    "test_reg_uid",
		"uid":             "test_pack_uid",
		"name":            "mysql",
		"properties": map[string]interface{}{
			"dbname": "testDB",
		},
	})
	_ = d.Set("pack", p)
	cp, _ := toApplicationProfileCreate(d)
	assert.Equal(t, p[0]["type"], string(*cp.Spec.Template.AppTiers[0].Type))
	assert.Equal(t, p[0]["source_app_tier"], cp.Spec.Template.AppTiers[0].SourceAppTierUID)
	assert.Equal(t, p[0]["registry_uid"], cp.Spec.Template.AppTiers[0].RegistryUID)
	assert.Equal(t, "dbname", string(cp.Spec.Template.AppTiers[0].Properties[0].Name))
	assert.Equal(t, "testDB", string(cp.Spec.Template.AppTiers[0].Properties[0].Value))
}

func TestToApplicationTiersUpdate(t *testing.T) {
	d := getBaseResourceData()
	d.SetId("test-app-profile-id")
	var p []map[string]interface{}
	p = append(p, map[string]interface{}{
		"type":            "operator-instance",
		"source_app_tier": "testSUID",
		"registry_uid":    "test_reg_uid",
		"uid":             "test_pack_uid",
		"name":            "mysql",
		"properties": map[string]interface{}{
			"dbname": "testDB",
		},
	})
	_ = d.Set("pack", p)

	_, _, _, err := toApplicationTiersUpdate(d, getV1ClientWithResourceContext(unitTestMockAPIClient, ""))
	assert.Empty(t, err)
}

func TestResourceApplicationProfileCreate(t *testing.T) {
	d := getBaseResourceData()
	var ctx context.Context
	_ = d.Set("context", "project")
	s := resourceApplicationProfileCreate(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, false, s.HasError())

}

func TestResourceApplicationProfileUpdate(t *testing.T) {
	d := getBaseResourceData()
	var ctx context.Context
	_ = d.Set("context", "project")
	s := resourceApplicationProfileUpdate(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, false, s.HasError())

}

func TestResourceApplicationProfileDelete(t *testing.T) {
	d := getBaseResourceData()
	var ctx context.Context
	d.SetId("test-app-profile-id")
	r := resourceApplicationProfileDelete(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, false, r.HasError())
}
