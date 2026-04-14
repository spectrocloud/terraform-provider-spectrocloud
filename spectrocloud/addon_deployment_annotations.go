package spectrocloud

import (
	"encoding/json"
	"log"
	"sort"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

// clusterAnnotationTerraformAddonDeployments stores JSON array of cluster profile template UIDs
// managed exclusively by spectrocloud_addon_deployment (PLT-2184 / design solution 1).
const clusterAnnotationTerraformAddonDeployments = "tf_addon_deployments"

func copyStringMap(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// terraformAddonManagedProfileUIDSet returns UIDs listed in the tf_addon_deployments annotation.
func terraformAddonManagedProfileUIDSet(cluster *models.V1SpectroCluster) map[string]struct{} {
	set := make(map[string]struct{})
	if cluster == nil || cluster.Metadata == nil || cluster.Metadata.Annotations == nil {
		return set
	}
	raw, ok := cluster.Metadata.Annotations[clusterAnnotationTerraformAddonDeployments]
	if !ok || raw == "" {
		return set
	}
	for _, uid := range parseTerraformAddonDeploymentAnnotation(raw) {
		if uid != "" {
			set[uid] = struct{}{}
		}
	}
	return set
}

func parseTerraformAddonDeploymentAnnotation(raw string) []string {
	var uids []string
	if err := json.Unmarshal([]byte(raw), &uids); err != nil {
		log.Printf("[WARN] %s: invalid JSON, ignoring: %v", clusterAnnotationTerraformAddonDeployments, err)
		return nil
	}
	return uids
}

func serializeTerraformAddonDeploymentUIDs(set map[string]struct{}) (string, error) {
	if len(set) == 0 {
		return "", nil
	}
	list := make([]string, 0, len(set))
	for uid := range set {
		if uid != "" {
			list = append(list, uid)
		}
	}
	sort.Strings(list)
	b, err := json.Marshal(list)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// mergeTerraformAddonDeploymentProfileUIDs read-modify-writes the annotation via UpdateClusterMetadata.
// add/remove are applied to the current annotation value; empty result removes the key.
func mergeTerraformAddonDeploymentProfileUIDs(c *client.V1Client, clusterUID string, add, remove []string) error {
	cluster, err := c.GetCluster(clusterUID)
	if err != nil {
		return err
	}
	if cluster.Metadata == nil {
		cluster.Metadata = &models.V1ObjectMeta{}
	}
	set := terraformAddonManagedProfileUIDSet(cluster)
	for _, uid := range remove {
		if uid != "" {
			delete(set, uid)
		}
	}
	for _, uid := range add {
		if uid != "" {
			set[uid] = struct{}{}
		}
	}

	ann := copyStringMap(cluster.Metadata.Annotations)
	if len(set) == 0 {
		delete(ann, clusterAnnotationTerraformAddonDeployments)
	} else {
		s, err := serializeTerraformAddonDeploymentUIDs(set)
		if err != nil {
			return err
		}
		ann[clusterAnnotationTerraformAddonDeployments] = s
	}

	md := &models.V1ObjectMetaInputEntity{
		Name:        cluster.Metadata.Name,
		Labels:      copyStringMap(cluster.Metadata.Labels),
		Annotations: ann,
	}
	return c.UpdateClusterMetadata(clusterUID, &models.V1ObjectMetaInputEntitySchema{Metadata: md})
}

// reconcileAddonDeploymentAnnotationAfterUpdate updates tf_addon_deployments after an addon deployment
// create/update. oldCompositeID is the resource id before SetId (empty on first create from update path).
func reconcileAddonDeploymentAnnotationAfterUpdate(c *client.V1Client, oldCompositeID string, newClusterUID string, newProfileUID string) error {
	if newProfileUID == "" {
		return nil
	}
	if oldCompositeID == "" {
		return mergeTerraformAddonDeploymentProfileUIDs(c, newClusterUID, []string{newProfileUID}, nil)
	}
	oldClusterUID := getClusterUID(oldCompositeID)
	oldProfileUID, err := getClusterProfileUID(oldCompositeID)
	if err != nil {
		oldProfileUID = ""
	}
	if oldClusterUID != "" && oldClusterUID != newClusterUID {
		if oldProfileUID != "" {
			if err := mergeTerraformAddonDeploymentProfileUIDs(c, oldClusterUID, nil, []string{oldProfileUID}); err != nil {
				return err
			}
		}
		return mergeTerraformAddonDeploymentProfileUIDs(c, newClusterUID, []string{newProfileUID}, nil)
	}
	if oldProfileUID != "" && oldProfileUID != newProfileUID {
		return mergeTerraformAddonDeploymentProfileUIDs(c, newClusterUID, []string{newProfileUID}, []string{oldProfileUID})
	}
	return mergeTerraformAddonDeploymentProfileUIDs(c, newClusterUID, []string{newProfileUID}, nil)
}
