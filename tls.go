package promsentry

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// CreateTLSConfiguration process the paths to certificate and the client authentication type based on the
// values on the function parameter, then create a *tls.Config struct that can be passed to multiple stuff.
// Some of them being the HTTP server or gRPC server.
//
// Please do not insert empty string values for each function parameter. You will get an error instead.
func CreateTLSConfiguration(serverCertPath string, serverKeyPath string, clientCARootPath string, clientAuthenticationType string) (*tls.Config, error) {
	var clientAuthType tls.ClientAuthType
	switch clientAuthenticationType {
	case "NoClientCert":
		clientAuthType = tls.NoClientCert
	case "RequestClientCert":
		clientAuthType = tls.RequestClientCert
	case "RequireAnyClientCert":
		clientAuthType = tls.RequireAnyClientCert
	case "VerifyClientCertIfGiven":
		clientAuthType = tls.VerifyClientCertIfGiven
	case "RequireAndVerifyClientCert":
		clientAuthType = tls.RequireAndVerifyClientCert
	default:
		clientAuthType = tls.NoClientCert
	}

	serverCert, err := os.ReadFile(serverCertPath)
	if err != nil {
		return &tls.Config{}, fmt.Errorf("reading server certificate: %w", err)
	}

	serverKey, err := os.ReadFile(serverKeyPath)
	if err != nil {
		return &tls.Config{}, fmt.Errorf("reading server key: %w", err)
	}

	clientCARoot, err := os.ReadFile(clientCARootPath)
	if err != nil {
		return &tls.Config{}, fmt.Errorf("reading client CA root: %w", err)
	}

	serverCertificatePair, err := tls.X509KeyPair(serverCert, serverKey)
	if err != nil {
		return &tls.Config{}, fmt.Errorf("converting x509 key pair: %w", err)
	}

	caCertificatePool := x509.NewCertPool()

	if ok := caCertificatePool.AppendCertsFromPEM(clientCARoot); !ok {
		return &tls.Config{}, fmt.Errorf("invalid ca certificate")
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{serverCertificatePair},
		RootCAs:            caCertificatePool,
		ClientAuth:         clientAuthType,
		ClientCAs:          caCertificatePool,
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS11,
	}, nil
}
