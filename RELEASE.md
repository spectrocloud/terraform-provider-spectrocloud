# Releasing the Custom Terraform Provider

This guide outlines the steps to release the custom Terraform provider, ensuring it is published on GitHub and verified on the Terraform Registry.

## 1. Prepare for Release

### Ensure Code Quality and Functionality
- Perform thorough testing to ensure all functionalities are working as expected.
- Ensure that the code adheres to best practices and is well-documented.

### Update Release Notes
- Ensure that the release notes reflect release changes. Check with engineering manager if there is a doubt.
- Ensure that latest documentation is merged to the branch. Work with engineering manager to sign off that documentation is latest.

## 2. GitHub Release

### Draft a New Release
- Navigate to the [Releases](https://github.com/spectrocloud/terraform-provider-spectrocloud/releases) section of the GitHub repository.
- Click on "Draft a new release" and fill in the tag version, release title, and a description of the changes. Default process is for example named and tagged as `v0.15.0` from `main` branch. If custom branch is used it should be specified in the release ticket.

### Attach Binary Files
- Binary files will be built by github action. Monitor the progress of the build and attach the binary files to the release once the build is complete.
- If it fails and retry is needed new version should be released.
- Binaries are uploaded to terraform registry automatically see next steps.

## 3. Verify on Terraform Registry

### Ensure New Version is Available
- After publishing the release on GitHub, check the [Terraform Registry page](https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest) to ensure the new version is available. Publish to terraform registry is done automatically.
- It may take some time usually 20 minutes for the new version to be reflected on the Terraform Registry from the previous release published.

### Verify Documentation and Usage Examples
- Ensure that the documentation on the Terraform Registry is accurate and reflects the latest changes. 

### Test the Provider
- Implement the provider in a Terraform configuration and initialize it using `terraform init`.
- Ensure that the provider is fetched from the Terraform Registry and works as expected in a sanity test scenario.

### Publish the Release notification on slack channel.