# Basic Cluster Profile demo

Simple template to create a cluster profile.

Run with a command like this:

```
terraform apply \
   -var 'sc_username={your_spectro_cloud_username}' \
   -var 'sc_password={your_spectro_cloud_password}'
```

Alternatively to using `-var` with each command, the `terraform.template.tfvars` file can be copied to `terraform.tfvars` and updated.
