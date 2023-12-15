//	Copyright 2023 Reinaldy Rafli <aldy505@proton.me>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/aldy505/promsentry"
	"github.com/aldy505/promsentry/sentry"
)

func main() {
	var configurationFilePath string
	flag.StringVar(&configurationFilePath, "config-file", "", "Path to configuration file (JSON, or YAML)")
	flag.Parse()
	if v, ok := os.LookupEnv("CONFIG_FILE_PATH"); ok {
		configurationFilePath = v
	}

	configuration, err := promsentry.ParseConfiguration(configurationFilePath)
	if err != nil {
		log.Fatalln(err)
		return
	}

	err = sentry.Init(sentry.ClientOptions{
		Dsn:        configuration.SentryDsn,
		Debug:      configuration.Debug,
		SampleRate: 1.0,
		ServerName: "promsentry",
	})
	if err != nil {
		log.Fatalln(err)
		return
	}

	var tlsConfig *tls.Config = nil
	if configuration.TLS.ServerCertificatePath != "" {
		tlsConfig, err = promsentry.CreateTLSConfiguration(
			configuration.TLS.ServerCertificatePath,
			configuration.TLS.ServerKeyPath,
			configuration.TLS.CertificateAuthorityPath,
			configuration.TLS.ClientAuthenticationType)
		if err != nil {
			log.Fatalln(err)
			return
		}
	}

	server, err := promsentry.NewServer(configuration.ListenAddress, tlsConfig)
	if err != nil {
		log.Fatalln(err)
		return
	}

	go func() {
		if tlsConfig == nil {
			log.Printf("Server starting on http://%s\n", server.Addr)
			err := server.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Println(err)
			}
		} else {
			log.Printf("Server starting on https://%s\n", server.Addr)
			err := server.ListenAndServeTLS("", "")
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Println(err)
			}
		}
	}()

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, os.Interrupt)

	<-exitSignal
	err = server.Shutdown(context.Background())
	if err != nil {
		log.Println(err)
	}
}
