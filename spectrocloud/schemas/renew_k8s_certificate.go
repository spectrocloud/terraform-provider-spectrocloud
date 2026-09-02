package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func RenewK8sCertificatesNowSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.IsRFC3339Time,
		Description: "Timestamp to trigger an immediate renewal of control plane Kubernetes PKI certificates for this cluster. " +
			"NOTE: The renewal is initiated immediately when this value changes - the timestamp does NOT schedule a future renewal. " +
			"Set this to the current timestamp each time you want to trigger certificate renewal. " +
			"This field can also be used for tracking when renewals were triggered. " +
			"Renewal may take several minutes depending on cluster size. Only control plane certificates are renewed; worker node certificates are not supported. " +
			"Format: RFC3339 (e.g., '2024-01-15T10:30:00Z').",
	}
}
