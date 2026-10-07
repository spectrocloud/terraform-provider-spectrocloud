output "advisory" {
  value = <<-EOT
    It takes between one to three minutes for DNS to properly resolve the public load balancer URL.
    We recommend waiting a few minutes before clicking on the service URL to prevent the browser from caching an unresolved DNS request.
  EOT
}

#######################
# VMware SSH Key Output
#######################

output "ssh_key_location" {
  description = "Location of the generated private SSH key file."
  value       = length(tls_private_key.tutorial_ssh_key) > 0 && var.deploy-vmware == true ? "This is the location of the generated private SSH key file: ${local_sensitive_file.private_key_file[0].filename}." : null
}

output "ssh_public_key_location" {
  description = "Location of the generated public SSH key file."
  value       = length(tls_private_key.tutorial_ssh_key) > 0 && var.deploy-vmware == true ? "This is the location of the generated public SSH key file: ${local_file.public_key_file[0].filename}." : null
}

output "ssh_connection_command" {
  description = "Command to use the generated private SSH key to access the nodes."
  value       = length(tls_private_key.tutorial_ssh_key) > 0 && var.deploy-vmware == true ? "To access your nodes, use the following command, replacing <username> with the username and <hostname> with the IP address of your node:  ssh -i ${local_sensitive_file.private_key_file[0].filename} <username>@<hostname>" : null
}

output "ssh_connection_command_user" {
  description = "Command to use the user's private SSH key to access the nodes."
  value       = var.ssh_key != "" && var.deploy-vmware == true ? "To access your nodes, use the following command, replacing <username> with the username and <hostname> with the IP address of your node:  ssh -i ${var.ssh_key_private} <username>@<hostname>" : null
}
