module github.com/spectrocloud/terraform-provider-spectrocloud

go 1.15

require (
	emperror.dev/errors v0.8.0
	github.com/go-openapi/runtime v0.19.28
	github.com/go-openapi/strfmt v0.20.1
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320
	github.com/hashicorp/terraform-plugin-docs v0.3.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.6.1
	github.com/prometheus/common v0.23.0
	github.com/robfig/cron v1.2.0
	github.com/spectrocloud/gomi v1.9.1-0.20210519044035-5333c9359877
	github.com/spectrocloud/hapi v1.14.0
)

//replace github.com/spectrocloud/hapi => ../hapi
