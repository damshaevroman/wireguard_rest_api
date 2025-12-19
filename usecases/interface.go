package usecases

import (
	"sync"
	"time"
	"wireguard_api/db"
)

type ServerRepo interface {
	CreateServerCert(cert *db.ServerCert) error
	GetServerCertByIfname(ifname string) (db.ServerCert, error)
	DeleteServer(private, ifname string) error
	GetServerArchive() ([]db.ArchiveServerCert, error)
	GetServerInterfaces() ([]db.ServerCert, error)
	GetServerCertificates() ([]db.ServerCert, error)
	CreateForward(
		position int,
		port string,
		action string,
		source string,
		destination string,
		protocol string,
		comment string,
		isList bool,
		except bool,
	) error

	DeleteForward(comment string) error
	GetForward() ([]db.Forward, error)

	CreateMasquerade(source, ifname, comment string) error
	DeleteMasquerade(source, ifname, comment string) error
	GetMasquerade() ([]db.Masquerade, error)
}

type ClientRepo interface {
	GetPublicEnpointPort(ifname string) (db.ServerCert, error)
	GetListIp(ifname string) ([]string, error)
	CreateClientCert(cert *db.ClientCert) error
	GetAllClient() ([]db.ClientCert, error)
	DeleteClientCert(public string) (db.ClientCert, error)
	GetClientArchive() ([]db.ArchiveClientCert, error)
	GetClientCertsByIfname(ifname string) ([]db.ClientCert, error)
}

type IPTables interface {
	SetForwardList(
		position int,
		port, action, command, source, destination, protocol, comment string,
		except bool,
	) error

	SetForward(
		position int,
		port, action, command, source, destination, protocol, comment string,
		except bool,
	) error

	SetMasquerade(command, subnet, ifname, comment string) error

	GetMasqueradeList() ([]string, error)
	GetForwardList() ([]string, error)

	FlushForward() error
}

type PingService interface {
	Ping(target string, wg *sync.WaitGroup)
	Read(ip string) (bool, time.Duration)
	Delete(ip string)
}

type Service interface {
	GetStatus() ([]InterfaceListStatus, error)

	GetAllClients() ([]ClientResponse, error)
	NewClient(ifname, ip, allowed string) (ClientResponse, error)
	DeleteClient(public string) error
	GetClientArchive() ([]ClientResponse, error)

	NewInterface(ifname, ip, endpoint string, port int) (ServerInterfaces, error)
	DeleteServer(private, ifname string) error
	StartInterface(ifname string) error
	StopInterface(ifname string) error
	GetServerArchive() ([]ServerInterfaces, error)
	GetServerInterfaces() ([]ServerInterfaces, error)

	SetUsForward(
		position int,
		action, command, source, destination, protocol, port, comment string,
		isList, except bool,
	) error

	UpdateIpSetList(command, name string, ipList []string, single bool) error
	SetUsMasquerade(command, source, ifname, comment string) error
	GetIptablesRules() (IptablesRulesData, error)
}
