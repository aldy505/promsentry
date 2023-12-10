package promsentry

type Configuration struct {
	ListenAddress string `json:"listen_address" yaml:"listen_address"`
	TLS           struct {
		CertificateAuthorityPath string `json:"certificate_authority_path" yaml:"certificate_authority_path"`
		ServerCertificatePath    string `json:"server_certificate_path" yaml:"server_certificate_path"`
		ServerKeyPath            string `json:"server_key_path" yaml:"server_key_path"`
		ClientAuthenticationType string `json:"client_authentication_type" yaml:"client_authentication_type"`
	} `json:"tls" yaml:"tls"`
	SentryDsn string `json:"sentry_dsn" yaml:"sentry_dsn"`
}
