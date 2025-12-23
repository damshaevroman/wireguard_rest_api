package db

import (
	"os"
	"path/filepath"
	"testing"
	"wireguard_api/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitAndClose_OK(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.ServerConfig{
		Database: dbPath,
	}
	dbStruct := Init(cfg)
	require.NotNil(t, dbStruct.DbInstance)
	_, err := os.Stat(dbPath)
	assert.NoError(t, err)
	sqlDB, err := dbStruct.DbInstance.DB()
	require.NoError(t, err)
	assert.NoError(t, sqlDB.Ping())
	err = dbStruct.Close()
	assert.NoError(t, err)
	err = sqlDB.Ping()
	assert.Error(t, err)
}
