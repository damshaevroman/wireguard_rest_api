package db

import "gorm.io/gorm"

type DatabaseStruct struct {
	DbInstance *gorm.DB
}

type ServerCert struct {
	gorm.Model
	Private  string `gorm:"not null"`
	Public   string `gorm:"not null"`
	Endpoint string `gorm:"not null"`
	Ip       string `gorm:"unique;not null"`
	Config   string `gorm:"not null"`
	Ifname   string `gorm:"unique;not null"`
	Port     int    `gorm:"unique;not null"`
}

type ClientCert struct {
	gorm.Model
	Ifname     string `gorm:"not null"`
	Private    string `gorm:"not null"`
	Public     string `gorm:"not null"`
	IP         string `gorm:"unique;not null"`
	AllowedIPs string
	Config     string `gorm:"not null"`
}

type ArchiveClientCert struct {
	gorm.Model
	Ifname     string
	Private    string
	Public     string
	IP         string
	AllowedIPs string
	Config     string
}

type ArchiveServerCert struct {
	gorm.Model
	Private  string
	Public   string
	Endpoint string
	Ip       string
	Config   string
	Ifname   string
	Port     int
}

type Forward struct {
	gorm.Model
	Source      string `gorm:"not null"`
	Destination string `gorm:"not null"`
	Protocol    string `gorm:"not null"`
	Position    int    `gorm:"unique"`
	Port        string
	Action      string `gorm:"not null"` // action to perform on the forward rule, e.g., allow or deny
	Comment     string `gorm:"unique;not null"`
	IsList      bool
	Except      bool
}

type Masquerade struct {
	gorm.Model
	Source  string `gorm:"not null"`
	Ifname  string `gorm:"not null"`
	Comment string `gorm:"not null"`
}
