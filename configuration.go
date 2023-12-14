package promsentry

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Configuration struct {
	ListenAddress string `json:"listen_address" yaml:"listen_address"`
	TLS           struct {
		CertificateAuthorityPath string `json:"certificate_authority_path" yaml:"certificate_authority_path"`
		ServerCertificatePath    string `json:"server_certificate_path" yaml:"server_certificate_path"`
		ServerKeyPath            string `json:"server_key_path" yaml:"server_key_path"`
		ClientAuthenticationType string `json:"client_authentication_type" yaml:"client_authentication_type"`
	} `json:"tls" yaml:"tls"`
	SentryDsn string `json:"sentry_dsn" yaml:"sentry_dsn"`
	Debug     bool   `json:"debug" yaml:"debug"`
}

func ParseConfiguration(filePath string) (*Configuration, error) {
	var configuration Configuration
	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("configuration file does not exists")
			}

			if errors.Is(err, os.ErrPermission) {
				return nil, fmt.Errorf("can't open the file as we're missing some permission to read it")
			}

			return nil, fmt.Errorf("unhandled file error: %w", err)
		}
		defer func() {
			_ = file.Close()
		}()

		switch path.Ext(filePath) {
		case "json":
			err := json.NewDecoder(file).Decode(&configuration)
			if err != nil {
				return nil, fmt.Errorf("failed parsing the config file for json format, make sure you got everything right")
			}
		case "yaml", "yml":
			err := yaml.NewDecoder(file).Decode(&configuration)
			if err != nil {
				return nil, fmt.Errorf("failed parsing the config file for yaml format, make sure you got everything right")
			}
		default:
			return nil, fmt.Errorf("configuration file format is not supported")
		}
	}

	// Read from environment variable (this takes priority)
	if v, ok := os.LookupEnv("LISTEN_ADDRESS"); ok {
		configuration.ListenAddress = v
	}

	if v, ok := os.LookupEnv("TLS_CERTIFICATE_AUTHORITY_PATH"); ok {
		configuration.TLS.CertificateAuthorityPath = v
	}

	if v, ok := os.LookupEnv("TLS_SERVER_CERTIFICATE_PATH"); ok {
		configuration.TLS.ServerCertificatePath = v
	}

	if v, ok := os.LookupEnv("TLS_SERVER_KEY_PATH"); ok {
		configuration.TLS.ServerKeyPath = v
	}

	if v, ok := os.LookupEnv("TLS_CLIENT_AUTHENTICATION_TYPE"); ok {
		configuration.TLS.ClientAuthenticationType = v
	}

	if v, ok := os.LookupEnv("SENTRY_DSN"); ok {
		configuration.SentryDsn = v
	}

	if v, ok := os.LookupEnv("DEBUG"); ok {
		b, err := strconv.ParseBool(v)
		if err == nil {
			configuration.Debug = b
		}
	}

	return &configuration, nil
}
