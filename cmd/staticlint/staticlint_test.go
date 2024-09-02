// NewStaticcheckConfig returns nil or empty analyzer list
package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain_NewStaticcheckConfigReturnsEmpty(t *testing.T) {
	config := NewStaticcheckConfig()
	analyzers := config.GetAnalizers()

	require.NotNil(t, analyzers)
	require.Equal(t, true, len(analyzers) > 0)

}

func TestGetCheckers(t *testing.T) {
	checks := getCheckers()
	assert.Equal(t, true, len(checks) >= 100)
}
