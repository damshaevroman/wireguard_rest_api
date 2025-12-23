package controllers

import (
	"wireguard_api/config"
	"wireguard_api/usecases"
)

type Controller struct {
	service usecases.UsecaseService
	cfg     *config.ServerConfig
}

type addClient struct {
	Ifname    string `json:"ifname" binding:"required"`
	Ip        string `json:"ip"`
	AllowedIp string `json:"alloweip"`
}

type deleteClient struct {
	Public string `json:"public" binding:"required"`
}

type addServer struct {
	Ifname   string `json:"ifname" binding:"required" `
	Ip       string `json:"ip" binding:"required"`
	Endpoint string `json:"endpoint" binding:"required"`
	Port     int    `json:"port" binding:"required"`
}

type deleteServer struct {
	Private string `json:"private" binding:"required"`
	Ifname  string `json:"ifname" binding:"required"`
}

type ServerStartStop struct {
	Ifname string `json:"ifname" binding:"required"`
}

type ServerForward struct {
	Command     string   `json:"command" binding:"required"`
	Source      string   `json:"source" binding:"required"`
	Destination string   `json:"destination"`
	Protocol    string   `json:"protocol" binding:"required,oneof=tcp udp icmp"`
	Position    int      `json:"position" binding:"required,min=1,max=65535"`
	Port        string   `json:"port"`
	Comment     string   `json:"comment" binding:"required,min=1"`
	List        bool     `json:"list"`                                        // if list tru this mean what we recived list of ip address i ndestionation fiels like 192.168.0.1, 10.0.0.2
	IpList      []string `json:"ip_list" binding:"omitempty,dive,ip"`         // used when List is true, contains multiple IPs for destination}
	Action      string   `json:"action" binding:"required,oneof=ACCEPT DROP"` // action to perform on the forward rule
	Except      bool     `json:"except"`
}
type ServerForwardUpdateList struct {
	Command   string   `json:"command" binding:"required"`
	IpsetName string   `json:"ipset_name" binding:"required"`
	Single    bool     `json:"single"`
	IpList    []string `json:"ip_list" binding:"omitempty,dive,ip"` // used when List is true, contains multiple IPs for destination
}

type ServerMasquerade struct {
	Command string `json:"command" binding:"required"`
	Source  string `json:"source" binding:"required"`
	Ifname  string `json:"ifname" binding:"required"`
	Comment string `json:"comment" binding:"required"`
}
