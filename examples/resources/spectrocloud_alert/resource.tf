# Note: Palette allows a maximum of two alerts per project for the "ClusterHealth" component -
# one email alert and one HTTP webhook alert.

resource "spectrocloud_alert" "alert_email" {
  # Required, ForceNew. Changing this recreates the alert under a different project.
  project = "Default"
  # Required. Whether the alert is active. Updatable in place.
  is_active = true
  # Required. The system component the alert covers. "ClusterHealth" is currently the only
  # supported value.
  component = "ClusterHealth"
  # Optional, default "". The alert mechanism: "email" or "http" (or "" to auto-detect from
  # whichever of identifiers/http you configure).
  type = "email"
  # Optional, default false. If true, sends to every user in the project instead of just the
  # addresses listed in `identifiers`.
  alert_all_users = true
  # Optional. Email addresses to notify (each must be a valid email address). Only used when
  # `alert_all_users` is false.
  identifiers = ["abc@spectrocloud.com", "cba@spectrocloud.com"]

  # `created_by` (string) and `status` (list: is_succeeded/message/time) also exist in the
  # schema, but `status` is populated internally by Palette for delivery tracking and isn't
  # meant to be set here - leave both unset.
}

resource "spectrocloud_alert" "alert_http" {
  project         = "Default"
  is_active       = true
  component       = "ClusterHealth"
  type            = "http"
  alert_all_users = true

  # Required block when type = "http". Exactly one http block per alert.
  http {
    # Required. One of "POST", "GET", "PUT".
    method = "POST"
    # Required. The webhook URL to call when the alert fires.
    url = "https://openhook.com/put/notify"
    # Required. The request body sent to the URL above.
    body = "{ \"text\": \"{{message}}\" }"
    # Optional. Extra headers to send with the request.
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
