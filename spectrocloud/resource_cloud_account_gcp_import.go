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
// resolveAccountByName finds a cloud account by name and scope using the SDK Get*ByName helpers.
func resolveAccountByName(c *client.V1Client, accountType, name, scope string) (string, error) {
	switch accountType {
	case "gcp":
		acc, err := c.GetCloudAccountGcpByName(name, scope)
		if err != nil {
			return "", err
		}
		if acc != nil && acc.Metadata != nil {
			return acc.Metadata.UID, nil
		}
	case "aws":
		acc, err := c.GetCloudAccountAwsByName(name, scope)
		if err != nil {
			return "", err
		}
		if acc != nil && acc.Metadata != nil {
			return acc.Metadata.UID, nil
		}
	case "azure":
		acc, err := c.GetCloudAccountAzureByName(name, scope)
		if err != nil {
			return "", err
		}
		if acc != nil && acc.Metadata != nil {
			return acc.Metadata.UID, nil
		}
	case "openstack":
		acc, err := c.GetCloudAccountOpenStackByName(name, scope)
		if err != nil {
			return "", err
		}
		if acc != nil && acc.Metadata != nil {
			return acc.Metadata.UID, nil
		}
	case "vsphere":
		acc, err := c.GetCloudAccountVsphereByName(name, scope)
		if err != nil {
			return "", err
		}
		if acc != nil && acc.Metadata != nil {
			return acc.Metadata.UID, nil
		}
	case "maas":
		acc, err := c.GetCloudAccountMaasByName(name, scope)
		if err != nil {
			return "", err
		}
		if acc != nil && acc.Metadata != nil {
			return acc.Metadata.UID, nil
		}
	case "apache-cloudstack":
		acc, err := c.GetCloudAccountCloudStackByName(name, scope)
		if err != nil {
			return "", err
		}
		if acc != nil && acc.Metadata != nil {
			return acc.Metadata.UID, nil
		}
	default:
		// Custom cloud: accountType is the custom cloud type (e.g. "nutanix")
		acc, err := c.GetCloudAccountCustomByName(accountType, name, scope)
		if err != nil {
			return "", fmt.Errorf("import by name not supported for account type %q: %w", accountType, err)
		}
		if acc != nil && acc.Metadata != nil {
			return acc.Metadata.UID, nil
		}
	}
	return "", fmt.Errorf("no cloud account found with name %q in scope %q", name, scope)
}
