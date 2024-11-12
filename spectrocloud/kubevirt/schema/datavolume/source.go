package datavolume

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func dataVolumeSourceFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"blank":    dataVolumeSourceBlankSchema(),
		"http":     dataVolumeSourceHTTPSchema(),
		"pvc":      dataVolumeSourcePVCSchema(),
		"registry": dataVolumeSourceRegistrySchema(),
	}
}

func dataVolumeSourceBlankSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "DataVolumeSourceBlank provides the parameters to create a Data Volume from an empty source.",
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
}

func dataVolumeSourceSchema() *schema.Schema {
	fields := dataVolumeSourceFields()

	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Source is the src of the data for the requested DataVolume.",
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}

}

func dataVolumeSourceHTTPFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Description: "url is the URL of the http source.",
			Optional:    true,
		},
		"secret_ref": {
			Type:        schema.TypeString,
			Description: "Secret_ref provides the secret reference needed to access the HTTP source.",
			Optional:    true,
		},
		"cert_config_map": {
			Type:        schema.TypeString,
			Description: "Cert_config_map provides a reference to the Registry certs.",
			Optional:    true,
		},
	}
}

func dataVolumeSourceHTTPSchema() *schema.Schema {
	fields := dataVolumeSourceHTTPFields()

	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "DataVolumeSourceHTTP provides the parameters to create a Data Volume from an HTTP source.",
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}

}

func dataVolumeSourcePVCFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"namespace": {
			Type:        schema.TypeString,
			Description: "The namespace which the PVC located in.",
			Optional:    true,
		},
		"name": {
			Type:        schema.TypeString,
			Description: "The name of the PVC.",
			Optional:    true,
		},
	}
}

func dataVolumeSourceRegistryFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"image_url": {
			Type:        schema.TypeString,
			Description: "The registry URL of the image to download.",
			Optional:    true,
		},
	}
}

func dataVolumeSourcePVCSchema() *schema.Schema {
	fields := dataVolumeSourcePVCFields()

	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "DataVolumeSourcePVC provides the parameters to create a Data Volume from an existing PVC.",
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}

}

func dataVolumeSourceRegistrySchema() *schema.Schema {
	fields := dataVolumeSourceRegistryFields()

	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "DataVolumeSourceRegistry provides the parameters to create a Data Volume from an existing PVC.",
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}
}

// Expanders

func expandDataVolumeSource(dataVolumeSource []interface{}) *cdiv1.DataVolumeSource {
	result := &cdiv1.DataVolumeSource{}

	if len(dataVolumeSource) == 0 || dataVolumeSource[0] == nil {
		return result
	}

	in := dataVolumeSource[0].(map[string]interface{})

	result.Blank = expandDataVolumeSourceBlank(in["blank"].([]interface{}))
	result.HTTP = expandDataVolumeSourceHTTP(in["http"].([]interface{}))
	result.PVC = expandDataVolumeSourcePVC(in["pvc"].([]interface{}))
	result.Registry = expandDataVolumeSourceRegistry(in["registry"].([]interface{}))

	return result
}

func expandDataVolumeSourceBlank(dataVolumeSourceBlank []interface{}) *cdiv1.DataVolumeBlankImage {
	if len(dataVolumeSourceBlank) == 0 || dataVolumeSourceBlank[0] == nil {
		return nil
	}

	result := &cdiv1.DataVolumeBlankImage{}

	return result
}

func expandDataVolumeSourceHTTP(dataVolumeSourceHTTP []interface{}) *cdiv1.DataVolumeSourceHTTP {
	if len(dataVolumeSourceHTTP) == 0 || dataVolumeSourceHTTP[0] == nil {
		return nil
	}

	result := &cdiv1.DataVolumeSourceHTTP{}

	in := dataVolumeSourceHTTP[0].(map[string]interface{})

	if v, ok := in["url"].(string); ok {
		result.URL = v
	}
	if v, ok := in["secret_ref"].(string); ok {
		result.SecretRef = v
	}
	if v, ok := in["cert_config_map"].(string); ok {
		result.CertConfigMap = v
	}

	return result
}

func expandDataVolumeSourcePVC(dataVolumeSourcePVC []interface{}) *cdiv1.DataVolumeSourcePVC {
	if len(dataVolumeSourcePVC) == 0 || dataVolumeSourcePVC[0] == nil {
		return nil
	}

	result := &cdiv1.DataVolumeSourcePVC{}

	in := dataVolumeSourcePVC[0].(map[string]interface{})

	if v, ok := in["namespace"].(string); ok {
		result.Namespace = v
	}
	if v, ok := in["name"].(string); ok {
		result.Name = v
	}

	return result
}

func expandDataVolumeSourceRegistry(dataVolumeSourceRegistry []interface{}) *cdiv1.DataVolumeSourceRegistry {
	if len(dataVolumeSourceRegistry) == 0 || dataVolumeSourceRegistry[0] == nil {
		return nil
	}

	result := &cdiv1.DataVolumeSourceRegistry{}

	in := dataVolumeSourceRegistry[0].(map[string]interface{})

	if v, ok := in["image_url"].(string); ok {
		result.URL = ptr.To(v)
	}

	return result
}

// Flatteners

func flattenDataVolumeSource(in *cdiv1.DataVolumeSource) []interface{} {
	att := make(map[string]interface{})
	if in != nil {
		if in.Blank != nil {
			att["blank"] = flattenDataVolumeSourceBlank()
		}
		if in.HTTP != nil {
			att["http"] = flattenDataVolumeSourceHTTP(*in.HTTP)
		}
		if in.PVC != nil {
			att["pvc"] = flattenDataVolumeSourcePVC(*in.PVC)
		}
		if in.Registry != nil {
			att["registry"] = flattenDataVolumeSourceRegistry(*in.Registry)
		}

		return []interface{}{att}
	}
	return []interface{}{}
}

func flattenDataVolumeSourceBlank() []interface{} {
	att := map[string]interface{}{}
	return []interface{}{att}
}

func flattenDataVolumeSourceHTTP(in cdiv1.DataVolumeSourceHTTP) []interface{} {
	att := map[string]interface{}{
		"url":             in.URL,
		"secret_ref":      in.SecretRef,
		"cert_config_map": in.CertConfigMap,
	}
	return []interface{}{att}
}

func flattenDataVolumeSourcePVC(in cdiv1.DataVolumeSourcePVC) []interface{} {
	att := map[string]interface{}{
		"namespace": in.Namespace,
		"name":      in.Name,
	}
	return []interface{}{att}
}

func flattenDataVolumeSourceRegistry(in cdiv1.DataVolumeSourceRegistry) []interface{} {
	att := map[string]interface{}{
		"image_url": in.URL,
	}
	return []interface{}{att}
}
