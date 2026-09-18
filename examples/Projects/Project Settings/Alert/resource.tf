# Note: Palette allows a maximum of two alerts per project for the "ClusterHealth" component -
# one email alert and one HTTP webhook alert.
#
# Attributes:
#   project         - Required, ForceNew. Changing this recreates the alert under a different
#                     project.
#   is_active       - Required. Whether the alert is active. Updatable in place.
#   component       - Required. The system component the alert covers. "ClusterHealth" is
#                     currently the only supported value.
#   type            - Optional, default "". The alert mechanism: "email" or "http" (or "" to
#                     auto-detect from whichever of identifiers/http you configure).
#   alert_all_users - Optional, default false. If true, sends to every user in the project
#                     instead of just the addresses listed in `identifiers`.
#   identifiers     - Optional. Email addresses to notify (each must be a valid email address).
#                     Only used when `alert_all_users` is false.
#
# `created_by` (string) and `status` (list: is_succeeded/message/time) also exist in the schema,
# but `status` is populated internally by Palette for delivery tracking and isn't meant to be set
# here - leave both unset.
resource "spectrocloud_alert" "alert_email" {
  project         = "Default"
  is_active       = true
  component       = "ClusterHealth"
  type            = "email"
  alert_all_users = true
  identifiers     = ["abc@spectrocloud.com", "cba@spectrocloud.com"]
}

# http block (Required when type = "http", exactly one per alert):
#   method  - Required. One of "POST", "GET", "PUT".
#   url     - Required. The webhook URL to call when the alert fires.
#   body    - Required. The request body sent to the URL above.
#   headers - Optional. Extra headers to send with the request.
resource "spectrocloud_alert" "alert_http" {
  project         = "Default"
  is_active       = true
  component       = "ClusterHealth"
  type            = "http"
  alert_all_users = true

  http {
    method = "POST"
    url    = "https://openhook.com/put/notify"
    body   = "{ \"text\": \"{{message}}\" }"
    headers = {
      tag    = "Health"
      source = "spectrocloud"
    }
  }
}

# Import example:
# terraform import spectrocloud_alert.alert_email "alertUid:ClusterHealth"
#
# Where:
# - alertUid is the unique identifier of the alert
# - ClusterHealth is the component type
