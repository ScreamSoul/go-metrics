package utils_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/screamsoul/go-metrics-tpl/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestConfig struct {
	Name string
}

func TestFillFromFile_NoConfigPath(t *testing.T) {
	cfg := TestConfig{}

	err := utils.FillFromFile(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "", cfg.Name, "Expected Path to remain empty when no config file is specified")
}

func TestFillFromFile_FileReadError(t *testing.T) {

	require.NoError(t, os.Setenv("CONFIG", "/non/existent/path"))
	defer func() {
		assert.NoError(t, os.Unsetenv("CONFIG"))
	}()

	var cfg map[string]interface{}
	err := json.Unmarshal([]byte(`{"key":"value"}`), &cfg) // Ensure cfg is addressable
	if err != nil {
		t.Fatal(err)
	}

	err = utils.FillFromFile(&cfg)
	require.Error(t, err)
	// Since FillFromFile doesn't return an error, we check if cfg remains unchanged after attempting to fill it from a non-existent file
	assert.Equal(t, map[string]interface{}{"key": "value"}, cfg)
}

func TestFillFromFile_Success(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		assert.NoError(t, os.Remove(tmpfile.Name())) // Clean up
	}()

	configData := []byte(`{"Name":"some name"}`)
	if _, err = tmpfile.Write(configData); err != nil {
		t.Fatal(err)
	}
	if err = tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	require.NoError(t, os.Setenv("CONFIG", tmpfile.Name()))
	defer func() {
		assert.NoError(t, os.Unsetenv("CONFIG"))
	}()

	cfg := TestConfig{}
	err = utils.FillFromFile(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "some name", cfg.Name, "Expected Path to match the value from the temporary config file")
}
