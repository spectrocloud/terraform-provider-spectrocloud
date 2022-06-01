package herr

import "github.com/spectrocloud/hapi/apiutil"

func IsNotFound(err error) bool {
	return apiutil.ToV1ErrorObj(err).Code == "ResourceNotFound"
}

func IsEdgeHostDeviceNotRegistered(err error) bool {
	return apiutil.ToV1ErrorObj(err).Code == "EdgeHostDeviceNotRegistered"
}

func IsBackupNotConfigured(err error) bool {
	return apiutil.ToV1ErrorObj(err).Code == "BackupNotConfigured"
}
