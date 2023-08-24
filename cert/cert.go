package cert

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"log"
	"os"
	"time"

	certTLS "github.com/sergey-shpilevskiy/go-cert/tls"
	certX509 "github.com/sergey-shpilevskiy/go-cert/x509"
	"google.golang.org/grpc/credentials"

	"github.com/intellisoftalpin/cardano-wallet-backend/config"
	"github.com/intellisoftalpin/cardano-wallet-backend/helpers"
)

// TODO: Check if server's certificate and private key are expired

var (
	serverCertFile = "server-cert.pem"
	serverKeyFile  = "server-key.pem"
	caCertFile     = "ca-cert.pem"
	// caKeyFile      = "ca-key.pem"
	// clientCertFile = "client-cert.pem"
	// clientKeyFile  = "client-key.pem"
)

type CertInfo struct {
	certGenerator *certX509.CertGenerator
	tlsConfig     config.TLSConfig
}

// SetupTLS sets up TLS for the server
func SetupTLS(tlsConfig config.TLSConfig) (tlsCredentials credentials.TransportCredentials, err error) {
	// Load server's certificate and private key
	certificate, err := tls.LoadX509KeyPair(tlsConfig.CertPath+"/"+serverCertFile, tlsConfig.CertPath+"/"+serverKeyFile)
	if err != nil {
		log.Println("cannot load server's certificate and private key: ", err)
		log.Println("generating new server's certificate and private key")
		certInfo := CertInfo{}
		return certInfo.GenerateCerts(tlsConfig)
	}

	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(tlsConfig.CertPath + "/" + caCertFile)
	if err != nil {
		log.Println("cannot load CA certificate: ", err)
		log.Println("generating new CA certificate")
		certInfo := CertInfo{}
		return certInfo.GenerateCerts(tlsConfig)
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatal("cannot append CA certificate to the pool")
	}

	// Create the credentials and return it
	config := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	}

	tlsCredentials = credentials.NewTLS(config)
	return tlsCredentials, err
}

func (c *CertInfo) GenerateCerts(tlsConf config.TLSConfig) (tlsCredentials credentials.TransportCredentials, err error) {
	c.certGenerator = certX509.NewCertGenerator(tlsConf.CertPath, time.Now().Add(time.Hour*24*365*5))
	c.tlsConfig = tlsConf

	if err = c.certGenerator.GenerateCA(serverCAName); err != nil {
		log.Println("cannot generate CA: ", err)
		return tlsCredentials, err
	}

	if tlsCredentials, err = c.GenerateServerCert(); err != nil {
		log.Println(err)
		return tlsCredentials, err
	}

	if err = c.GenerateClientCert(); err != nil {
		log.Println(err)
		return tlsCredentials, err
	}

	return tlsCredentials, err
}

func (c *CertInfo) GenerateServerCert() (tlsCredentials credentials.TransportCredentials, err error) {
	if err = c.certGenerator.GenerateServerCert(serverCertName, helpers.StringToIP(c.tlsConfig.IPs)); err != nil {
		log.Println(err)
		return tlsCredentials, err
	}

	tlsCredentials, err = certTLS.LoadTLSCredentials(c.tlsConfig.CertPath+"/"+serverCertFile, c.tlsConfig.CertPath+"/"+serverKeyFile)
	if err != nil {
		log.Println("cannot load TLS credentials: ", err)
		return tlsCredentials, err
	}

	return tlsCredentials, err
}

func (c *CertInfo) GenerateClientCert() (err error) {
	if err = c.certGenerator.GenerateClientCert(clientCertName, helpers.StringToIP(c.tlsConfig.IPs)); err != nil {
		log.Println(err)
		return err
	}

	return err
}

// ------------------------------------------------------------------------------------------

var serverCAName pkix.Name = pkix.Name{
	CommonName: "Token-lib-cnode CA",
	// Organization: []string{"Cardano Node Operator"},
	// Country:       []string{"US"},
	// Province:      []string{""},
	// Locality:      []string{"San Francisco"},
	// StreetAddress: []string{"Golden Gate Bridge"},
	// PostalCode:    []string{"94016"},
}

var serverCertName pkix.Name = pkix.Name{
	CommonName: "Token-lib-cnode Cert",
	// Organization: []string{"Cardano Node Operator"},
	// Country:       []string{"US"},
	// Province:      []string{""},
	// Locality:      []string{"San Francisco"},
	// StreetAddress: []string{"Golden Gate Bridge"},
	// PostalCode:    []string{"94016"},
}

var clientCertName pkix.Name = pkix.Name{
	CommonName: "Token-lib-proxy Cert",
	// Organization: []string{"Cardano Node Operator"},
	// Country:       []string{"US"},
	// Province:      []string{""},
	// Locality:      []string{"San Francisco"},
	// StreetAddress: []string{"Golden Gate Bridge"},
	// PostalCode:    []string{"94016"},
}

// []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
