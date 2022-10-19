module github.com/spectrocloud/terraform-provider-spectrocloud

go 1.15

require (
	github.com/go-openapi/runtime v0.19.28
	github.com/go-openapi/strfmt v0.20.1
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320
	github.com/hashicorp/terraform-plugin-docs v0.3.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.6.1
	github.com/robfig/cron v1.2.0
	github.com/spectrocloud/gomi v1.14.1-0.20220727130240-fc64f32e394d
	github.com/spectrocloud/hapi v1.14.1-0.20221019101938-4a9c4e9e4bcf
)

//replace github.com/spectrocloud/hapi => ../hapi
