module github.com/spectrocloud/terraform-provider-spectrocloud

go 1.15

require (
	emperror.dev/errors v0.7.0
	github.com/go-openapi/runtime v0.19.24
	github.com/go-openapi/strfmt v0.20.0
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320
	github.com/hashicorp/terraform-plugin-docs v0.3.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.4.0
	github.com/prometheus/common v0.10.0
	github.com/robfig/cron v1.2.0
	github.com/spectrocloud/gomi v0.0.0-20201228182306-1e482c900582
	github.com/spectrocloud/hapi v1.9.1-0.20210511133047-7b41ab750e0d
)

// replace github.com/spectrocloud/hapi => ../hapi
