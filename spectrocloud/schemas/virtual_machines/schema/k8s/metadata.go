package k8s

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// filterSystemAnnotations removes system-managed annotations that should not be managed by Terraform
func filterSystemAnnotations(annotations map[string]string) map[string]string {
	if annotations == nil {
		return nil
	}

	filtered := make(map[string]string)

	for key, value := range annotations {
		// Filter out all kubevirt.io/ system annotations
		if !strings.HasPrefix(key, "kubevirt.io/") {
			filtered[key] = value
		}
	}

	return filtered
}

func metadataFields(objectName string) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"annotations": {
			Type:         schema.TypeMap,
			Description:  fmt.Sprintf("An unstructured key value map stored with the %s that may be used to store arbitrary metadata. More info: http://kubernetes.io/docs/user-guide/annotations", objectName),
			Optional:     true,
			Elem:         &schema.Schema{Type: schema.TypeString},
			ValidateFunc: utils.ValidateAnnotations,
			Computed:     true,
		},
		"generation": {
			Type:        schema.TypeInt,
			Description: "A sequence number representing a specific generation of the desired state.",
			Computed:    true,
		},
		"labels": {
			Type:         schema.TypeMap,
			Description:  fmt.Sprintf("Map of string keys and values that can be used to organize and categorize (scope and select) the %s. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels", objectName),
			Optional:     true,
			Elem:         &schema.Schema{Type: schema.TypeString},
			ValidateFunc: utils.ValidateLabels,
		},
		"name": {
			Type:         schema.TypeString,
			Description:  fmt.Sprintf("Name of the %s, must be unique. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names", objectName),
			Optional:     true,
			Computed:     true,
			ForceNew:     true,
			ValidateFunc: utils.ValidateName,
		},
		"resource_version": {
			Type:        schema.TypeString,
			Description: fmt.Sprintf("An opaque value that represents the internal version of this %s that can be used by clients to determine when %s has changed. Read more: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency", objectName, objectName),
			Computed:    true,
		},
		"uid": {
			Type:        schema.TypeString,
			Description: fmt.Sprintf("The unique in time and space value for this %s. More info: http://kubernetes.io/docs/user-guide/identifiers#uids", objectName),
			Computed:    true,
		},
	}
}

func NamespacedMetadataSchema(objectName string, generatableName bool) *schema.Schema {
	return namespacedMetadataSchemaIsTemplate(objectName, generatableName, false)
}

func namespacedMetadataSchemaIsTemplate(objectName string, generatableName, isTemplate bool) *schema.Schema {
	fields := metadataFields(objectName)
	fields["namespace"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: fmt.Sprintf("Namespace defines the space within which name of the %s must be unique.", objectName),
		Optional:    true,
		ForceNew:    true,
		Default: (func() interface{} {
			if isTemplate {
				return nil
			}
			return "default"
		})(),
	}
	if generatableName {
		fields["generate_name"] = &schema.Schema{
			Type:          schema.TypeString,
			Description:   "Prefix, used by the server, to generate a unique name ONLY IF the `name` field has not been provided. This value will also be combined with a unique suffix. Read more: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#idempotency",
			Optional:      true,
			ValidateFunc:  utils.ValidateGenerateName,
			ConflictsWith: []string{"metadata.name"},
		}
		fields["name"].ConflictsWith = []string{"metadata.generate_name"}
	}

	return &schema.Schema{
		Type:        schema.TypeList,
		Description: fmt.Sprintf("Standard %s's metadata. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata", objectName),
		Required:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}
}

func ConvertToBasicMetadata(d *schema.ResourceData) metav1.ObjectMeta {
	var meta []interface{}
	metaValues := make(map[string]interface{})
	if v, ok := d.GetOk("annotations"); ok && len(v.(map[string]interface{})) > 0 {
		metaValues["annotations"] = utils.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("labels"); ok && len(v.(map[string]interface{})) > 0 {
		metaValues["labels"] = utils.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("generate_name"); ok {
		metaValues["generate_name"] = v.(string)
	}
	if v, ok := d.GetOk("name"); ok {
		metaValues["name"] = v.(string)
	}
	if v, ok := d.GetOk("namespace"); ok {
		metaValues["namespace"] = v.(string)
	}
	if v, ok := d.GetOk("resource_version"); ok {
		metaValues["resource_version"] = v.(string)
	}
	meta = append(meta, metaValues)
	return ExpandMetadata(meta)
}

func ExpandMetadata(in []interface{}) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{}
	if len(in) < 1 {
		return meta
	}
	m := in[0].(map[string]interface{})

	if v, ok := m["annotations"].(map[string]string); ok && len(v) > 0 {
		meta.Annotations = m["annotations"].(map[string]string) //utils.ExpandStringMap(m["annotations"].(map[string]interface{}))
	} else if v, ok := m["annotations"].(map[string]interface{}); ok && len(v) > 0 { // for supporting data volume templates annotations
		meta.Annotations = utils.ExpandStringMap(m["annotations"].(map[string]interface{}))
	}

	if v, ok := m["labels"].(map[string]string); ok && len(v) > 0 {
		meta.Labels = m["labels"].(map[string]string) //utils.ExpandStringMap(m["labels"])
	}

	if v, ok := m["generate_name"]; ok {
		meta.GenerateName = v.(string)
	}
	if v, ok := m["name"]; ok {
		meta.Name = v.(string)
	}
	if v, ok := m["namespace"]; ok {
		meta.Namespace = v.(string)
	}
	if v, ok := m["resource_version"]; ok {
		meta.ResourceVersion = v.(string)
	}

	return meta
}

func FlattenMetadataDataVolume(meta metav1.ObjectMeta) []interface{} {
	m := make(map[string]interface{})
	// Filter out system-managed annotations for data volumes as well
	filteredAnnotations := filterSystemAnnotations(meta.Annotations)
	m["annotations"] = utils.FlattenStringMap(filteredAnnotations)
	if meta.GenerateName != "" {
		m["generate_name"] = meta.GenerateName
	}
	m["labels"] = utils.FlattenStringMap(meta.Labels)
	m["name"] = meta.Name
	m["resource_version"] = meta.ResourceVersion
	m["uid"] = fmt.Sprintf("%v", meta.UID)
	m["generation"] = meta.Generation

	if meta.Namespace != "" {
		m["namespace"] = meta.Namespace
	}

	return []interface{}{m}
}

func FlattenMetadata(meta metav1.ObjectMeta, resourceData *schema.ResourceData) error {
	var err error
	if resourceData == nil {
		return err
	}
	// Filter out system-managed annotations before setting them in Terraform state
	filteredAnnotations := filterSystemAnnotations(meta.Annotations)
	if err = resourceData.Set("annotations", utils.FlattenStringMap(filteredAnnotations)); err != nil {
		return err
	}
	if err = resourceData.Set("labels", utils.FlattenStringMap(meta.Labels)); err != nil {
		return err
	}
	if err = resourceData.Set("name", meta.Name); err != nil {
		return err
	}
	if err = resourceData.Set("resource_version", meta.ResourceVersion); err != nil {
		return err
	}
	if err = resourceData.Set("uid", fmt.Sprintf("%v", meta.UID)); err != nil {
		return err
	}
	if err = resourceData.Set("generation", int(meta.Generation)); err != nil { //fmt.Sprintf("%v", meta.Generation)
		return err
	}
	if err = resourceData.Set("namespace", fmt.Sprintf("%v", meta.Namespace)); err != nil {
		return err
	}

	return err
}
