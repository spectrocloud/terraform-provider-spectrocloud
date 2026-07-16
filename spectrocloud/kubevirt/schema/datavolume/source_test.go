package datavolume

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// source_test.go — Batch 11. Covers the three exp* source helpers.

func TestExpandDataVolumeSourceHTTP(t *testing.T) {
	assert.Nil(t, expandDataVolumeSourceHTTP(nil))
	assert.Nil(t, expandDataVolumeSourceHTTP([]interface{}{nil}))

	got := expandDataVolumeSourceHTTP([]interface{}{
		map[string]interface{}{
			"url":             "https://example.com/img",
			"secret_ref":      "s1",
			"cert_config_map": "cm1",
		},
	})
	require.NotNil(t, got)
	require.NotNil(t, got.URL)
	assert.Equal(t, "https://example.com/img", *got.URL)
	assert.Equal(t, "s1", got.SecretRef)
	assert.Equal(t, "cm1", got.CertConfigMap)
}

func TestExpandDataVolumeSourcePVC(t *testing.T) {
	assert.Nil(t, expandDataVolumeSourcePVC(nil))
	assert.Nil(t, expandDataVolumeSourcePVC([]interface{}{nil}))

	got := expandDataVolumeSourcePVC([]interface{}{
		map[string]interface{}{"namespace": "src-ns", "name": "src-pvc"},
	})
	require.NotNil(t, got)
	require.NotNil(t, got.Namespace)
	require.NotNil(t, got.Name)
	assert.Equal(t, "src-ns", *got.Namespace)
	assert.Equal(t, "src-pvc", *got.Name)
}

func TestExpandDataVolumeSourceRegistry(t *testing.T) {
	assert.Nil(t, expandDataVolumeSourceRegistry(nil))
	assert.Nil(t, expandDataVolumeSourceRegistry([]interface{}{nil}))

	got := expandDataVolumeSourceRegistry([]interface{}{
		map[string]interface{}{"image_url": "docker://registry/img"},
	})
	require.NotNil(t, got)
	assert.Equal(t, "docker://registry/img", got.URL)
}
