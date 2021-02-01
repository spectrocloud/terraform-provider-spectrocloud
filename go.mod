module github.com/spectrocloud/terraform-provider-spectrocloud

go 1.15

require (
	github.com/go-openapi/runtime v0.19.4
	github.com/go-openapi/strfmt v0.19.3
	github.com/hashicorp/terraform-plugin-docs v0.3.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.4.0
	github.com/prometheus/common v0.10.0
	github.com/spectrocloud/gomi v0.0.0-20201228182306-1e482c900582
	github.com/spectrocloud/hapi v1.6.1-0.20210128095342-4df47d1cb56a
)

// replace github.com/spectrocloud/hapi => ../hapi
