package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_FileNotFound(t *testing.T) {
	cfg, err := LoadConfig("/path/does/not/exist.ini")

	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load config")
}

func TestLoadConfig_EmptyToken(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config-*.ini")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	configData := `
ip_port = 0.0.0.0:8080
database = test.db
`

	_, err = tmpFile.WriteString(configData)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	cfg, err := LoadConfig(tmpFile.Name())

	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty token")
}

func TestLoadConfig_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config-*.ini")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	configData := `
ip_port = 127.0.0.1:8080
tls_private = private.pem
tls_public = public.pem
database = test.db
token = super-secret
delete_interface = true
delete_client = false
whitelist_ip_access = 127.0.0.1,10.0.0.1
`

	_, err = tmpFile.WriteString(configData)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	cfg, err := LoadConfig(tmpFile.Name())

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "127.0.0.1:8080", cfg.IpPort)
	assert.Equal(t, "private.pem", cfg.TlsPrivate)
	assert.Equal(t, "public.pem", cfg.TlsPublic)
	assert.Equal(t, "test.db", cfg.Database)
	assert.Equal(t, "super-secret", cfg.Token)
	assert.True(t, cfg.DeleteInterface)
	assert.False(t, cfg.ClientDelete)
	assert.Equal(t, []string{"127.0.0.1", "10.0.0.1"}, cfg.WhiteListIpAccess)
}
