package spectrocloud

import (
	"errors"
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/apiutil/transport"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

// TestIsForbiddenErr covers isForbiddenErr's classification of API errors,
// added for PLT-2355 (admin_kube_config reads must tolerate a caller that
// lacks cluster.adminKubeconfigDownload, but must still fail hard on any
// other error).
func TestIsForbiddenErr(t *testing.T) {
	t.Run("OperationForbidden", func(t *testing.T) {
		err := &transport.TransportError{
			HttpCode: 403,
			Payload:  &models.V1Error{Code: "OperationForbidden"},
		}
		assert.True(t, isForbiddenErr(err))
	})

	t.Run("ResOperationForbidden", func(t *testing.T) {
		err := &transport.TransportError{
			HttpCode: 403,
			Payload:  &models.V1Error{Code: "ResOperationForbidden"},
		}
		assert.True(t, isForbiddenErr(err))
	})

	t.Run("unrelated error code does not match", func(t *testing.T) {
		err := &transport.TransportError{
			HttpCode: 404,
			Payload:  &models.V1Error{Code: "ResourceNotFound"},
		}
		assert.False(t, isForbiddenErr(err))
	})

	t.Run("plain go error does not match", func(t *testing.T) {
		assert.False(t, isForbiddenErr(errors.New("boom")))
	})

	t.Run("nil error does not match", func(t *testing.T) {
		assert.False(t, isForbiddenErr(nil))
	})
}
