# Spectro Cloud credentials
sc_host         = "{Enter Spectro Cloud API Host}" #e.g: api.spectrocloud.com (for SaaS)
sc_api_key      = "{Enter Spectro Cloud API Key}"
sc_project_name = "{Enter Spectro Cloud Project Name}" #e.g: Default
sc_trace        = false

tke_ssh_key_name = "{enter Spectro Cloud ssh key for tke}"
tke_region       = "{enter region name for tke cluster}"
tke_vpc_id       = "{enter tke vpc id}"
cp_tke_subnets_map = {
  "{enter subnet key}" : "{enter subnet id}"
}
worker_tke_subnets_map = {
  "{enter subnet key}" : "{enter subnet id}"
}