package repository

import (
	"testing"
	dbtest "wireguard_api/db"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB() *gorm.DB {
	var err error
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	err = db.AutoMigrate(&dbtest.ClientCert{}, &dbtest.ServerCert{}, &dbtest.ArchiveClientCert{}, &dbtest.ArchiveServerCert{})
	if err != nil {
		panic("Failed to migrate database: " + err.Error())
	}

	return db
}
func TestCreateClientCert(t *testing.T) {
	db := setupTestDB()
	repo := NewClientCertRepository(db)

	clientCert := &dbtest.ClientCert{
		Public:     "public_key",
		Private:    "private_key",
		Ifname:     "ifname",
		IP:         "192.168.1.12",
		AllowedIPs: "192.168.1.0/24",
		Config:     "test-config",
	}

	err := repo.CreateClientCert(clientCert)
	assert.NoError(t, err)

	var cert dbtest.ClientCert
	err = db.First(&cert, "public = ?", "public_key").Error
	assert.NoError(t, err)

	assert.NotEmpty(t, cert.Public, "Public key should not be empty")
	assert.NotEmpty(t, cert.Private, "Private key should not be empty")
	assert.NotEmpty(t, cert.Ifname, "Interface name should not be empty")
	assert.NotEmpty(t, cert.IP, "IP address should not be empty")
	assert.NotEmpty(t, cert.AllowedIPs, "Allowed IPs should not be empty")
	assert.NotEmpty(t, cert.Config, "Config should not be empty")

	assert.Equal(t, "public_key", cert.Public)
	assert.Equal(t, "private_key", cert.Private)
	assert.Equal(t, "ifname", cert.Ifname)
	assert.Equal(t, "192.168.1.12", cert.IP)
	assert.Equal(t, "192.168.1.0/24", cert.AllowedIPs)
	assert.Equal(t, "test-config", cert.Config)
}
func TestDeleteClientCert(t *testing.T) {
	db := setupTestDB()
	repo := NewClientCertRepository(db)

	clientCert := &dbtest.ClientCert{
		Public:     "public_key",
		Private:    "private_key",
		Ifname:     "ifname",
		IP:         "192.168.1.12",
		AllowedIPs: "192.168.1.0/24",
		Config:     "test-config",
	}

	err := repo.CreateClientCert(clientCert)
	assert.NoError(t, err)

	client, err := repo.DeleteClientCert("public_key")
	assert.Equal(t, client.Ifname, "ifname")
	assert.NoError(t, err)

	var cert dbtest.ClientCert
	err = db.First(&cert, "public = ?", "public_key").Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	data, err := repo.GetClientArchive()
	assert.NoError(t, err)
	assert.Len(t, data, 1)

	archivedCert := data[0]
	assert.Equal(t, "public_key", archivedCert.Public)
	assert.Equal(t, "private_key", archivedCert.Private)
	assert.Equal(t, "ifname", archivedCert.Ifname)
	assert.Equal(t, "192.168.1.12", archivedCert.IP)
	assert.Equal(t, "192.168.1.0/24", archivedCert.AllowedIPs)
	assert.Equal(t, "test-config", archivedCert.Config)
}
func TestGetClientCertsByIfname(t *testing.T) {
	db := setupTestDB()
	repo := NewClientCertRepository(db)

	clientCert1 := &dbtest.ClientCert{
		Public: "test-public-1",
		Ifname: "test-ifname",
		IP:     "192.168.1.1",
	}
	clientCert2 := &dbtest.ClientCert{
		Public: "test-public-2",
		Ifname: "test-ifname",
		IP:     "192.168.1.2",
	}
	db.Create(clientCert1)
	db.Create(clientCert2)

	certs, err := repo.GetClientCertsByIfname("test-ifname")
	assert.NoError(t, err)
	assert.Len(t, certs, 2)
}
func TestGetListIp(t *testing.T) {
	db := setupTestDB()
	repo := NewClientCertRepository(db)

	clientCert1 := &dbtest.ClientCert{Public: "test-public-1", Ifname: "test-ifname", IP: "192.168.1.1"}
	clientCert2 := &dbtest.ClientCert{Public: "test-public-2", Ifname: "test-ifname", IP: "192.168.1.2"}
	db.Create(clientCert1)
	db.Create(clientCert2)

	ips, err := repo.GetListIp("test-ifname")
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"192.168.1.1", "192.168.1.2"}, ips)
}
func TestGetAllClient(t *testing.T) {
	db := setupTestDB()
	repo := NewClientCertRepository(db)

	clientCert1 := &dbtest.ClientCert{Public: "test-public-1", Ifname: "test-ifname", IP: "192.168.1.1"}
	clientCert2 := &dbtest.ClientCert{Public: "test-public-2", Ifname: "test-ifname", IP: "192.168.1.2"}
	db.Create(clientCert1)
	db.Create(clientCert2)

	certs, err := repo.GetAllClient()
	assert.NoError(t, err)
	assert.Len(t, certs, 2)

	ips := []string{"192.168.1.1", "192.168.1.2"}
	for _, ip := range ips {
		found := false
		for _, cert := range certs {
			if cert.IP == ip {
				found = true
				break
			}
		}
		assert.True(t, found, "IP address %s not found in the result set", ip)
	}
}
func TestGetClientArchive(t *testing.T) {
	db := setupTestDB()
	repo := NewClientCertRepository(db)

	archivedCert := &dbtest.ArchiveClientCert{
		Public: "test-archived-public",
		Ifname: "test-ifname",
		IP:     "192.168.1.1",
	}
	db.Create(archivedCert)

	archive, err := repo.GetClientArchive()
	assert.NoError(t, err)
	assert.Len(t, archive, 1)
	assert.Equal(t, "test-archived-public", archive[0].Public)
}
