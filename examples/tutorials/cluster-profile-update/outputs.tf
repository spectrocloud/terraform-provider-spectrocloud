output "advisory" {
  value = <<-EOT
    It takes between one to three minutes for DNS to properly resolve the public load balancer URL.
    We recommend waiting a few moments before clicking on the service URL to prevent the browser from caching an unresolved DNS request.
  EOT
}

output "aws_hello_universe_api_ip" {
  description = "Instructions to retrieve the IP address of the hello-universe-api service deployed to AWS."
  value       = var.deploy-aws ? "Use the following command to get the IP address of hello-universe-api service:\n\nexport KUBECONFIG=$(pwd)/aws-cluster-api.kubeconfig && \\\nkubectl get service hello-universe-api-service --namespace hello-universe-api --output jsonpath='{.status.loadBalancer.ingress[0].hostname}'" : null
}

output "azure_hello_universe_api_ip" {
  description = "Instructions to retrieve the IP address of the hello-universe-api service deployed to Azure."
  value       = var.deploy-azure ? "Use the following command to get the IP address of hello-universe-api service:\n\nexport KUBECONFIG=$(pwd)/azure-cluster-api.kubeconfig && \\\nkubectl get service hello-universe-api-service --namespace hello-universe-api --output jsonpath='{.status.loadBalancer.ingress[0].ip}'" : null
}

output "gcp_hello_universe_api_ip" {
  description = "Instructions to retrieve the IP address of the hello-universe-api service deployed to GCP."
  value       = var.deploy-gcp ? "Use the following command to get the IP address of hello-universe-api service:\n\nexport KUBECONFIG=$(pwd)/gcp-cluster-api.kubeconfig && \\\nkubectl get service hello-universe-api-service --namespace hello-universe-api --output jsonpath='{.status.loadBalancer.ingress[0].ip}'" : null
}
