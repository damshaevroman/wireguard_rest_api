package pingstatus

import (
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"log"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type ICMPConn interface {
	WriteTo([]byte, net.Addr) (int, error)
	ReadFrom([]byte) (int, net.Addr, error)
	SetReadDeadline(time.Time) error
	Close() error
}

type ICMPFactory interface {
	Listen() (ICMPConn, error)
}

func NewICMPFactory() ICMPFactory {
	return &RealICMPFactory{}
}

type RealICMPFactory struct{}

func (f *RealICMPFactory) Listen() (ICMPConn, error) {
	return icmp.ListenPacket("ip4:icmp", "0.0.0.0")
}

type PingStatus struct {
	Mu         sync.Mutex
	PingStatus map[string]pingStatus
	Factory    ICMPFactory
}

func Init(factory ICMPFactory) *PingStatus {
	return &PingStatus{
		PingStatus: make(map[string]pingStatus),
		Factory:    factory,
	}
}

type pingStatus struct {
	Status   bool
	PintTime time.Duration
}

func (u *PingStatus) Ping(target string, wg *sync.WaitGroup) {
	defer wg.Done()
	icmpMessage := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("HELLO-R-U-THERE"),
		},
	}

	msgData, err := icmpMessage.Marshal(nil)
	if err != nil {
		log.Printf("ping 1: %s", err.Error())
		return
	}

	conn, err := u.Factory.Listen()
	if err != nil {
		log.Printf("ping 2: %s", err.Error())
		return
	}
	defer conn.Close()

	start := time.Now()
	if _, err := conn.WriteTo(msgData, &net.IPAddr{IP: net.ParseIP(target)}); err != nil {
		if !strings.Contains(err.Error(), "destination address required") {
			log.Printf("ping 3: %s", err.Error())
		}

		return
	}

	reply := make([]byte, 1500)
	err = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		log.Printf("ping 4: %s", err.Error())
		return
	}

	n, _, err := conn.ReadFrom(reply)
	if err != nil {
		if strings.Contains(err.Error(), "i/o timeout") {
			u.write(false, target, 0)
		}
		return
	}

	duration := time.Since(start)
	receivedMsg, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), reply[:n])
	if err != nil {
		log.Printf("ping 5: %s", err.Error())
		return
	}

	switch receivedMsg.Type {
	case ipv4.ICMPTypeEchoReply:
		u.write(true, target, duration)
	default:
		u.write(false, target, 0)
	}
}

func (u *PingStatus) write(status bool, ip string, duration time.Duration) {
	u.Mu.Lock()
	defer u.Mu.Unlock()

	u.PingStatus[ip] = pingStatus{
		Status:   status,
		PintTime: duration,
	}
}

func (u *PingStatus) Read(ip string) (bool, time.Duration) {
	u.Mu.Lock()
	defer u.Mu.Unlock()
	if status, exists := u.PingStatus[ip]; exists {
		return status.Status, status.PintTime
	} else {
		return false, 0
	}
}

func (u *PingStatus) Delete(ip string) {
	u.Mu.Lock()
	defer u.Mu.Unlock()
	delete(u.PingStatus, ip)
}
