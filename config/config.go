package config

import (
	"fmt"

	"gopkg.in/ini.v1"
)

// var Server ServerConfig
var Version = "1.0.0"

type ServerConfig struct {
	IpPort            string   `ini:"ip_port"`
	TlsPrivate        string   `ini:"tls_private"`
	TlsPublic         string   `ini:"tls_public"`
	Database          string   `ini:"database"`
	Token             string   `ini:"token"`
	DeleteInterface   bool     `ini:"delete_interface"`
	ClientDelete      bool     `ini:"delete_client"`
	WhiteListIpAccess []string `ini:"whitelist_ip_access"`
}

func LoadConfig(path string) (*ServerConfig, error) {
	cfg := &ServerConfig{}

	iniFile, err := ini.Load(path)
	if err != nil {
		return nil, err
	}

	err = iniFile.Section("Server").MapTo(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Token == "" {
		return nil, fmt.Errorf("empty token — please check config")
	}

	return cfg, nil
}
