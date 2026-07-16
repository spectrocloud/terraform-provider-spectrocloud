package spectrocloud

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
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

// TestFlattenAppPacksPLT2297 pins down the fix for PLT-2297:
// after adding a second pack with an install_order and tag, `terraform plan`
// used to show drift on `tag`, `install_order`, and `uid` for every pack on
// every subsequent apply. That was because flattenAppPacks never populated
// these fields into state. The regression this guards against is a re-read
// returning state that doesn't match the config that was just applied.
func TestFlattenAppPacksPLT2297(t *testing.T) {
	d := getBaseResourceData()
	d.SetId("test-app-profile-id")

	// Two packs, same shape as the ticket's repro (manifest packs, both
	// with an install_order and a tag).
	packs := []map[string]interface{}{
		{
			"type":          "manifest",
			"name":          "tier-manifest",
			"uid":           "6a548778bb5b6f34a1c5d0cf", // pack catalog UID from config
			"tag":           "1.0.0",
			"install_order": 0,
			"registry_uid":  "reg-uid-1",
			"values":        "",
			"manifest":      []interface{}{},
		},
		{
			"type":          "manifest",
			"name":          "tier-backend",
			"uid":           "6a54878ba1e393c473cf8277",
			"tag":           "1.0.0",
			"install_order": 1,
			"registry_uid":  "reg-uid-2",
			"values":        "",
			"manifest":      []interface{}{},
		},
	}
	if err := d.Set("pack", packs); err != nil {
		t.Fatalf("failed to seed pack config: %v", err)
	}

	// Fabricate what the API would return for these two tiers. tier-manifest
	// carries the values the user set (tag 1.0.0, installOrder 0), and its
	// Metadata.UID is the *tier* UID (distinct from the user's pack UID).
	manifestType := models.V1AppTierType("manifest")
	tierDet := []*models.V1AppTier{
		{
			Metadata: &models.V1ObjectMeta{
				Name: "tier-manifest",
				UID:  "6a548d78f74be122b0963090", // tier metadata UID from the ticket
			},
			Spec: &models.V1AppTierSpec{
				Type:         &manifestType,
				Version:      "1.0.0",
				InstallOrder: 0,
				RegistryUID:  "reg-uid-1",
			},
		},
		{
			Metadata: &models.V1ObjectMeta{
				Name: "tier-backend",
				UID:  "6a548ebabb5b6f39a6257717",
			},
			Spec: &models.V1AppTierSpec{
				Type:         &manifestType,
				Version:      "1.0.0",
				InstallOrder: 1,
				RegistryUID:  "reg-uid-2",
			},
		},
	}
	// tiers slice only sets the length of the returned ps; content is unused.
	tiers := make([]*models.V1AppTierRef, len(tierDet))

	// diagPacks is a passthrough; empty is fine because we don't rely on it
	// in this scenario (no registry_name lookups, no manifest content fetch).
	// Client is not exercised because every tier already has RegistryUID and
	// no tier has manifests.
	out, err := flattenAppPacks(nil, nil, tiers, tierDet, d, context.Background())
	if err != nil {
		t.Fatalf("flattenAppPacks returned error: %v", err)
	}

	if got, want := len(out), 2; got != want {
		t.Fatalf("expected %d flattened packs, got %d", want, got)
	}

	first := out[0].(map[string]interface{})
	assert.Equal(t, "tier-manifest", first["name"], "name flattened")
	assert.Equal(t, "1.0.0", first["tag"], "PLT-2297: tag must be populated from tier.Spec.Version")
	assert.Equal(t, 0, first["install_order"], "PLT-2297: install_order must be populated from tier.Spec.InstallOrder")
	assert.Equal(t, "6a548778bb5b6f34a1c5d0cf", first["uid"],
		"PLT-2297: uid must mirror the user-supplied pack catalog UID, not tier.Metadata.UID")

	second := out[1].(map[string]interface{})
	assert.Equal(t, "tier-backend", second["name"], "name flattened")
	assert.Equal(t, "1.0.0", second["tag"], "PLT-2297: tag must be populated on the second pack too")
	assert.Equal(t, 1, second["install_order"],
		"PLT-2297: install_order=1 on the second pack must round-trip through state")
	assert.Equal(t, "6a54878ba1e393c473cf8277", second["uid"],
		"PLT-2297: uid must mirror the second pack's user-supplied UID")
}

// TestFlattenAppPacksPLT2297_UnorderedTiers exercises the same PLT-2297 fix
// under a subtler scenario: the API returns tiers in a different order than
// the user's HCL. The old positional-by-index uid preservation would fail
// here; the fix matches by tier name so state stays consistent.
func TestFlattenAppPacksPLT2297_UnorderedTiers(t *testing.T) {
	d := getBaseResourceData()
	d.SetId("test-app-profile-id")

	packs := []map[string]interface{}{
		{
			"type":          "manifest",
			"name":          "tier-manifest",
			"uid":           "user-uid-manifest",
			"tag":           "1.0.0",
			"install_order": 0,
			"registry_uid":  "reg-uid-1",
			"values":        "",
			"manifest":      []interface{}{},
		},
		{
			"type":          "manifest",
			"name":          "tier-backend",
			"uid":           "user-uid-backend",
			"tag":           "2.0.0",
			"install_order": 1,
			"registry_uid":  "reg-uid-2",
			"values":        "",
			"manifest":      []interface{}{},
		},
	}
	if err := d.Set("pack", packs); err != nil {
		t.Fatalf("failed to seed pack config: %v", err)
	}

	manifestType := models.V1AppTierType("manifest")
	// Note the API returns backend BEFORE manifest (different order than HCL).
	tierDet := []*models.V1AppTier{
		{
			Metadata: &models.V1ObjectMeta{Name: "tier-backend", UID: "tier-meta-backend"},
			Spec: &models.V1AppTierSpec{
				Type: &manifestType, Version: "2.0.0", InstallOrder: 1, RegistryUID: "reg-uid-2",
			},
		},
		{
			Metadata: &models.V1ObjectMeta{Name: "tier-manifest", UID: "tier-meta-manifest"},
			Spec: &models.V1AppTierSpec{
				Type: &manifestType, Version: "1.0.0", InstallOrder: 0, RegistryUID: "reg-uid-1",
			},
		},
	}
	tiers := make([]*models.V1AppTierRef, len(tierDet))

	out, err := flattenAppPacks(nil, nil, tiers, tierDet, d, context.Background())
	if err != nil {
		t.Fatalf("flattenAppPacks returned error: %v", err)
	}

	// Each flattened pack should carry the uid the user set in HCL for THAT
	// pack name, regardless of the order tiers came back from the API.
	byName := map[string]map[string]interface{}{}
	for _, item := range out {
		m := item.(map[string]interface{})
		byName[m["name"].(string)] = m
	}
	assert.Equal(t, "user-uid-manifest", byName["tier-manifest"]["uid"],
		"PLT-2297: uid must be matched by pack name, not slice index")
	assert.Equal(t, "user-uid-backend", byName["tier-backend"]["uid"],
		"PLT-2297: uid must be matched by pack name, not slice index")
	assert.Equal(t, 0, byName["tier-manifest"]["install_order"])
	assert.Equal(t, 1, byName["tier-backend"]["install_order"])
	assert.Equal(t, "1.0.0", byName["tier-manifest"]["tag"])
	assert.Equal(t, "2.0.0", byName["tier-backend"]["tag"])
}

// TestToApplicationProfilePackCreate_InstallOrder_PLT2297 pins the second half
// of the PLT-2297 fix: without setting InstallOrder on the create entity, new
// tiers land in Palette with installOrder=0 regardless of config, and the next
// read plans drift back to the configured value. This was the source of the
// residual `install_order = 0 -> 1` diff on tier-backend even after the read
// path was fixed.
func TestToApplicationProfilePackCreate_InstallOrder_PLT2297(t *testing.T) {
	pack := map[string]interface{}{
		"type":            "manifest",
		"name":            "tier-backend",
		"tag":             "1.0.0",
		"install_order":   1,
		"source_app_tier": "",
		"values":          "",
		"manifest":        []interface{}{},
	}
	entity, err := toApplicationProfilePackCreate(pack)
	if err != nil {
		t.Fatalf("toApplicationProfilePackCreate returned error: %v", err)
	}
	assert.EqualValues(t, 1, entity.InstallOrder,
		"PLT-2297: install_order must be sent to the API on create")
}

func TestToApplicationProfilePackCreate_InstallOrder_DefaultsToZero(t *testing.T) {
	pack := map[string]interface{}{
		"type":            "manifest",
		"name":            "tier-manifest",
		"tag":             "1.0.0",
		"source_app_tier": "",
		"values":          "",
		"manifest":        []interface{}{},
		// install_order omitted → should default to 0, not panic.
	}
	entity, err := toApplicationProfilePackCreate(pack)
	if err != nil {
		t.Fatalf("toApplicationProfilePackCreate returned error: %v", err)
	}
	assert.EqualValues(t, 0, entity.InstallOrder,
		"missing install_order should default to 0 without panic")
}

// TestToApplicationProfilePackUpdate_InstallOrder_PLT2297 covers the update
// path: reordering install_order on an existing pack in HCL must reach the API.
func TestToApplicationProfilePackUpdate_InstallOrder_PLT2297(t *testing.T) {
	pack := map[string]interface{}{
		"type":          "manifest",
		"name":          "tier-backend",
		"tag":           "1.0.0",
		"install_order": 2,
		"values":        "",
		"manifest":      []interface{}{},
	}
	entity := toApplicationProfilePackUpdate(pack)
	assert.EqualValues(t, 2, entity.InstallOrder,
		"PLT-2297: install_order changes must be propagated on update")
}

func TestResourceApplicationProfileDelete(t *testing.T) {
	d := getBaseResourceData()
	var ctx context.Context
	d.SetId("test-app-profile-id")
	r := resourceApplicationProfileDelete(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, false, r.HasError())
}
