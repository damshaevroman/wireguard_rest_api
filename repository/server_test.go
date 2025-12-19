package repository

import (
	"testing"
	dbtest "wireguard_api/db"

	"github.com/stretchr/testify/assert"
)

func TestCreateServerCert(t *testing.T) {
	db := setupTestDB()
	repo := NewServerCertRepository(db)
	serv := &dbtest.ServerCert{
		Public:   "test-public",
		Private:  "test-private",
		Ifname:   "test-ifname",
		Endpoint: "10.0.0.1",
		Ip:       "192.168.1.2/24",
		Config:   "test-config",
		Port:     1000,
	}
	err := repo.CreateServerCert(serv)
	assert.NoError(t, err)
	var cert dbtest.ServerCert
	err = db.First(&cert, "private = ?", "test-private").Error
	assert.NoError(t, err)
	assert.Equal(t, cert.Public, serv.Public)
	assert.Equal(t, cert.Private, serv.Private)
	assert.Equal(t, cert.Ifname, serv.Ifname)
	assert.Equal(t, cert.Endpoint, serv.Endpoint)
	assert.Equal(t, cert.Ip, serv.Ip)
	assert.Equal(t, cert.Config, serv.Config)
	assert.Equal(t, cert.Port, serv.Port)
}

func TestDeleteServerCert(t *testing.T) {
	db := setupTestDB()

	repoClient := NewClientCertRepository(db)
	repoServ := NewServerCertRepository(db)

	clientCert1 := &dbtest.ClientCert{
		Public:     "test-public",
		Private:    "test-private",
		Ifname:     "server-ifname",
		IP:         "192.168.1.10",
		AllowedIPs: "192.168.1.0/24",
		Config:     "test-config",
	}

	clientCert2 := &dbtest.ClientCert{
		Public:     "test-public",
		Private:    "test-private",
		Ifname:     "server-ifname",
		IP:         "192.168.1.11",
		AllowedIPs: "192.168.1.0/24",
		Config:     "test-config",
	}

	serv := &dbtest.ServerCert{
		Public:   "test-public",
		Private:  "test-private",
		Ifname:   "server-ifname",
		Endpoint: "10.0.0.1",
		Ip:       "192.168.1.2/24",
		Config:   "test-config",
		Port:     1000,
	}
	err := repoServ.CreateServerCert(serv)
	assert.NoError(t, err)

	err = repoClient.CreateClientCert(clientCert1)
	assert.NoError(t, err)
	err = repoClient.CreateClientCert(clientCert2)
	assert.NoError(t, err)

	err = repoServ.DeleteServer(serv.Private, serv.Ifname)
	assert.NoError(t, err)

	var cert dbtest.ServerCert
	err = db.First(&cert, "private = ?", serv.Private).Error
	assert.Equal(t, err.Error(), "record not found")

	var sArchive dbtest.ArchiveServerCert
	err = db.First(&sArchive, "private = ?", serv.Private).Error
	assert.Error(t, err)
	assert.Equal(t, sArchive.Ifname, cert.Ifname)
	assert.Equal(t, sArchive.Private, cert.Private)
	certs, err := repoClient.GetClientCertsByIfname(cert.Ifname)
	assert.NoError(t, err)
	assert.Len(t, certs, 0)

	aCerts, err := repoClient.GetClientArchive()
	assert.NoError(t, err)
	assert.Len(t, aCerts, 2)

	aServ, err := repoServ.GetServerArchive()
	assert.NoError(t, err)
	assert.Len(t, aServ, 1)
}
