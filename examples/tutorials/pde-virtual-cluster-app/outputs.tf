output "advisory" {
  value = <<-EOT
    It takes between one to three minutes for DNS to properly resolve the public load balancer URL.
    We recommend waiting a few moments before clicking on the service URL to prevent the browser from caching an unresolved DNS request.
  EOT
}
