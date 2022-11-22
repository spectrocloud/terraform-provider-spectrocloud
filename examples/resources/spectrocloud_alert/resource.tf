/*
**Note** - We can set up a maximum of two alerts for cluster health per project. Webhook can be configured in the HTTP component,
and for email, we can add a target email recipient or enable alerts for all users in the corresponding project
*/

resource "spectrocloud_alert" "alert_email" {
  project = "Default"
  is_active = true
  component = "ClusterHealth"
  type = "email"
  identifiers = ["siva@spectrocloud.com", "anand@spectrocloud.com"]
  alert_all_users = false
}

resource "spectrocloud_alert" "alert_http" {
  project = "Default"
  is_active = true
  component = "ClusterHealth"
  http {
    method  = "POST"
    url     = "https://openhook.com/put/edit2"
    body    = "{ \"text\": \"{{message}}\" }"
    headers = {
      type = "test--key--dev0"
      tag    = "Health"
      source = "spectrocloud"
    }
  }
  type = "http"
  alert_all_users = false
}
