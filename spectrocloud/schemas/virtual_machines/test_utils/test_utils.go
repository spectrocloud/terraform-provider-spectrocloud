package test_utils

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func NullifySchemaSetFunction(ss *schema.Set) {
	ss.F = nil
}

func GetPVCRequirements(dataVolume interface{}) interface{} {
	return dataVolume.(map[string]interface{})["spec"].([]interface{})[0].(map[string]interface{})["pvc"].([]interface{})[0].(map[string]interface{})["resources"].([]interface{})[0]
}
