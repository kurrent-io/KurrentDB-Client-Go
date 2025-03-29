package test_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUUIDParsingSuite(t *testing.T) {
	t.Run("UUIDParsingTests", func(t *testing.T) {
		expected := uuid.New()
		most, least := kurrentdb.UUIDAsInt64(expected)
		actual, err := kurrentdb.ParseUUIDFromInt64(most, least)

		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}
