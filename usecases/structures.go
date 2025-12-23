package usecases

import (
	"time"
)

type Usecases struct {
	ServerRepo ServerRepo
	ClientRepo ClientRepo
	IpTables   IPTables
	PingStatus PingService
}

var _ UsecaseService = (*Usecases)(nil)

type ClientResponsePing struct {
	Status   bool  `json:"status"`
	PintTime int64 `json:"ping_time"`
}

type ClientResponse struct {
	Ifname     string             `json:"ifname"`
	Private    string             `json:"private"`
	Public     string             `json:"public"`
	Ip         string             `json:"ip"`
	AllowedIPs string             `json:"alloweip"`
	Config     string             `json:"config"`
	PingStatus ClientResponsePing `json:"ping_status"`
}

type InterfaceListStatus struct {
	Ifname string   `json:"ifname"`
	Status []Status `json:"status"`
}
type Status struct {
	Public            string      `json:"public"`
	LastHandshakeTime time.Time   `json:"handshake"`
	Recieved          int64       `json:"reciev"`
	Transmit          int64       `json:"trasmit"`
	AllowedIp         []AllowerIP `json:"alloweip"`
	Endpoint          string      `json:"endpoint"`
}

type AllowerIP struct {
	Ip   string `json:"ip"`
	Mask int    `json:"mask"`
}

type ServerInterfaces struct {
	Ifname   string `json:"ifname"`
	Ip       string `json:"ip"`
	Port     int    `json:"port"`
	Private  string `json:"private"`
	Public   string `json:"public"`
	Endpoint string `json:"endpoint"`
	Config   string `json:"config"`
}

type UsForward struct {
	Bytes       string   `json:"bytes"`
	Source      string   `json:"source"`
	Destination string   `json:"destination"`
	Protocol    string   `json:"protocol"`
	Position    int      `json:"position"`
	Action      string   `json:"action"` // action to perform on the forward rule, e.g., allow or deny
	Port        string   `json:"port"`
	Comment     string   `json:"comment"`
	List        bool     `json:"list"`
	IpList      []string `json:"ip_list,omitempty"` // Used when List is true, contains multiple IPs for destination
	Except      bool     `json:"except"`
}

type UsMasquerade struct {
	Ifname  string `json:"ifname"`
	Source  string `json:"source"`
	Comment string `json:"comment"`
}

type IptablesRulesData struct {
	Forward       []UsForward    `json:"forward"`
	Masquerade    []UsMasquerade `json:"masquerade"`
	InterfaceList []string       `json:"interfaces"`
}
