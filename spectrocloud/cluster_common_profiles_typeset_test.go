package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
)

func clusterProfileTestData(t *testing.T, profiles []interface{}) *schema.ResourceData {
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"cluster_profile": schemas.ClusterProfileSchemaV2(),
	}, map[string]interface{}{})
	if err := d.Set("cluster_profile", profiles); err != nil {
		t.Fatalf("set cluster_profile: %v", err)
	}
	return d
}

// PLT-2493: the empty-id entry left behind after changing a profile block with
// pack overrides + variables must be dropped from state; real entries are kept intact.
func TestDropEmptyClusterProfileEntries_RemovesPhantom(t *testing.T) {
	real := map[string]interface{}{
		"id": "profile-1",
		"pack": []interface{}{
			map[string]interface{}{"name": "kubernetes-eks", "tag": "1.35", "values": "a: b"},
		},
		"variables": map[string]interface{}{"aws_account_id": "1"},
	}
	phantom := map[string]interface{}{"id": "", "pack": []interface{}{}, "variables": map[string]interface{}{}}
	d := clusterProfileTestData(t, []interface{}{phantom, real})
	if n := len(normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))); n != 2 {
		t.Fatalf("precondition: expected 2 entries, got %d", n)
	}

	if err := dropEmptyClusterProfileEntries(d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))
	if len(out) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out))
	}
	m := out[0].(map[string]interface{})
	if m["id"] != "profile-1" {
		t.Fatalf("expected id=profile-1, got %v", m["id"])
	}
	if len(m["pack"].([]interface{})) != 1 || m["variables"].(map[string]interface{})["aws_account_id"] != "1" {
		t.Fatalf("expected packs and variables preserved, got %+v", m)
	}
}

func TestDropEmptyClusterProfileEntries_NoPhantomNoChange(t *testing.T) {
	d := clusterProfileTestData(t, []interface{}{
		map[string]interface{}{"id": "profile-1"},
		map[string]interface{}{"id": "profile-2"},
	})

	if err := dropEmptyClusterProfileEntries(d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n := len(normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))); n != 2 {
		t.Fatalf("expected 2 entries, got %d", n)
	}
}

func TestDropEmptyClusterProfileEntries_Unset(t *testing.T) {
	d := clusterProfileTestData(t, nil)

	if err := dropEmptyClusterProfileEntries(d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n := len(normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))); n != 0 {
		t.Fatalf("expected 0 entries, got %d", n)
	}
}

func TestNormalizeInterfaceSliceFromListOrSet_List(t *testing.T) {
	in := []interface{}{
		map[string]interface{}{"id": "profile-1"},
		map[string]interface{}{"id": "profile-2"},
	}

	out := normalizeInterfaceSliceFromListOrSet(in)
	if len(out) != 2 {
		t.Fatalf("expected 2 items, got %d", len(out))
	}
	if out[0].(map[string]interface{})["id"] != "profile-1" {
		t.Fatalf("expected first id=profile-1, got %v", out[0].(map[string]interface{})["id"])
	}
	if out[1].(map[string]interface{})["id"] != "profile-2" {
		t.Fatalf("expected second id=profile-2, got %v", out[1].(map[string]interface{})["id"])
	}
}

func TestNormalizeInterfaceSliceFromListOrSet_Set(t *testing.T) {
	elems := []interface{}{
		map[string]interface{}{"id": "profile-1"},
		map[string]interface{}{"id": "profile-2"},
	}

	hashByID := func(v interface{}) int {
		m := v.(map[string]interface{})
		return schema.HashString(m["id"].(string))
	}

	set := schema.NewSet(hashByID, elems)
	out := normalizeInterfaceSliceFromListOrSet(set)

	ids := make(map[string]bool)
	for _, v := range out {
		ids[v.(map[string]interface{})["id"].(string)] = true
	}

	if !ids["profile-1"] || !ids["profile-2"] {
		t.Fatalf("expected ids profile-1 and profile-2 to be present, got %+v", ids)
	}
}

func TestResourceClusterEksStateUpgradeV3_ClusterProfileListPreserved(t *testing.T) {
	raw := map[string]interface{}{
		"cluster_profile": []interface{}{
			map[string]interface{}{"id": "profile-1"},
			map[string]interface{}{"id": "profile-2"},
		},
	}

	out, err := resourceClusterEksStateUpgradeV3(context.Background(), raw, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cp, ok := out["cluster_profile"].([]interface{})
	if !ok {
		t.Fatalf("expected cluster_profile to be []interface{}, got %T", out["cluster_profile"])
	}
	if len(cp) != 2 {
		t.Fatalf("expected 2 cluster_profile items, got %d", len(cp))
	}
}

func TestResourceClusterAwsStateUpgradeV2_ClusterProfileListPreserved(t *testing.T) {
	raw := map[string]interface{}{
		"cluster_profile": []interface{}{
			map[string]interface{}{"id": "profile-1"},
			map[string]interface{}{"id": "profile-2"},
		},
	}

	out, err := resourceClusterAwsStateUpgradeV2(context.Background(), raw, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cp, ok := out["cluster_profile"].([]interface{})
	if !ok {
		t.Fatalf("expected cluster_profile to be []interface{}, got %T", out["cluster_profile"])
	}
	if len(cp) != 2 {
		t.Fatalf("expected 2 cluster_profile items, got %d", len(cp))
	}
}
