package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"wireguard_api/config"
	"wireguard_api/db"
	"wireguard_api/iptablerules"
	"wireguard_api/pingstatus"
	"wireguard_api/repository"
	"wireguard_api/usecases"
	"wireguard_api/webserver"
)

func main() {
	cfg, err := config.LoadConfig("/etc/wireguard_api.cfg")
	if err != nil {
		log.Fatalf("Loading config: %s", err)
	}
	fmt.Println("Version:", config.Version)
	fmt.Println("Server started:", cfg.IpPort)
	db := db.Init(cfg)
	ipt, err := iptablerules.CreateGoIptables()
	if err != nil {
		log.Fatalf("startup failed: %v", err)
		return
	}

	uc := &usecases.Usecases{
		ServerRepo: repository.NewServerCertRepository(db.DbInstance),
		ClientRepo: repository.NewClientCertRepository(db.DbInstance),
		IpTables:   iptablerules.Init(ipt),
		PingStatus: pingstatus.Init(pingstatus.NewICMPFactory()),
	}
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)
	go uc.PingLoop(ctx)
	uc.FirstStartIptables()
	uc.StartInterfaces()
	server := webserver.NewServer(uc)
	go server.StartWebServer(ctx, cfg)
	sig := <-sigs
	uc.StopInterfaces()
	log.Printf("Received signal: %s", sig.String())
	cancel()
	if err := db.Close(); err != nil {
		log.Printf("Closing database: %s", err)
	}

}
