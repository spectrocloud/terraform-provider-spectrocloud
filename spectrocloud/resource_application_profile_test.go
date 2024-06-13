package spectrocloud

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-api-go/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
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
	assert.Equal(t, packOne["type"], string(profileEntity.Type))
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
	assert.Equal(t, packOne["type"], string(profileEntity.Type))
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
	assert.Equal(t, packOne["type"], string(profileEntity.Type))
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

func TestGetAppTiersContent(t *testing.T) {
	appUid := "test-app-tier-id"
	d := getBaseResourceData()
	d.SetId(appUid)
	m := &client.V1Client{
		GetApplicationProfileTiersFn: func(appProfileID string) ([]*models.V1AppTier, error) {
			var appTierSet []*models.V1AppTier
			tier := &models.V1AppTier{
				Metadata: &models.V1ObjectMeta{
					UID:  appUid,
					Name: "mysql",
				},
				Spec: &models.V1AppTierSpec{
					Type:             "operator-instance",
					SourceAppTierUID: "test-source-uid",
					Version:          "5.25",
					RegistryUID:      "test-registry-id",
					InstallOrder:     10,
				},
			}
			appTierSet = append(appTierSet, tier)
			return appTierSet, nil
		},
	}
	appTiers, _, _ := getAppTiersContent(m, d)
	assert.Equal(t, appUid, appTiers[0].Metadata.UID)
	assert.Equal(t, "mysql", appTiers[0].Metadata.Name)
	assert.Equal(t, "test-source-uid", appTiers[0].Spec.SourceAppTierUID)
	assert.Equal(t, "5.25", appTiers[0].Spec.Version)
	assert.Equal(t, "test-registry-id", appTiers[0].Spec.RegistryUID)
	assert.Equal(t, 10, int(appTiers[0].Spec.InstallOrder))
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

func TestFlattenAppPacks(t *testing.T) {
	d := getBaseResourceData()
	ctx := context.Background()
	m := &client.V1Client{
		GetPackRegistryCommonByNameFn: func(regName string) (*models.V1RegistryMetadata, error) {
			reg := &models.V1RegistryMetadata{
				IsPrivate: false,
				Kind:      "pack",
				Name:      "Public Repo",
				Scope:     "project",
				UID:       "test-pub-registry-uid",
			}
			return reg, nil
		},
		GetApplicationProfileTierManifestContentFn: func(appProfileUID, tierUID, manifestUID string) (string, error) {
			return "test: \n content", nil
		},
	}

	var diagPack []*models.V1PackManifestEntity
	diagPack = append(diagPack, &models.V1PackManifestEntity{
		UID:         "test-pack-uid",
		Name:        types.Ptr("kafka"),
		RegistryUID: "test-pub-registry-uid",
		Type:        "manifest",
		Values:      "test values",
	})

	var tiers []*models.V1AppTierRef
	tiers = append(tiers, &models.V1AppTierRef{
		Type:    "manifest",
		UID:     "test-tier-uid",
		Name:    "kafka",
		Version: "5.1",
	})

	var tierDet []*models.V1AppTier
	var manifest []*models.V1ObjectReference
	manifest = append(manifest, &models.V1ObjectReference{
		Name:            "kafka-dep",
		UID:             "test-manifest-uid",
		APIVersion:      "apps/v1",
		Kind:            "Deployment",
		ResourceVersion: "v1",
	})

	var props []*models.V1AppTierProperty
	props = append(props, &models.V1AppTierProperty{
		Name:   "prop_key",
		Value:  "prop_value",
		Type:   "string",
		Format: "",
	})
	tierDet = append(tierDet, &models.V1AppTier{
		Metadata: &models.V1ObjectMeta{
			UID:  "test-uid",
			Name: "kafka",
		},
		Spec: &models.V1AppTierSpec{
			Type:             "manifest",
			SourceAppTierUID: "test-source-uid",
			Version:          "5.25",
			RegistryUID:      "test-registry-id",
			InstallOrder:     10,
			Manifests:        manifest,
			Properties:       props,
		},
	})

	re, _ := flattenAppPacks(m, diagPack, tiers, tierDet, d, ctx)
	assert.Equal(t, "test-uid", re[0].(map[string]interface{})["uid"])
	assert.Equal(t, "test-registry-id", re[0].(map[string]interface{})["registry_uid"])
	assert.Equal(t, "kafka", re[0].(map[string]interface{})["name"])
	assert.Equal(t, "test-source-uid", re[0].(map[string]interface{})["source_app_tier"])
	assert.Equal(t, "prop_value", re[0].(map[string]interface{})["properties"].(map[string]string)["prop_key"])
	assert.Equal(t, "kafka-dep", re[0].(map[string]interface{})["manifest"].([]interface{})[0].(map[string]interface{})["name"])
	assert.Equal(t, "test-manifest-uid", re[0].(map[string]interface{})["manifest"].([]interface{})[0].(map[string]interface{})["uid"])
	assert.Equal(t, "test: \n content", re[0].(map[string]interface{})["manifest"].([]interface{})[0].(map[string]interface{})["content"])
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
	d.Set("pack", p)
	cp, _ := toApplicationProfileCreate(d)
	assert.Equal(t, p[0]["type"], string(cp.Spec.Template.AppTiers[0].Type))
	assert.Equal(t, p[0]["source_app_tier"], cp.Spec.Template.AppTiers[0].SourceAppTierUID)
	assert.Equal(t, p[0]["registry_uid"], cp.Spec.Template.AppTiers[0].RegistryUID)
	assert.Equal(t, "dbname", string(cp.Spec.Template.AppTiers[0].Properties[0].Name))
	assert.Equal(t, "testDB", string(cp.Spec.Template.AppTiers[0].Properties[0].Value))
}

func TestToApplicationTiersUpdate(t *testing.T) {
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
	d.Set("pack", p)
	m := &client.V1Client{
		GetApplicationProfileTiersFn: func(appProfileID string) ([]*models.V1AppTier, error) {
			var appTierSet []*models.V1AppTier
			tier := &models.V1AppTier{
				Metadata: &models.V1ObjectMeta{
					UID:  "test-uid",
					Name: "mysql",
				},
				Spec: &models.V1AppTierSpec{
					Type:             "operator-instance",
					SourceAppTierUID: "test-source-uid",
					Version:          "5.25",
					RegistryUID:      "test-registry-id",
					InstallOrder:     10,
				},
			}
			appTierSet = append(appTierSet, tier)
			return appTierSet, nil
		},
	}
	_, ut, _, _ := toApplicationTiersUpdate(d, m)
	assert.Equal(t, "mysql", ut["test-uid"].Name)
	assert.Equal(t, "dbname", string(ut["test-uid"].Properties[0].Name))
	assert.Equal(t, "testDB", string(ut["test-uid"].Properties[0].Value))
}

func TestResourceApplicationProfileCreate(t *testing.T) {
	d := getBaseResourceData()
	ctx := context.Background()
	m := &client.V1Client{
		CreateApplicationProfileFn: func(entity *models.V1AppProfileEntity, s string) (string, error) {
			return "test_application_profile_uid", nil
		},
		GetApplicationProfileTiersFn: func(appProfileID string) ([]*models.V1AppTier, error) {
			var appTierSet []*models.V1AppTier
			tier := &models.V1AppTier{
				Metadata: &models.V1ObjectMeta{
					UID:  "appUid",
					Name: "mysql",
				},
				Spec: &models.V1AppTierSpec{
					Type:             "operator-instance",
					SourceAppTierUID: "test-source-uid",
					Version:          "5.25",
					RegistryUID:      "test-registry-id",
					InstallOrder:     10,
				},
			}
			appTierSet = append(appTierSet, tier)
			return appTierSet, nil
		},
		GetApplicationProfileFn: func(uid string) (*models.V1AppProfile, error) {
			var tiers []*models.V1AppTierRef
			tiers = append(tiers, &models.V1AppTierRef{
				Type:    "manifest",
				UID:     "test-tier-uid",
				Name:    "kafka",
				Version: "5.1",
			})
			ap := &models.V1AppProfile{
				Metadata: &models.V1ObjectMeta{
					UID:  "test_application_profile_uid",
					Name: "test_application_profile",
					Labels: map[string]string{
						"owner": "siva",
					},
				},
				Spec: &models.V1AppProfileSpec{
					Version: "5.4",
					Template: &models.V1AppProfileTemplate{
						AppTiers: tiers,
					},
				},
			}
			return ap, nil
		},
	}
	s := resourceApplicationProfileCreate(ctx, d, m)
	assert.Equal(t, false, s.HasError())

}

func TestResourceApplicationProfileDelete(t *testing.T) {
	d := getBaseResourceData()
	ctx := context.Background()
	m := &client.V1Client{
		DeleteApplicationProfileFn: func(s string) error {
			return nil
		},
	}
	r := resourceApplicationProfileDelete(ctx, d, m)
	assert.Equal(t, false, r.HasError())
}
