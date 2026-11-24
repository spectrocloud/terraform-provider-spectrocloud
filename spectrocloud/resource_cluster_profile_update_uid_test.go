package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

// Day-2 pack tag change: an update must send the UID of the tag being applied, not the UID Read stored in state for
// the previous tag. pack.uid is Optional+Computed, so a configuration without uid gets the old UID back from state.

const testPackRegistryUID = "test-oci-registry-for-day2" // the OCI basic sync-status mock reports sync supported

func day2PackMap(name, tag, stateUID string) map[string]interface{} {
	return map[string]interface{}{
		"name":          name,
		"tag":           tag,
		"uid":           stateUID,
		"type":          string(models.V1PackTypeOci),
		"registry_uid":  testPackRegistryUID,
		"registry_name": "",
		"values":        "",
		"manifest":      []interface{}{},
	}
}

func day2PackRef(name, tag, packUID, registryUID string) *models.V1PackRef {
	return &models.V1PackRef{Name: types.Ptr(name), Tag: tag, PackUID: packUID, RegistryUID: registryUID}
}

func day2Client(t *testing.T) *client.V1Client {
	t.Helper()
	return castV1Client(t, unitTestMockAPIClient)
}

func TestPackUpdateUIDResolution(t *testing.T) {
	c := day2Client(t)
	no, yes := false, true
	existing := []*models.V1PackRef{
		day2PackRef(routes.TwoVersionPackName, routes.TwoVersionPackOld, routes.TwoVersionPackOldID, testPackRegistryUID),
	}

	tests := []struct {
		name          string
		pack          map[string]interface{}
		refs          []*models.V1PackRef
		uidConfigured *bool
		wantUID       string
	}{
		{
			name:          "tag change without uid in config resolves the new tag's UID",
			pack:          day2PackMap(routes.TwoVersionPackName, routes.TwoVersionPackNew, routes.TwoVersionPackOldID),
			refs:          existing,
			uidConfigured: &no,
			wantUID:       routes.TwoVersionPackNewID,
		},
		{
			name: "unchanged pack reuses the profile's UID without a lookup",
			pack: day2PackMap(routes.TwoVersionPackName, routes.TwoVersionPackOld, routes.TwoVersionPackOldID),
			// A UID the lookup would never return proves no resolution happened.
			refs:          []*models.V1PackRef{day2PackRef(routes.TwoVersionPackName, routes.TwoVersionPackOld, "profile-current-uid", testPackRegistryUID)},
			uidConfigured: &no,
			wantUID:       "profile-current-uid",
		},
		{
			name:          "uid set in config is always kept, even when the tag changes",
			pack:          day2PackMap(routes.TwoVersionPackName, routes.TwoVersionPackNew, "user-set-uid"),
			refs:          existing,
			uidConfigured: &yes,
			wantUID:       "user-set-uid",
		},
		{
			name: "new pack carrying another pack's UID from state resolves its own",
			pack: day2PackMap(routes.TwoVersionPackName, routes.TwoVersionPackNew, "uid-of-a-different-pack"),
			refs: []*models.V1PackRef{
				day2PackRef("some-other-pack", "1.0", "uid-of-a-different-pack", testPackRegistryUID),
			},
			uidConfigured: &no,
			wantUID:       routes.TwoVersionPackNewID,
		},
		{
			name: "registry change resolves against the new registry",
			pack: day2PackMap(routes.TwoVersionPackName, routes.TwoVersionPackOld, "uid-from-old-registry"),
			refs: []*models.V1PackRef{
				day2PackRef(routes.TwoVersionPackName, routes.TwoVersionPackOld, "uid-from-old-registry", "old-registry-uid"),
			},
			uidConfigured: &no,
			wantUID:       routes.TwoVersionPackOldID,
		},
		{
			name:    "without raw config, a tag change still re-resolves",
			pack:    day2PackMap(routes.TwoVersionPackName, routes.TwoVersionPackNew, routes.TwoVersionPackOldID),
			refs:    existing,
			wantUID: routes.TwoVersionPackNewID,
		},
		{
			name:    "without raw config, an unchanged tag keeps the UID",
			pack:    day2PackMap(routes.TwoVersionPackName, routes.TwoVersionPackOld, routes.TwoVersionPackOldID),
			refs:    existing,
			wantUID: routes.TwoVersionPackOldID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toClusterProfilePackUpdateWithResolution(tt.pack, tt.refs, c, tt.uidConfigured)
			require.NoError(t, err)
			assert.Equal(t, tt.wantUID, got.UID)
			assert.Equal(t, tt.pack["tag"], got.Tag)
		})
	}
}

func TestPackUpdateManifestPackUnchanged(t *testing.T) {
	no := false
	pack := day2PackMap("my-manifest", "1.0.0", "")
	pack["type"] = string(models.V1PackTypeManifest)
	pack["registry_uid"] = ""
	got, err := toClusterProfilePackUpdateWithResolution(pack, nil, day2Client(t), &no)
	require.NoError(t, err)
	assert.Equal(t, "spectro-manifest-pack", got.UID)
}

// clusterProfileDay2Data builds the ResourceData Terraform passes to Update: prior state from statePacks, planned
// from configPacks (whose uid may be left out), with the configuration available as raw config, as during apply.
func clusterProfileDay2Data(t *testing.T, statePacks []map[string]interface{}, configPacks []map[string]cty.Value) *schema.ResourceData {
	t.Helper()
	res := resourceClusterProfile()

	stateRaw := map[string]interface{}{
		"name":    "day2-profile",
		"type":    "add-on",
		"cloud":   "all",
		"version": "1.0.0",
	}
	var packs []interface{}
	for _, p := range statePacks {
		packs = append(packs, p)
	}
	stateRaw["pack"] = packs
	stateRD := schema.TestResourceDataRaw(t, res.Schema, stateRaw)
	stateRD.SetId("day2-profile-uid")
	state := stateRD.State()
	require.NotNil(t, state)

	var cfgPacks []cty.Value
	for _, p := range configPacks {
		cfgPacks = append(cfgPacks, cty.ObjectVal(p))
	}
	block := res.CoreConfigSchema()
	configVal, err := block.CoerceValue(cty.ObjectVal(map[string]cty.Value{
		"name":    cty.StringVal("day2-profile"),
		"type":    cty.StringVal("add-on"),
		"cloud":   cty.StringVal("all"),
		"version": cty.StringVal("1.0.0"),
		"pack":    cty.ListVal(cfgPacks),
	}))
	require.NoError(t, err)
	state.RawConfig = configVal

	diff, err := res.Diff(context.Background(), state, terraform.NewResourceConfigShimmed(configVal, block), nil)
	require.NoError(t, err)
	d, err := schema.InternalMap(res.Schema).Data(state, diff)
	require.NoError(t, err)
	d.SetId("day2-profile-uid")
	return d
}

// day2ConfigPack is a pack block as written in the configuration, with uid left out (null).
func day2ConfigPack(name, tag string) map[string]cty.Value {
	return map[string]cty.Value{
		"name":         cty.StringVal(name),
		"tag":          cty.StringVal(tag),
		"type":         cty.StringVal(string(models.V1PackTypeOci)),
		"registry_uid": cty.StringVal(testPackRegistryUID),
		"uid":          cty.NullVal(cty.String),
	}
}

func day2Profile(refs ...*models.V1PackRef) *models.V1ClusterProfile {
	return &models.V1ClusterProfile{
		Spec: &models.V1ClusterProfileSpec{Published: &models.V1ClusterProfileTemplate{Packs: refs}},
	}
}

func packUIDsByName(entity *models.V1ClusterProfileUpdateEntity) map[string]string {
	out := map[string]string{}
	for _, p := range entity.Spec.Template.Packs {
		out[*p.Name] = p.UID
	}
	return out
}

// The ticket's scenario through a real plan: the configuration leaves uid out, state holds the 0.8.1 UID, the tag
// changes to 0.16.2. The update must carry the 0.16.2 UID.
func TestClusterProfileUpdateTagChangeSendsNewUID(t *testing.T) {
	d := clusterProfileDay2Data(t,
		[]map[string]interface{}{day2PackMap(routes.TwoVersionPackName, routes.TwoVersionPackOld, routes.TwoVersionPackOldID)},
		[]map[string]cty.Value{day2ConfigPack(routes.TwoVersionPackName, routes.TwoVersionPackNew)},
	)
	require.Equal(t, routes.TwoVersionPackOldID, d.Get("pack.0.uid"), "precondition: the plan carries the old UID from state")

	entity, err := toClusterProfileUpdateWithResolution(d, day2Profile(
		day2PackRef(routes.TwoVersionPackName, routes.TwoVersionPackOld, routes.TwoVersionPackOldID, testPackRegistryUID),
	), day2Client(t))
	require.NoError(t, err)
	require.Len(t, entity.Spec.Template.Packs, 1)
	assert.Equal(t, routes.TwoVersionPackNew, entity.Spec.Template.Packs[0].Tag)
	assert.Equal(t, routes.TwoVersionPackNewID, entity.Spec.Template.Packs[0].UID)
}

// Adding a pack in front shifts the UIDs state holds by position. Each pack must still get its own UID.
func TestClusterProfileUpdateInsertedPackKeepsUIDsWithTheirPacks(t *testing.T) {
	d := clusterProfileDay2Data(t,
		[]map[string]interface{}{
			day2PackMap("pack-a", "1.0", "uid-a"),
			day2PackMap("pack-b", "1.0", "uid-b"),
		},
		[]map[string]cty.Value{
			day2ConfigPack(routes.TwoVersionPackName, routes.TwoVersionPackNew),
			day2ConfigPack("pack-a", "1.0"),
			day2ConfigPack("pack-b", "1.0"),
		},
	)

	// Precondition: the plan pairs UIDs with packs by position, so state's UIDs have shifted onto the wrong packs.
	require.Equal(t, "uid-a", d.Get("pack.0.uid"), "the new pack carries pack-a's UID")
	require.Equal(t, "uid-b", d.Get("pack.1.uid"), "pack-a carries pack-b's UID")

	entity, err := toClusterProfileUpdateWithResolution(d, day2Profile(
		day2PackRef("pack-a", "1.0", "uid-a", testPackRegistryUID),
		day2PackRef("pack-b", "1.0", "uid-b", testPackRegistryUID),
	), day2Client(t))
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		routes.TwoVersionPackName: routes.TwoVersionPackNewID,
		"pack-a":                  "uid-a",
		"pack-b":                  "uid-b",
	}, packUIDsByName(entity))
}

func TestPackUIDsConfigured(t *testing.T) {
	withUID := day2ConfigPack("pack-a", "1.0")
	withUID["uid"] = cty.StringVal("user-uid")
	d := clusterProfileDay2Data(t,
		[]map[string]interface{}{day2PackMap("pack-a", "1.0", "user-uid"), day2PackMap("pack-b", "1.0", "uid-b")},
		[]map[string]cty.Value{withUID, day2ConfigPack("pack-b", "1.0")},
	)
	assert.Equal(t, []bool{true, false}, packUIDsConfigured(d))

	assert.Nil(t, packUIDsConfigured(resourceClusterProfile().TestResourceData()), "no raw config means unknown")
}
