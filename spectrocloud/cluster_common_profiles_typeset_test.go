package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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
