module github.com/spectrocloud/terraform-provider-spectrocloud

go 1.14

require (
	github.com/go-openapi/runtime v0.19.4
	github.com/go-openapi/strfmt v0.19.3
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.0.0-rc.2
	github.com/prometheus/common v0.10.0
	github.com/spectrocloud/gomi v0.0.0-20201113051324-08a1179400db
	github.com/spectrocloud/hapi v1.6.1-0.20201215134830-ba674db801be
)

// replace github.com/spectrocloud/hapi => ../hapi
