package convert

import (
	"encoding/json"
	"fmt"

	"github.com/spectrocloud/hapi/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

func FromHapiVolume(hapiVolume *models.V1VMAddVolumeEntity) (*cdiv1.DataVolume, error) {
	var GracePeriodSeconds *int64
	if hapiVolume.DataVolumeTemplate.Metadata.DeletionGracePeriodSeconds != 0 {
		GracePeriodSeconds = &hapiVolume.DataVolumeTemplate.Metadata.DeletionGracePeriodSeconds
	}

	Spec, err := FromHapiVolumeSpecM(hapiVolume.DataVolumeTemplate.Spec)
	if err != nil {
		return nil, err
	}

	volume := &cdiv1.DataVolume{
		ObjectMeta: metav1.ObjectMeta{
			Annotations:                hapiVolume.DataVolumeTemplate.Metadata.Annotations,
			DeletionGracePeriodSeconds: GracePeriodSeconds,
			Finalizers:                 hapiVolume.DataVolumeTemplate.Metadata.Finalizers,
			GenerateName:               hapiVolume.DataVolumeTemplate.Metadata.GenerateName,
			Generation:                 hapiVolume.DataVolumeTemplate.Metadata.Generation,
			Labels:                     hapiVolume.DataVolumeTemplate.Metadata.Labels,
			Name:                       hapiVolume.DataVolumeTemplate.Metadata.Name,
			Namespace:                  hapiVolume.DataVolumeTemplate.Metadata.Namespace,
			ResourceVersion:            hapiVolume.DataVolumeTemplate.Metadata.ResourceVersion,
			UID:                        types.UID(hapiVolume.DataVolumeTemplate.Metadata.UID),
		},
		Spec: *Spec,
	}
	return volume, nil
}

func FromHapiVolumeSpecM(hapiVolumeSpec *models.V1VMDataVolumeSpec) (*cdiv1.DataVolumeSpec, error) {
	var spec cdiv1.DataVolumeSpec

	// Marshal the input hapiVolumeSpec to JSON
	specJson, err := json.Marshal(hapiVolumeSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal models.V1VMDataVolumeSpec to JSON: %v", err)
	}

	// Unmarshal the JSON to the desired DataVolumeSpec
	err = json.Unmarshal(specJson, &spec)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to cdiv1.DataVolumeSpec: %v", err)
	}

	return &spec, nil
}
