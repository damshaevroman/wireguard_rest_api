package repository

import (
	"errors"
	"wireguard_api/db"

	"gorm.io/gorm"
)

type ClientCertRepository struct {
	db *gorm.DB
}

func NewClientCertRepository(db *gorm.DB) *ClientCertRepository {
	return &ClientCertRepository{db: db}
}

func (r *ClientCertRepository) CreateClientCert(cert *db.ClientCert) error {
	return r.db.Create(cert).Error
}

func (r *ClientCertRepository) DeleteClientCert(public string) (db.ClientCert, error) {
	var cert db.ClientCert
	var arch db.ArchiveClientCert
	err := r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("public = ?", public).First(&cert).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		} else if err != nil {
			return err
		}
		arch = db.ArchiveClientCert{
			Public:     cert.Public,
			Private:    cert.Private,
			Ifname:     cert.Ifname,
			IP:         cert.IP,
			AllowedIPs: cert.AllowedIPs,
			Config:     cert.Config,
		}
		err = tx.Create(&arch).Error
		if err != nil {
			return err
		}
		err = tx.Unscoped().Where("public = ?", public).Delete(&db.ClientCert{}).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return db.ClientCert{}, err
	}
	return cert, nil
}

func (r *ClientCertRepository) GetClientCertsByIfname(ifname string) ([]db.ClientCert, error) {
	var certs []db.ClientCert
	err := r.db.Where("ifname = ?", ifname).Find(&certs).Error
	return certs, err
}

func (r *ClientCertRepository) GetPublicEnpointPort(ifname string) (db.ServerCert, error) {
	var cert db.ServerCert
	err := r.db.Where("ifname = ?", ifname).Find(&cert).Error
	if err != nil {
		return db.ServerCert{}, err
	}
	return cert, nil
}

func (r *ClientCertRepository) GetListIp(ifname string) ([]string, error) {
	var ips []string
	err := r.db.Model(&db.ClientCert{}).Where("ifname = ?", ifname).Pluck("ip", &ips).Error
	if err != nil {
		return nil, err
	}
	return ips, nil

}

func (r *ClientCertRepository) GetAllClient() ([]db.ClientCert, error) {
	var certs []db.ClientCert
	err := r.db.Find(&certs).Error
	if err != nil {
		return nil, err
	}
	return certs, nil
}

func (r *ClientCertRepository) GetClientArchive() ([]db.ArchiveClientCert, error) {
	var archive []db.ArchiveClientCert
	err := r.db.Unscoped().Find(&archive).Error
	if err != nil {
		return []db.ArchiveClientCert{}, err
	}
	return archive, nil
}
