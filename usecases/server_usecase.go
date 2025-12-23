package usecases

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"
	"wireguard_api/db"
	"wireguard_api/wg"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func (u *Usecases) NewInterface(ifname, ip, endpoint string, port int) (ServerInterfaces, error) {
	ifname = strings.ToLower(strings.TrimSpace(ifname))
	ip = strings.TrimSpace(ip)
	endpoint = strings.TrimSpace(endpoint)
	ifaceList := u.getInterfaceList()
	for _, v := range ifaceList {
		if ifname == v {
			return ServerInterfaces{}, fmt.Errorf("interface %s already exist", ifname)
		}
	}

	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		log.Printf("NewInterface %v", err)
		return ServerInterfaces{}, err
	}
	publicKey := privateKey.PublicKey()
	_, _, err = net.ParseCIDR(ip)
	if err != nil {
		log.Printf("NewInterface %v", err)
		return ServerInterfaces{}, fmt.Errorf("invalid CIDR format: %v", err)
	}
	serverConfig := u.createServerCert(privateKey.String(), ip, port)
	data := &db.ServerCert{
		Private:  privateKey.String(),
		Public:   publicKey.String(),
		Endpoint: endpoint,
		Ip:       ip,
		Ifname:   ifname,
		Config:   serverConfig,
		Port:     port,
	}
	err = u.ServerRepo.CreateServerCert(data)
	if err != nil {
		log.Printf("NewInterface %v", err)
		return ServerInterfaces{}, err
	}

	err = u.startInterface(ifname)
	if err != nil {
		log.Printf("NewInterface %v", err)
		return ServerInterfaces{}, err
	}

	return ServerInterfaces{
		Private:  privateKey.String(),
		Public:   publicKey.String(),
		Endpoint: endpoint,
		Ip:       ip,
		Ifname:   ifname,
		Config:   serverConfig,
		Port:     port,
	}, nil
}

func (u *Usecases) startInterface(ifname string) error {
	server, err := u.ServerRepo.GetServerCertByIfname(ifname)
	if err != nil {
		log.Printf("startInterface %v", err)
		return err
	}
	err = exec.Command("ip", "link", "add", "dev", ifname, "type", "wireguard").Run()
	if err != nil {
		log.Printf("startInterface %v", err)
		return err
	}

	err = exec.Command("ip", "addr", "add", server.Ip, "dev", ifname).Run()
	if err != nil {
		log.Printf("startInterface %v", err)
		return err
	}

	err = exec.Command("ip", "link", "set", ifname, "up").Run()
	if err != nil {
		log.Printf("startInterface %v", err)
		return err
	}

	private, err := wgtypes.ParseKey(server.Private)
	if err != nil {
		log.Printf("startInterface %v", err)
		return err
	}
	serverIntereface, err := wgctrl.New()
	if err != nil {
		log.Printf("startInterface %v", err)
	}
	defer serverIntereface.Close()

	cfg := wgtypes.Config{
		PrivateKey:   &private,
		ListenPort:   &server.Port,
		ReplacePeers: true,
	}

	err = serverIntereface.ConfigureDevice(ifname, cfg)
	if err != nil {
		log.Printf("ConfigureDevice %v", err)
		return err
	}

	clients, err := u.ClientRepo.GetClientCertsByIfname(ifname)
	if err != nil {
		log.Printf("ConfigureDevice %v", err)
		return nil
	}
	for _, peer := range clients {
		err := u.setClient(ifname, peer.IP, peer.AllowedIPs, peer.Public)
		if err != nil {
			log.Printf("ConfigureDevice %v", err)
		}
	}
	return nil
}

func (u *Usecases) stopInterface(ifname string) error {
	err := exec.Command("ip", "link", "del", "dev", ifname, "type", "wireguard").Run()
	if err != nil {
		log.Printf("stopInterface %v", err)
		return err
	}

	return nil
}

func (u *Usecases) createServerCert(private, ip string, listenPort int) string {
	var builder strings.Builder
	builder.WriteString("[Interface]\n")
	builder.WriteString(fmt.Sprintf("PrivateKey = %s\n", private))
	builder.WriteString(fmt.Sprintf("Address = %s\n", ip))
	builder.WriteString(fmt.Sprintf("ListenPort = %d\n", listenPort))
	return builder.String()
}

func (u *Usecases) DeleteServer(private, ifname string) error {
	err := u.ServerRepo.DeleteServer(strings.TrimSpace(private), strings.TrimSpace(ifname))
	if err != nil {
		log.Printf("DeleteServer %v", err)
		return err
	}
	err = u.stopInterface(ifname)
	if err != nil {
		log.Printf("DeleteServer %v", err)
		return err
	}
	return nil

}

func (u *Usecases) StopInterface(ifname string) error {

	err := wg.CheckUpInterface(wg.NewWGFactory(), ifname)
	if err == nil {
		log.Printf("StopInterface %v", err)
		return nil
	}

	return u.stopInterface(ifname)
}

func (u *Usecases) StartInterface(ifname string) error {
	err := wg.CheckUpInterface(wg.NewWGFactory(), ifname)
	if err != nil {
		if err.Error() == fmt.Sprintf("exist up interface %s", ifname) {
			return nil
		}
		log.Printf("StartInterface %v", err)
		return err
	}
	return u.startInterface(ifname)
}

func (u *Usecases) CreateIptablesList(comment string, ips []string) error {

	stdout, err := exec.Command("ipset", "create", comment, "hash:ip").CombinedOutput()
	if err != nil && !strings.Contains(string(stdout), "already exists") {
		log.Printf("createIptablesList 1: %v", err)
		return err
	}
	stdout, err = exec.Command("ipset", "flush", comment).CombinedOutput()
	if err != nil {
		log.Printf("createIptablesList 1: err=%v out=%s", err, string(stdout))
	}

	for _, ip := range ips {
		out, err := exec.Command("ipset", "add", comment, strings.TrimSpace(ip)).CombinedOutput()
		if err != nil {
			log.Printf("createIptablesList 2: err=%v out=%s", err, string(out))
		}
	}
	return nil
}

func (u *Usecases) UpdateIpSetList(command, name string, ips []string, single bool) error {
	if single {
		for _, ip := range ips {
			out, err := exec.Command("ipset", command, name, strings.TrimSpace(ip)).CombinedOutput()
			if err != nil {
				log.Printf("UpdateIpSetList: err=%v out=%s", err, string(out))
				if strings.Contains(string(out), "name does not exist") {
					return fmt.Errorf("UpdateIpSetList: ipset %s does not exist check in iptables rules created ipset rules", name)
				}

				if strings.Contains(string(out), "already added") || strings.Contains(string(out), "not added") {
					return nil
				}
				return fmt.Errorf("UpdateIpSetList: %v", string(out))
			}

		}

	} else {
		err := u.CreateIptablesList(name, ips)
		if err != nil {
			log.Printf("UpdateIpSetList: CreateIptablesList failed: %v", err)
			return err
		}
	}

	return nil
}

func (u *Usecases) DeleteIptablesList(comment string) error {
	stdout, err := exec.Command("ipset", "destroy", comment).CombinedOutput()
	if err != nil {
		log.Printf("deleteIptablesList 1: err=%v out=%s", err, string(stdout))
		return errors.New(string(stdout))
	}
	return nil
}

func (u *Usecases) ipsStringToList(ips string) []string {
	iplist := strings.Split(strings.ReplaceAll(ips, " ", ""), ",")
	return iplist
}
func (u *Usecases) SetUsForward(position int, actionRaw, command, source, destination, protocol, port string, comment string, isList, except bool) error {
	var err error
	action := strings.ToUpper(actionRaw)
	switch command {
	case "write":
		if isList {
			err = u.CreateIptablesList(comment, u.ipsStringToList(destination))
			if err != nil {
				log.Printf("SetUsForward: createIptablesList failed: %v", err)
				return err
			}
			err = u.IpTables.SetForwardList(position, port, action, command, source, comment, protocol, comment, except)
			if err != nil {
				log.Printf("SetUsForward: SetForwardList failed: %v", err)
				return err
			}
		} else {
			err = u.IpTables.SetForward(position, port, action, command, source, destination, protocol, comment, except)
			if err != nil {
				log.Printf("SetUsForward: SetForward failed: %v", err)
				return err
			}
		}
		err := u.ServerRepo.CreateForward(position, port, action, source, destination, protocol, comment, isList, except)
		if err != nil {
			log.Printf("SetUsForward: CreateForward failed: %v", err)
			return err
		}
		return nil
	case "delete":
		if isList {
			err = u.IpTables.SetForwardList(position, port, action, command, source, comment, protocol, comment, except)
			if err != nil {
				log.Printf("SetUsForward: SetForwardList (delete) failed: %v", err)
				return err
			}
			err = u.DeleteIptablesList(comment)
			if err != nil {
				log.Printf("SetUsForward: deleteIptablesList failed: %v", err)
				return err
			}

		} else {
			err = u.IpTables.SetForward(position, port, action, command, source, destination, protocol, comment, except)
			if err != nil {
				log.Printf("SetUsForward: SetForward (delete) failed: %v", err)
				return err
			}
		}

		err = u.ServerRepo.DeleteForward(comment)
		if err != nil {
			log.Printf("SetUsForward: DeleteForward failed: %v", err)
			return err
		}
		return nil

	default:
		return fmt.Errorf("SetUsForward: unknown command: %s", command)
	}
}

func (u *Usecases) SetUsMasquerade(command, source, ifname, comment string) error {
	err := u.IpTables.SetMasquerade(command, source, ifname, comment)
	if err != nil {
		log.Printf("SetUsMasquerade %v", err)
		return err
	}
	err = u.ServerRepo.CreateMasquerade(source, ifname, comment)
	if err != nil {
		log.Printf("SetUsMasquerade %v", err)
		return err
	}

	switch command {
	case "write":
		err = u.ServerRepo.CreateMasquerade(source, ifname, comment)
		if err != nil {
			log.Printf("SetUsMasquerade %v", err)
			return err
		}
	case "delete":
		err = u.ServerRepo.DeleteMasquerade(source, ifname, comment)
		if err != nil {
			log.Printf("SetUsMasquerade %v", err)
			return err
		}
	default:
		return fmt.Errorf("command did not find %v", err)

	}
	return nil
}

func (u *Usecases) GetServerArchive() ([]ServerInterfaces, error) {
	data, err := u.ServerRepo.GetServerArchive()
	if err != nil {
		log.Printf("GetServerArchive %v", err)
		return []ServerInterfaces{}, err
	}
	if len(data) == 0 {
		return []ServerInterfaces{}, nil
	}
	var serIfname []ServerInterfaces
	for _, v := range data {
		serIfname = append(serIfname, ServerInterfaces{Ifname: v.Ifname, Ip: v.Ip, Port: v.Port, Private: v.Private, Public: v.Public, Endpoint: v.Endpoint})
	}
	return serIfname, err

}

func (u *Usecases) GetServerInterfaces() ([]ServerInterfaces, error) {
	data, err := u.ServerRepo.GetServerInterfaces()
	if err != nil || len(data) == 0 {
		return []ServerInterfaces{}, err
	}
	var serIfname []ServerInterfaces
	for _, v := range data {
		serIfname = append(serIfname, ServerInterfaces{Ifname: v.Ifname, Ip: v.Ip, Port: v.Port, Private: v.Private, Public: v.Public, Endpoint: v.Endpoint})
	}
	return serIfname, err

}

func (u *Usecases) StartInterfaces() {
	serverData, err := u.ServerRepo.GetServerCertificates()
	if err != nil {
		log.Printf("StartInterfaces %v", err)
		return
	}
	for _, v := range serverData {
		err := u.StartInterface(v.Ifname)
		if err != nil {
			log.Printf("StartInterfaces %v", err)
		}
	}
	clinetData, err := u.ClientRepo.GetAllClient()
	if err != nil {
		log.Printf("StartInterfaces %v", err)
		return
	}
	for _, v := range clinetData {
		err := u.setClient(v.Ifname, v.IP, v.AllowedIPs, v.Public)
		if err != nil {
			log.Printf("StartInterfaces %v", err)
		}
	}
}

func (u *Usecases) StopInterfaces() {
	serverData, err := u.ServerRepo.GetServerCertificates()
	if err != nil {
		log.Printf("StopInterfaces %v", err)
		return
	}
	for _, v := range serverData {
		err := u.stopInterface(v.Ifname)
		if err != nil {
			log.Printf("StopInterfaces %v", err)
		}
	}

}

func (u *Usecases) FirstStartIptables() {

	cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("FirstStartIptables %v", err)
	}
	err = u.IpTables.FlushForward()
	if err != nil {
		log.Printf("FirstStartIptables %v", err)
	}
	fwrd, err := u.ServerRepo.GetForward()
	if err != nil {
		log.Printf("FirstStartIptables %v", err)
	} else {
		for _, rule := range fwrd {
			if rule.IsList {
				dest := strings.Split(rule.Destination, ",")
				err := u.CreateIptablesList(rule.Comment, dest)
				if err != nil {
					log.Printf("FirstStartIptables %v", err)
				}
				err = u.IpTables.SetForwardList(rule.Position, rule.Port, rule.Action, "write", rule.Source, rule.Comment, rule.Protocol, rule.Comment, rule.Except)
				if err != nil {
					log.Printf("FirstStartIptables %v", err)
				}

			} else {
				err := u.IpTables.SetForward(rule.Position, rule.Port, rule.Action, "write", rule.Source, rule.Destination, rule.Protocol, rule.Comment, rule.Except)
				if err != nil {
					log.Printf("FirstStartIptables %v", err)
				}
			}
		}
	}
	masqr, err := u.ServerRepo.GetMasquerade()
	if err != nil {
		log.Printf("FirstStartIptables/GetMasquerade %v", err)
	} else {
		for _, v := range masqr {
			err := u.IpTables.SetMasquerade("write", v.Source, v.Ifname, v.Comment)
			if err != nil {
				log.Printf("FirstStartIptables/GetMasquerade %v", err)
			}
		}
	}

}

func (u *Usecases) backBytes(commnet string) string {
	cmd := exec.Command("iptables", "-t", "filter", "-L", "FORWARD", "-v", "-x", "-n", "--line-numbers")
	out, _ := cmd.Output()
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, commnet) {
			fields := strings.Fields(line)
			if len(fields) > 2 {
				bytes := fields[2]
				return bytes
			}
		}
	}
	return "none"
}

func (u *Usecases) GetIptablesRules() (IptablesRulesData, error) {
	masq := []UsMasquerade{}
	frwd := []UsForward{}

	forwardList, err := u.ServerRepo.GetForward()
	if err != nil {
		log.Printf("GetIptablesRules %v", err)
	}
	masqueradeList, err := u.ServerRepo.GetMasquerade()
	if err != nil {
		log.Printf("GetIptablesRules %v", err)
	}
	for _, v := range forwardList {
		frwd = append(frwd, UsForward{
			Bytes:       u.backBytes(v.Comment),
			Source:      v.Source,
			Destination: v.Destination,
			Position:    v.Position,
			Protocol:    v.Protocol,
			Port:        v.Port,
			Comment:     v.Comment,
			List:        v.IsList,
			Action:      v.Action,
			Except:      v.Except,
		})
	}

	for _, v := range masqueradeList {
		masq = append(masq, UsMasquerade{
			Ifname:  v.Ifname,
			Source:  v.Source,
			Comment: v.Comment,
		})
	}
	intList := u.getInterfaceList()
	return IptablesRulesData{Forward: frwd, Masquerade: masq, InterfaceList: intList}, nil
}

func (u *Usecases) getInterfaceList() []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("getInterfaceList %v", err)
	}

	var list = []string{}
	for _, iface := range interfaces {
		if iface.Name != "lo" {
			list = append(list, iface.Name)
		}
	}
	return list
}

func (u *Usecases) PingLoop(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			log.Println("PingLoop: context done, exiting ping loop")
			return
		default:
			data, err := u.ClientRepo.GetAllClient()
			if err != nil {
				log.Printf("PingLoop: %v", err)
				continue
			}
			var waitG sync.WaitGroup

			for _, client := range data {
				ip := strings.Split(client.IP, "/")
				if len(ip) > 0 {
					waitG.Add(1)
					u.PingStatus.Ping(ip[0], &waitG)
				}

			}
			waitG.Wait()
		}

		time.Sleep(5 * time.Second)
	}

}
