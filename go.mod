module github.com/spectrocloud/terraform-provider-spectrocloud

go 1.15

require (
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/go-openapi/runtime v0.19.4
	github.com/go-openapi/strfmt v0.19.3
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320
	github.com/hashicorp/go-getter v1.5.1 // indirect
	github.com/hashicorp/hcl/v2 v2.8.2 // indirect
	github.com/hashicorp/terraform-plugin-docs v0.3.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.4.0
	github.com/pkg/errors v0.9.1
	github.com/posener/complete v1.2.1 // indirect
	github.com/prometheus/common v0.10.0
	github.com/robfig/cron v1.2.0
	github.com/spectrocloud/gomi v0.0.0-20201228182306-1e482c900582
	github.com/spectrocloud/hapi v1.6.1-0.20210128095342-4df47d1cb56a
	github.com/zclconf/go-cty v1.7.1 // indirect
	golang.org/x/tools v0.0.0-20201028111035-eafbe7b904eb // indirect
	google.golang.org/api v0.34.0 // indirect
)

// replace github.com/spectrocloud/hapi => ../hapi
