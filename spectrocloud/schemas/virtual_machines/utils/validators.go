package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	apiValidation "k8s.io/apimachinery/pkg/api/validation"
	utilValidation "k8s.io/apimachinery/pkg/util/validation"
)

func ValidateAnnotations(value interface{}, key string) (ws []string, es []error) {
	m := value.(map[string]interface{})
	for k := range m {
		errors := utilValidation.IsQualifiedName(strings.ToLower(k))
		if len(errors) > 0 {
			for _, e := range errors {
				es = append(es, fmt.Errorf("%s (%q) %s", key, k, e))
			}
		}
	}
	return
}

func ValidateName(value interface{}, key string) (ws []string, es []error) {
	v := value.(string)
	errors := apiValidation.NameIsDNSSubdomain(v, false)
	if len(errors) > 0 {
		for _, err := range errors {
			es = append(es, fmt.Errorf("%s %s", key, err))
		}
	}
	return
}

func ValidateGenerateName(value interface{}, key string) (ws []string, es []error) {
	v := value.(string)

	errors := apiValidation.NameIsDNSLabel(v, true)
	if len(errors) > 0 {
		for _, err := range errors {
			es = append(es, fmt.Errorf("%s %s", key, err))
		}
	}
	return
}

func ValidateLabels(value interface{}, key string) (ws []string, es []error) {
	m := value.(map[string]interface{})
	for k, v := range m {
		for _, msg := range utilValidation.IsQualifiedName(k) {
			es = append(es, fmt.Errorf("%s (%q) %s", key, k, msg))
		}
		val, isString := v.(string)
		if !isString {
			es = append(es, fmt.Errorf("%s.%s (%#v): Expected value to be string", key, k, v))
			return
		}
		for _, msg := range utilValidation.IsValidLabelValue(val) {
			es = append(es, fmt.Errorf("%s (%q) %s", key, val, msg))
		}
	}
	return
}

// validateTypeStringNullableInt provides custom error messaging for TypeString ints
// Some arguments require an int value or unspecified, empty field.
func ValidateTypeStringNullableInt(v interface{}, k string) (ws []string, es []error) {
	value, ok := v.(string)
	if !ok {
		es = append(es, fmt.Errorf("expected type of %s to be string", k))
		return
	}

	if value == "" {
		return
	}

	if _, err := strconv.ParseInt(value, 10, 64); err != nil {
		es = append(es, fmt.Errorf("%s: cannot parse '%s' as int: %s", k, value, err))
	}

	return
}

func StringIsIntInRange(min int, max int) schema.SchemaValidateDiagFunc {
	return func(i interface{}, p cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics

		stringVal, ok := i.(string)
		if !ok {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Expected type of %s to be string", p),
			})
			return diags
		}

		intVal, err := strconv.Atoi(stringVal)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Expected %s to be string representation of integer", p),
			})
			return diags
		}

		if intVal < min || intVal > max {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Expected %s to be between %d and %d inclusive", p, min, max),
			})
			return diags
		}

		return diags
	}
}
