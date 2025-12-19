package repository

import (
	"fmt"
	"net"
	"wireguard_api/db"

	"gorm.io/gorm"
)

type ServerCertRepository struct {
	db *gorm.DB
}

func NewServerCertRepository(db *gorm.DB) *ServerCertRepository {
	return &ServerCertRepository{db: db}
}

func (r *ServerCertRepository) CreateServerCert(cert *db.ServerCert) error {
	return r.db.Create(cert).Error
}

func (r *ServerCertRepository) GetServerCertByIfname(ifname string) (db.ServerCert, error) {
	var cert db.ServerCert
	err := r.db.Where("ifname = ?", ifname).First(&cert).Error
	return cert, err
}

func (r *ServerCertRepository) DeleteServer(private, ifname string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var serverCertExists db.ServerCert
		err := tx.Model(&db.ServerCert{}).Where("ifname = ?", ifname).First(&serverCertExists).Error
		if err != nil {
			return err
		}

		if serverCertExists.Private != private {
			return fmt.Errorf("did not find record with correct ifname: %s  and private %s in database", ifname, private)
		}

		errTx := tx.Exec(`
			INSERT INTO archive_client_certs (created_at, ifname, private, public, ip, allowed_ips, config, deleted_at)
			SELECT created_at, ifname, private, public, ip, allowed_ips, config, DATETIME('now')
			FROM client_certs
			WHERE ifname = ?;`, ifname)
		if errTx.Error != nil {
			return errTx.Error
		}
		errTx = tx.Exec(`
			INSERT INTO archive_server_certs (created_at, ifname, private, public, endpoint, ip, config, port, deleted_at)
			SELECT created_at, ifname, private, public, endpoint, ip, config, port, DATETIME('now')
			FROM server_certs
			WHERE ifname = ?;`, ifname)
		if errTx.Error != nil {
			return errTx.Error
		}
		err = tx.Unscoped().Where("ifname = ?", ifname).Delete(&db.ClientCert{}).Error
		if err != nil {
			return err
		}
		err = tx.Unscoped().Where("private = ? AND ifname = ?", private, ifname).Delete(&db.ServerCert{}).Error
		if err != nil {
			return err
		}

		return nil
	})
}

func (r *ServerCertRepository) GetServerArchive() ([]db.ArchiveServerCert, error) {
	var archive []db.ArchiveServerCert
	err := r.db.Unscoped().Find(&archive).Error
	if err != nil {
		return []db.ArchiveServerCert{}, err
	}
	return archive, nil
}

func (r *ServerCertRepository) GetServerInterfaces() ([]db.ServerCert, error) {
	var serv []db.ServerCert
	err := r.db.Find(&serv).Error
	if err != nil {
		return []db.ServerCert{}, err
	}
	return serv, nil
}

func (r *ServerCertRepository) GetServerCertificates() ([]db.ServerCert, error) {
	var cert []db.ServerCert
	err := r.db.Find(&cert).Error
	if err != nil {
		return []db.ServerCert{}, err
	}
	return cert, nil

}

func (r *ServerCertRepository) isCIDR(s string) error {
	_, _, err := net.ParseCIDR(s)
	return err
}

func (r *ServerCertRepository) CreateForward(position int, port, action, source, destination, protocol, comment string, isList, except bool) error {
	// Проверка формата CIDR
	if err := r.isCIDR(source); err != nil {
		return fmt.Errorf("source: %s is not subnet with cidr example 10.0.0.0/24", source)
	}
	if !isList {
		if err := r.isCIDR(destination); err != nil {
			return fmt.Errorf("destination: %s is not subnet with cidr 10.0.0.0/24", destination)
		}
	}

	// Проверка существования позиции 1
	var first db.Forward
	if err := r.db.Where("position = ?", 1).First(&first).Error; err != nil && position > 1 {
		return fmt.Errorf("don't have rule number 1, set position to 1")
	}

	// Сдвигаем все позиции >= новой позиции
	if err := r.db.Model(&db.Forward{}).
		Where("position >= ?", position).
		Update("position", gorm.Expr("position + 1")).Error; err != nil {
		return fmt.Errorf("failed to shift positions: %v", err)
	}

	// Создаем новую запись
	if err := r.db.Create(&db.Forward{
		Action:      action,
		Position:    position,
		Source:      source,
		Destination: destination,
		Protocol:    protocol,
		Port:        port,
		Comment:     comment,
		IsList:      isList,
		Except:      except,
	}).Error; err != nil {
		return err
	}

	return nil
}

func (r *ServerCertRepository) DeleteForward(comment string) error {
	var deletedForward db.Forward
	err := r.db.Where("comment = ?", comment).First(&deletedForward).Error
	if err != nil {
		return fmt.Errorf("record not found: %w", err)
	}
	result := r.db.Unscoped().Delete(&deletedForward)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no matching record found to delete")
	}
	err = r.db.Model(&db.Forward{}).
		Where("position > ?", deletedForward.Position).
		Update("position", gorm.Expr("position - 1")).Error
	if err != nil {
		return fmt.Errorf("failed to update positions: %w", err)
	}
	return nil
}

func (r *ServerCertRepository) CreateMasquerade(source, ifname, comment string) error {
	var existingCert db.Masquerade
	err := r.db.Where("ifname = ?", ifname).First(&existingCert).Error
	if err == nil {
		return nil
	}
	if err := r.db.Create(&db.Masquerade{Ifname: ifname, Source: source, Comment: comment}).Error; err != nil {
		return err
	}

	return nil
}

func (r *ServerCertRepository) DeleteMasquerade(source, ifname, comment string) error {
	err := r.db.Unscoped().
		Where("ifname = ? AND source = ? AND comment = ?", ifname, source, comment).
		Delete(&db.Masquerade{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *ServerCertRepository) GetMasquerade() ([]db.Masquerade, error) {
	var masq []db.Masquerade
	err := r.db.Unscoped().Order("id ASC").Find(&masq).Error
	if err != nil {
		return []db.Masquerade{}, err
	}
	return masq, nil
}

func (r *ServerCertRepository) GetForward() ([]db.Forward, error) {
	var fwrd []db.Forward
	err := r.db.Unscoped().Order("position ASC").Find(&fwrd).Error
	if err != nil {
		return []db.Forward{}, err
	}
	return fwrd, nil
}
