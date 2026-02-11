package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceAccountGcpImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	err := GetCommonAccount(d, c, "gcp")
	if err != nil {
		return nil, err
	}

	diags := resourceCloudAccountGcpRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read cluster for import: %v", diags)
	}

	// Return the resource data. In most cases, this method is only used to
	// import one resource at a time, so you should return the resource data
	// in a slice with a single element.
	return []*schema.ResourceData{d}, nil
}

func GetCommonAccount(d *schema.ResourceData, c *client.V1Client, accountType string) error {
	// Parse resource ID and scope (import format: "id_or_name:scope" e.g. "my-account:project")
	scope, accountID, err := ParseResourceID(d)
	if err != nil {
		return err
	}

	// Try by UID first, then by name when accountType is set
	cluster, err := c.GetCloudAccount(accountID)
	if err != nil && accountType != "" {
		uid, resolveErr := resolveAccountByName(c, accountType, accountID, scope)
		if resolveErr != nil {
			return fmt.Errorf("unable to retrieve cloud account by id or name: %s", resolveErr)
		}
		accountID = uid
		cluster, err = c.GetCloudAccount(accountID)
	}
	if err != nil {
		return fmt.Errorf("unable to retrieve cluster data: %s", err)
	}

	err = d.Set("name", cluster.Metadata.Name)
	if err != nil {
		return err
	}
	if cluster.Metadata.Annotations != nil {
		if scope != cluster.Metadata.Annotations["scope"] {
			return fmt.Errorf("CloudAccount scope mismatch: %s != %s", scope, cluster.Metadata.Annotations["scope"])
		}
		err = d.Set("context", cluster.Metadata.Annotations["scope"])
		if err != nil {
			return err
		}
	}
	d.SetId(accountID)
	return nil
}

// resolveAccountByName finds a cloud account by name and scope using the type-specific list API.
func resolveAccountByName(c *client.V1Client, accountType, name, scope string) (string, error) {
	matchScope := func(annotations map[string]string) bool {
		return scope == "" || (annotations != nil && annotations["scope"] == scope)
	}
	switch accountType {
	case "gcp":
		accounts, err := c.GetCloudAccountsGcp()
		if err != nil {
			return "", err
		}
		for _, a := range accounts {
			if a.Metadata.Name == name && matchScope(a.Metadata.Annotations) {
				return a.Metadata.UID, nil
			}
		}
		return "", fmt.Errorf("no GCP cloud account found with name %q in scope %q", name, scope)

	case "aws":
		accounts, err := c.GetCloudAccountsAws()
		if err != nil {
			return "", err
		}
		for _, a := range accounts {
			if a.Metadata.Name == name && matchScope(a.Metadata.Annotations) {
				return a.Metadata.UID, nil
			}
		}
		return "", fmt.Errorf("no AWS cloud account found with name %q in scope %q", name, scope)

	case "azure":
		accounts, err := c.GetCloudAccountsAzure()
		if err != nil {
			return "", err
		}
		for _, a := range accounts {
			if a.Metadata.Name == name && matchScope(a.Metadata.Annotations) {
				return a.Metadata.UID, nil
			}
		}
		return "", fmt.Errorf("no Azure cloud account found with name %q in scope %q", name, scope)

	case "openstack":
		accounts, err := c.GetCloudAccountsOpenStack()
		if err != nil {
			return "", err
		}
		for _, a := range accounts {
			if a.Metadata.Name == name && matchScope(a.Metadata.Annotations) {
				return a.Metadata.UID, nil
			}
		}
		return "", fmt.Errorf("no OpenStack cloud account found with name %q in scope %q", name, scope)

	case "vsphere":
		accounts, err := c.GetCloudAccountsVsphere()
		if err != nil {
			return "", err
		}
		for _, a := range accounts {
			if a.Metadata.Name == name && matchScope(a.Metadata.Annotations) {
				return a.Metadata.UID, nil
			}
		}
		return "", fmt.Errorf("no vSphere cloud account found with name %q in scope %q", name, scope)

	case "maas":
		accounts, err := c.GetCloudAccountsMaas()
		if err != nil {
			return "", err
		}
		for _, a := range accounts {
			if a.Metadata.Name == name && matchScope(a.Metadata.Annotations) {
				return a.Metadata.UID, nil
			}
		}
		return "", fmt.Errorf("no MAAS cloud account found with name %q in scope %q", name, scope)

	case "cloudstack":
		accounts, err := c.GetCloudAccountsCloudStack()
		if err != nil {
			return "", err
		}
		for _, a := range accounts {
			if a.Metadata.Name == name && matchScope(a.Metadata.Annotations) {
				return a.Metadata.UID, nil
			}
		}
		return "", fmt.Errorf("no Apache CloudStack cloud account found with name %q in scope %q", name, scope)

	default:
		return "", fmt.Errorf("import by name not supported for account type %q", accountType)
	}
}
