package webserver

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"
	"wireguard_api/config"
	"wireguard_api/controllers"
	"wireguard_api/usecases"

	"github.com/gin-gonic/gin"
)

func NewServer(usecaseServices *usecases.Usecases) *Server {
	return &Server{Service: usecaseServices}
}

func (c *Server) StartWebServer(ctx context.Context, cfg *config.ServerConfig) {

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(checkIpAccess(cfg))
	r.Use(checkToken(cfg))

	ctrl := controllers.NewController(c.Service, cfg)
	r.GET("/version", ctrl.GetVersion)

	//server certs
	r.POST("/interface/new", ctrl.AddInterface) // create new interface and server certificate
	r.DELETE("/interface", ctrl.CtrlDeleteServer)
	r.POST("/interface/stop", ctrl.CtrlStopServer)
	r.POST("/interface/start", ctrl.CtrlStartServer)
	r.GET("/interface/all", ctrl.CtrlGetInterfaces)
	r.GET("/interface/archive", ctrl.CtrlGetServerArchive)
	// iptables
	r.POST("/server/forward", ctrl.SetForward)
	r.POST("/server/forward/updateList", ctrl.SetForwardUpdateList)
	r.POST("/server/masquerade", ctrl.SetMasquerade)
	r.GET("/server/rules", ctrl.CtrlGetIptables)

	//clients certs
	r.POST("/clients/new", ctrl.AddClient)
	r.DELETE("/clients", ctrl.DeleteClient)
	r.GET("/clients/getall", ctrl.GetAllClients)
	r.GET("/clients/status", ctrl.GetStatus)
	r.GET("/clients/archive", ctrl.GetClientArchive)

	server := &http.Server{
		Addr:    cfg.IpPort,
		Handler: r,
		TLSConfig: &tls.Config{
			Certificates: getCert(cfg),
		},
	}
	go func() {
		if err := server.ListenAndServeTLS("", ""); err != nil {
			log.Print(err)
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func checkToken(cfg *config.ServerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		if tokenString != cfg.Token {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func checkIpAccess(cfg *config.ServerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		for _, ip := range cfg.WhiteListIpAccess {
			if clientIP == ip {
				c.Next()
				return
			}
		}
		c.JSON(http.StatusForbidden, gin.H{"result": "Access denied your ip is not in whitelist"})
		c.Abort()
	}
}

func getCert(cfg *config.ServerConfig) []tls.Certificate {
	var cert []tls.Certificate
	loadCert, err := tls.LoadX509KeyPair(cfg.TlsPublic, cfg.TlsPrivate)
	if err != nil {
		cert = append(cert, generateSelfSigned())
		log.Printf("Cannot load TLS certificates %v. Created and using self-signed", err)
	} else {
		cert = append(cert, loadCert)
	}

	return cert
}

func generateSelfSigned() tls.Certificate {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Printf(err.Error())
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		log.Printf(err.Error())
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{"wireguard_api"},
			CommonName:         "wireguard_api",
			OrganizationalUnit: []string{"wireguard_api"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Println(err.Error())
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		log.Println(err.Error())
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
	cert, err := tls.X509KeyPair(certPEM, privPEM)
	if err != nil {
		log.Println(err.Error())
	}
	return cert
}
