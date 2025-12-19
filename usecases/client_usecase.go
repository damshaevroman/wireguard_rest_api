package usecases

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
	"wireguard_api/db"

	"inet.af/netaddr"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func (u *Usecases) NewClient(ifname, ip, allowedIp string) (ClientResponse, error) {

	ifname = strings.TrimSpace(ifname)
	ip = strings.TrimSpace(ip)
	allowedIp = strings.TrimSpace(allowedIp)

	re := regexp.MustCompile(`[ ,]+`)
	normalAlloweIp := re.ReplaceAllString(allowedIp, ",")

	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		log.Printf("NewClient %v", err)
		return ClientResponse{}, err
	}
	publicKey := privateKey.PublicKey()

	err = u.checkIpMask(ifname, ip)
	if err != nil {
		log.Printf("NewClient %v", err)
		return ClientResponse{}, err
	}

	servData, err := u.ClientRepo.GetPublicEnpointPort(ifname)
	if err != nil {
		log.Printf("NewClient %v", err)
		return ClientResponse{}, err
	}

	if ip == "" {
		data, err := u.ClientRepo.GetPublicEnpointPort(ifname)
		if err != nil {
			log.Printf("NewClient %v", err)
			return ClientResponse{}, err
		}
		ip, err = u.generateIPs(ifname, data.Ip, servData.Ip)
		if err != nil {
			log.Printf("NewClient %v", err)
			return ClientResponse{}, err
		}

	}

	_, ipList, err := u.containsIp(normalAlloweIp, ifname)
	if err != nil {
		log.Printf("NewClient %v", err)
		return ClientResponse{}, err
	}

	config := u.createConfig(privateKey.String(), ip, servData.Public, ipList, servData.Endpoint, servData.Port)
	err = u.ClientRepo.CreateClientCert(&db.ClientCert{
		Ifname:     ifname,
		Private:    privateKey.String(),
		Public:     publicKey.String(),
		IP:         ip,
		AllowedIPs: normalAlloweIp,
		Config:     config,
	})
	if err != nil {
		log.Printf("NewClient %v", err)
		return ClientResponse{}, err
	}

	err = u.setClient(ifname, ip, normalAlloweIp, publicKey.String())
	if err != nil {
		log.Printf("NewClient %v", err)
		return ClientResponse{}, err
	}

	return ClientResponse{Ifname: ifname, Private: privateKey.String(), Public: publicKey.String(), Config: config, Ip: ip, AllowedIPs: normalAlloweIp}, nil

}

func (u *Usecases) getInterfaceSubnet(interfaceName string) (*net.IPNet, error) {
	iface, err := net.InterfaceByName(interfaceName)

	if err != nil {
		log.Printf("getInterfaceSubnet %v", err)
		return nil, fmt.Errorf("did not find interface %s: %w", interfaceName, err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		log.Printf("getInterfaceSubnet %v", err)
		return nil, fmt.Errorf("cannnot get address %s: %w", interfaceName, err)
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok {
			if ipNet.IP.To4() != nil {
				return ipNet, nil
			}
		}
	}
	return nil, fmt.Errorf("subnet %s not found", interfaceName)
}

func (u *Usecases) generateIPs(ifname, ipmask, serverIp string) (string, error) {
	prefix, err := netaddr.ParseIPPrefix(ipmask)
	if err != nil {
		log.Printf("generateIPs %v", err)
		return "", fmt.Errorf("invalid CIDR format: %v", err)
	}
	_, networkIp, err := net.ParseCIDR(ipmask)
	if err != nil {
		log.Printf("generateIPs %v", err)
		return "", fmt.Errorf("invalid CIDR format: %v", err)
	}

	listIp, err := u.ClientRepo.GetListIp(ifname)
	if err != nil {
		return "", err
	}
	ipSet := make(map[string]struct{})
	for _, ip := range listIp {
		ipSet[ip] = struct{}{}
	}

	ipSet[networkIp.String()] = struct{}{}
	ipSet[serverIp] = struct{}{}
	ipRange := prefix.Range()
	for ip := ipRange.From(); prefix.Contains(ip); ip = ip.Next() {
		newIp := fmt.Sprintf("%s/%d", ip.String(), prefix.Bits())
		if _, exists := ipSet[newIp]; !exists {
			return newIp, nil
		}
	}

	return "", fmt.Errorf("cannot find free ip for interface %s subnet %s", ifname, networkIp)
}

func (u *Usecases) createConfig(private, ip, public, allowedIp, endpoint string, port int) string {
	endpoint = fmt.Sprintf("%s:%d", endpoint, port)
	var builder strings.Builder
	builder.WriteString("[Interface]\n")
	builder.WriteString(fmt.Sprintf("PrivateKey = %s\n", private))
	builder.WriteString(fmt.Sprintf("Address = %s\n", ip))
	builder.WriteString("[Peer]\n")
	builder.WriteString(fmt.Sprintf("PublicKey = %s\n", public))
	builder.WriteString(fmt.Sprintf("AllowedIPs = %s\n", allowedIp))
	builder.WriteString(fmt.Sprintf("Endpoint = %s\n", endpoint))
	builder.WriteString("PersistentKeepalive = 20\n")
	return builder.String()
}

func (u *Usecases) containsIp(allowedIp, ifname string) ([]net.IPNet, string, error) {
	interfaceSubnet, err := u.getInterfaceSubnet(ifname)
	if err != nil {
		log.Printf("createConfig %v", err)
		return nil, "", err
	}

	_, subnetInt, err := net.ParseCIDR(interfaceSubnet.String())
	if err != nil {
		return nil, "", err
	}

	mapIP := make(map[string]net.IPNet)
	mapIP[subnetInt.String()] = *subnetInt

	allowedIpRaw := strings.Split(allowedIp, ",")
	for i := range allowedIpRaw {
		allowedIpRaw[i] = strings.TrimSpace(allowedIpRaw[i])
	}

	for _, v := range allowedIpRaw {
		if v == "" {
			continue
		}

		_, subnet, err := net.ParseCIDR(v)
		if err != nil {
			continue
		}
		mapIP[subnet.String()] = *subnet
	}

	var arrayModify []net.IPNet
	var stringModify []string
	for sub, valueNet := range mapIP {
		arrayModify = append(arrayModify, valueNet)
		stringModify = append(stringModify, sub)
	}

	return arrayModify, strings.Join(stringModify, ","), nil
}

func (u *Usecases) setClient(ifname string, ipClient, allowedIp, publicKey string) error {
	client, err := wgctrl.New()
	if err != nil {
		return err
	}
	defer client.Close()

	decodedKey, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		log.Printf("setClient %v", err)
		return err
	}

	ip, clientSubnet, err := net.ParseCIDR(ipClient)
	if err != nil {
		log.Printf("setClient %v", err)
		return err
	}
	clientSubnet.IP = ip
	clientSubnet.Mask = net.CIDRMask(32, 32)

	var arraNetIpNet []net.IPNet
	arraNetIpNet = append(arraNetIpNet, *clientSubnet)

	allowedIp = strings.TrimSpace(allowedIp)
	allowedIps := strings.Split(allowedIp, ",")
	if len(allowedIp) > 0 {
		for _, v := range allowedIps {
			if v != "" {
				_, sub, err := net.ParseCIDR(v)
				if err != nil {
					log.Printf("setClient %v", err)
					continue
				}
				if !sub.Contains(ip.To4()) {
					o, b := sub.Mask.Size()
					sub.Mask = net.CIDRMask(o, b)
					arraNetIpNet = append(arraNetIpNet, *sub)
				}
			}
		}
	}

	key := wgtypes.Key{}
	copy(key[:], decodedKey)

	peer := wgtypes.PeerConfig{
		PublicKey:         key,
		AllowedIPs:        arraNetIpNet,
		ReplaceAllowedIPs: true,
	}

	err = client.ConfigureDevice(ifname, wgtypes.Config{
		Peers: []wgtypes.PeerConfig{peer},
	})
	if err != nil {
		return err
	}

	return nil
}

func (u *Usecases) checkIpMask(ifname, ip string) error {
	var clientIp net.IP
	var networkIp *net.IPNet
	var err error

	if ip != "" {
		clientIp, networkIp, err = net.ParseCIDR(ip)
		if err != nil {
			log.Printf("checkIpMask %v", err)
			return fmt.Errorf("invalid CIDR format: %v", err)
		}
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("checkIpMask %v", err)
		return fmt.Errorf("error getting network interfaces: %v", err)
	}

	for _, iface := range interfaces {
		if iface.Name == ifname {
			if ip != "" {
				addrs, err := iface.Addrs()
				if err != nil {
					log.Printf("checkIpMask %v", err)
					return fmt.Errorf("error getting addresses for interface %s: %v", ifname, err)
				}
				for _, addr := range addrs {
					ipAddr, _, err := net.ParseCIDR(addr.String())
					if err != nil {
						log.Printf("checkIpMask %v", err)
						return fmt.Errorf("error parsing address %s: %v", addr.String(), err)
					}
					if ipAddr.To4() != nil {
						addr, ipNet, _ := net.ParseCIDR(addr.String())

						if networkIp.String() != ipNet.String() {
							return fmt.Errorf("incorrect subnet your ip %s and interface %s", networkIp.String(), ipNet.String())
						}
						if clientIp.String() == addr.String() {

							return fmt.Errorf("ip %s cannot be same as interface %s", clientIp.String(), ipNet.IP.String())
						}
					}

				}
			}
			return nil
		}
	}
	return fmt.Errorf("interface %s not found", ifname)

}

func (u *Usecases) GetStatus() ([]InterfaceListStatus, error) {
	client, err := wgctrl.New()
	if err != nil {
		log.Printf("GetStatus %v", err)
		return []InterfaceListStatus{}, err
	}

	var ifnameStatus []InterfaceListStatus
	devices, err := client.Devices()
	if err != nil {
		log.Printf("GetStatus %v", err)
		return []InterfaceListStatus{}, err
	}
	if len(devices) == 0 {
		return []InterfaceListStatus{}, nil
	}
	for _, device := range devices {
		var status []Status
		if len(device.Peers) > 0 {
			for _, peer := range device.Peers {
				var ipList []AllowerIP
				for _, v := range peer.AllowedIPs {
					o, _ := v.Mask.Size()
					ipList = append(ipList, AllowerIP{Ip: v.IP.String(), Mask: o})
				}
				status = append(status, Status{
					Public:            peer.PublicKey.String(),
					LastHandshakeTime: peer.LastHandshakeTime,
					Recieved:          peer.ReceiveBytes,
					Transmit:          peer.TransmitBytes,
					AllowedIp:         ipList,
					Endpoint:          peer.Endpoint.String(),
				})
			}
		}
		if len(status) > 0 {
			ifnameStatus = append(ifnameStatus, InterfaceListStatus{
				Ifname: device.Name,
				Status: status,
			})
		} else {
			ifnameStatus = append(ifnameStatus, InterfaceListStatus{
				Ifname: device.Name,
				Status: []Status{},
			})
		}
	}
	return ifnameStatus, nil
}

func (u *Usecases) GetAllClients() ([]ClientResponse, error) {
	allClients, err := u.ClientRepo.GetAllClient()
	if err != nil {
		return nil, err
	}
	if len(allClients) == 0 {
		return []ClientResponse{}, nil
	}
	var clientList []ClientResponse
	for _, v := range allClients {
		var tStatus bool
		var pTime int64
		ipData := strings.Split(v.IP, "/")
		if len(ipData) > 1 {
			status, pingTime := u.PingStatus.Read(ipData[0])
			pTime = pingTime.Microseconds()
			tStatus = status
		} else {
			tStatus = false
			pTime = 0
		}

		clientList = append(clientList, ClientResponse{
			Ifname:     v.Ifname,
			Private:    v.Private,
			Public:     v.Public,
			Ip:         v.IP,
			AllowedIPs: v.AllowedIPs,
			Config:     v.Config,
			PingStatus: ClientResponsePing{
				Status:   tStatus,
				PintTime: pTime,
			},
		})
	}
	return clientList, nil
}

func (u *Usecases) DeleteClient(public string) error {
	public = strings.TrimSpace(public)
	cert, err := u.ClientRepo.DeleteClientCert(public)
	if err != nil {
		log.Printf("DeleteClient %v", err)
		return err
	}
	client, err := wgctrl.New()
	if err != nil {
		log.Printf("DeleteClient %v", err)
	}
	defer client.Close()
	peerPubKey, err := wgtypes.ParseKey(cert.Public)
	if err != nil {
		log.Printf("DeleteClient %v", err)
	}
	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey: peerPubKey,
				Remove:    true, // <- вот тут указано удаление
			},
		},
	}
	err = client.ConfigureDevice(cert.Ifname, cfg)
	if err != nil {
		log.Printf("DeleteClient %v", err)
		return err
	}
	ip := strings.Split(cert.IP, "/")
	if len(ip) > 0 {
		u.PingStatus.Delete(ip[0])
	}

	return nil

}

func (u *Usecases) GetClientArchive() ([]ClientResponse, error) {
	data, err := u.ClientRepo.GetClientArchive()
	if err != nil {
		log.Printf("GetClientArchive %v", err)
		return []ClientResponse{}, err
	}
	if len(data) == 0 {
		return []ClientResponse{}, nil
	}
	var clientArchive []ClientResponse
	for _, v := range data {
		clientArchive = append(clientArchive, ClientResponse{
			Ifname:     v.Ifname,
			Private:    v.Private,
			Public:     v.Public,
			Ip:         v.IP,
			AllowedIPs: v.AllowedIPs,
			Config:     v.Config,
		})
	}
	return clientArchive, err

}
