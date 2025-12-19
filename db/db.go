package db

import (
	"log"
	"os"
	"wireguard_api/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Init(cfg *config.ServerConfig) DatabaseStruct {
	dbPath := cfg.Database
	file, err := os.OpenFile(dbPath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatalf("cannot create database: %v", err)
	}
	file.Close()

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("cannot connect to database: %v", err)
	}
	err = db.AutoMigrate(&ServerCert{}, &ClientCert{}, &ArchiveClientCert{}, &ArchiveServerCert{}, Forward{}, Masquerade{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	return DatabaseStruct{DbInstance: db}
}

func (d *DatabaseStruct) Close() error {
	sqlDB, err := d.DbInstance.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
