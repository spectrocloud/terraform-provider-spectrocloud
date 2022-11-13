---
page_title: "spectrocloud_alert Resource - terraform-provider-spectrocloud"
subcategory: ""
description: Provisioning Cluster Health alerts (email | http).|-
  
---

# Resource `spectrocloud_alert`



## Example Usage
#### Note
- Spectrocloud allow to create upto 2 alert in below combinations
    - 1 email alert (can add any number of email recipient) & 1 http webhook alert.
    - 2 http webhook alert configuration.
    - Any one alert type Email/http.

  (Documentation [Cluster Health Alerts](#https://docs.spectrocloud.com/clusters/cluster-management/health-alerts#overview))




```terraform
resource "spectrocloud_alert" "alert_dev" {
  project = "dev"
  is_active = true
  component = "ClusterHealth"
  http {
    method = "POST"
    url = "https://openhook.com/put/dev0"
    body = "{ \"text\": \"{{message}}\" }"
    headers = {
      ApiKey = "test--key--dev"
      tag = "test"
      source = "spectro"
    }
  }
  email {
    alert_all_users = false
    identifiers = ["abc@spectrocloud.com"]
  }
}
```

## Schema

### Required

- **project (name)** (String)
- **is_active** (Bool)
- **component** (Bool)
- **alert (http/email)** (Map)


### Optional

- **identifiers** (String)
- **body** (String)
- **headers** ([]Map)
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **delete** (String)
- **update** (String)
