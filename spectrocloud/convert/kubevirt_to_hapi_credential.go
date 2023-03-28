package convert

import (
	"github.com/spectrocloud/hapi/models"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func ToHapiVmAccessCredentials(credentials []kubevirtapiv1.AccessCredential) []*models.V1VMAccessCredential {
	ret := make([]*models.V1VMAccessCredential, len(credentials))
	for i, credential := range credentials {
		var secretName *string
		if credential.SSHPublicKey.Source.Secret != nil {
			secretName = types.Ptr(credential.SSHPublicKey.Source.Secret.SecretName)
		}

		var Users []string
		if credential.SSHPublicKey.PropagationMethod.QemuGuestAgent != nil {
			Users = credential.SSHPublicKey.PropagationMethod.QemuGuestAgent.Users
		}

		ret[i] = &models.V1VMAccessCredential{
			SSHPublicKey: &models.V1VMSSHPublicKeyAccessCredential{
				Source: &models.V1VMSSHPublicKeyAccessCredentialSource{
					Secret: &models.V1VMAccessCredentialSecretSource{
						SecretName: secretName,
					},
				},
				PropagationMethod: &models.V1VMSSHPublicKeyAccessCredentialPropagationMethod{
					QemuGuestAgent: &models.V1VMQemuGuestAgentSSHPublicKeyAccessCredentialPropagation{
						Users: Users,
					},
				},
			},
		}
	}

	return ret
}
