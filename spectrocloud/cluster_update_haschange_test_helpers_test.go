package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// (test file header) shared plumbing.
//
// schema.TestResourceData never produces a diff, so d.HasChange("foo")
// always returns false — which means every cluster Update func's
// HasChange-guarded branch (backup_policy, scan_policy, cloud_config,
// machine_pool, cluster_rbac_binding, etc.) sits at 0% coverage.
//
// This helper mirrors the pattern already used by
// resource_cluster_profile_test.go — build an InstanceState (baseline)
// + InstanceDiff (what changed), then call
// schema.InternalMap(res.Schema).Data(state, diff) to produce a
// ResourceData whose HasChange returns true for the diffed keys.
//
// Every Wave A test in Batches 14–16 uses this to fire the update
// branches on aws / aks / azure / gcp / gke / vsphere / maas / edge /
// cloudstack / attachment / brownfield.

// buildUpdateResourceData constructs a *schema.ResourceData from a
// baseline attribute map + a diff. The resulting object is what the
// Update function would see when Terraform's Plan detected changes to
// the keys in `diffAttrs`.
//
// baseAttrs is the "old" state — required attrs the resource schema
// treats as populated. diffAttrs is the delta — its keys are what
// d.HasChange(...) will return true for.
func buildUpdateResourceData(
	res *schema.Resource,
	id string,
	baseAttrs map[string]string,
	diffAttrs map[string]*terraform.ResourceAttrDiff,
) *schema.ResourceData {
	state := &terraform.InstanceState{
		ID:         id,
		Attributes: baseAttrs,
	}
	diff := &terraform.InstanceDiff{
		Attributes: diffAttrs,
	}
	d, err := schema.InternalMap(res.Schema).Data(state, diff)
	if err != nil {
		panic("buildUpdateResourceData: InternalMap.Data failed: " + err.Error())
	}
	return d
}

// simpleDiff builds an InstanceDiff.Attributes entry for a scalar
// old→new change. Terraform's diff format is a flat string map with
// "." separators; each ResourceAttrDiff records the old + new value at
// one key.
func simpleDiff(key, oldV, newV string) map[string]*terraform.ResourceAttrDiff {
	return map[string]*terraform.ResourceAttrDiff{
		key: {Old: oldV, New: newV},
	}
}

// mergeDiff combines multiple ResourceAttrDiff maps. Later maps
// overwrite earlier ones on key conflict.
func mergeDiff(diffs ...map[string]*terraform.ResourceAttrDiff) map[string]*terraform.ResourceAttrDiff {
	out := map[string]*terraform.ResourceAttrDiff{}
	for _, d := range diffs {
		for k, v := range d {
			out[k] = v
		}
	}
	return out
}
