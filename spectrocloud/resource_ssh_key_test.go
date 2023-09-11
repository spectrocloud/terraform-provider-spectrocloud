package spectrocloud

import (
	"testing"
)

func TestToSSHKey(t *testing.T) {
	// Create a sample ResourceData object
	d := resourceSSHKey().TestResourceData()
	name := "testSSHName"
	sshKey := "ssh-rsa AAAA6IEQhI1QLiicHLO5a== teerf2021"
	d.Set("name", name)
	d.Set("ssh_key", sshKey)

	result := toSSHKey(d)

	if result.Metadata.Name != name {
		t.Errorf("Expected Metadata Name to be %s, but got %s", name, result.Metadata.Name)
	}

	if result.Spec.PublicKey != sshKey {
		t.Errorf("Expected Spec PublicKey to be %s, but got %s", sshKey, result.Spec.PublicKey)
	}
}
