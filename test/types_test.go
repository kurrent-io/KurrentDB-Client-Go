package test_test

import (
	"testing"
	"time"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
)

func TestTypes(t *testing.T) {
	t.Run("TestConsistentMetadataSerializationStreamAcl", func(t *testing.T) {
		acl := kurrentdb.Acl{}
		acl.AddReadRoles("admin")
		acl.AddWriteRoles("admin")
		acl.AddDeleteRoles("admin")
		acl.AddMetaReadRoles("admin")
		acl.AddMetaWriteRoles("admin")

		expected := kurrentdb.StreamMetadata{}
		expected.SetMaxAge(2 * time.Second)
		expected.SetCacheControl(15 * time.Second)
		expected.SetTruncateBefore(1)
		expected.SetMaxCount(12)
		expected.SetAcl(acl)
		expected.AddCustomProperty("foo", "bar")

		bytes, err := expected.ToJson()
		assert.NoError(t, err, "failed to serialize in JSON")

		meta, err := kurrentdb.StreamMetadataFromJson(bytes)
		assert.NoError(t, err, "failed to parse Metadata from props")
		assert.Equal(t, expected, *meta, "consistency serialization failure")
	})

	t.Run("TestConsistentMetadataSerializationUserStreamAcl", func(t *testing.T) {
		expected := kurrentdb.StreamMetadata{}
		expected.SetMaxAge(2 * time.Second)
		expected.SetCacheControl(15 * time.Second)
		expected.SetTruncateBefore(1)
		expected.SetMaxCount(12)
		expected.SetAcl(kurrentdb.UserStreamAcl)
		expected.AddCustomProperty("foo", "bar")

		bytes, err := expected.ToJson()
		assert.NoError(t, err, "failed to serialize in JSON")

		meta, err := kurrentdb.StreamMetadataFromJson(bytes)
		assert.NoError(t, err, "failed to parse Metadata from props")
		assert.Equal(t, expected, *meta, "consistency serialization failure")
	})

	t.Run("TestConsistentMetadataSerializationSystemStreamAcl", func(t *testing.T) {
		expected := kurrentdb.StreamMetadata{}
		expected.SetMaxAge(2 * time.Second)
		expected.SetCacheControl(15 * time.Second)
		expected.SetTruncateBefore(1)
		expected.SetMaxCount(12)
		expected.SetAcl(kurrentdb.SystemStreamAcl)
		expected.AddCustomProperty("foo", "bar")

		bytes, err := expected.ToJson()
		assert.NoError(t, err, "failed to serialize in JSON")

		meta, err := kurrentdb.StreamMetadataFromJson(bytes)
		assert.NoError(t, err, "failed to parse Metadata from props")
		assert.Equal(t, expected, *meta, "consistency serialization failure")
	})

	t.Run("TestCustomPropertyRetrievalFromStreamMetadata", func(t *testing.T) {
		expected := kurrentdb.StreamMetadata{}
		expected.AddCustomProperty("foo", "bar")

		foo := expected.CustomProperty("foo")
		assert.Equal(t, "bar", foo, "custom property value mismatch")
	})

	t.Run("TestUnknownCustomPropertyRetrievalFromStreamMetadata", func(t *testing.T) {
		expected := kurrentdb.StreamMetadata{}
		expected.AddCustomProperty("foo", "bar")

		foo := expected.CustomProperty("foes")
		assert.Empty(t, foo, "custom property value mismatch")
	})

	t.Run("TestGetAllCustomPropertiesFromStreamMetadata", func(t *testing.T) {
		expected := kurrentdb.StreamMetadata{}
		expected.AddCustomProperty("foo", 123)
		expected.AddCustomProperty("foes", "baz")

		props := expected.CustomProperties()
		assert.Equal(t, map[string]interface{}{"foo": 123, "foes": "baz"}, props, "custom properties mismatch")
	})
}
