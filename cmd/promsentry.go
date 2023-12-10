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
	"github.com/aldy505/promsentry"
	"github.com/aldy505/promsentry/sentry"
	"log"
	"os"
	"os/signal"
)

func main() {
	sentry.Init(sentry.ClientOptions{
		Dsn:           "",
		Debug:         true, // TODO: Configure this
		SampleRate:    1.0,
		DebugWriter:   nil,
		Transport:     nil,
		ServerName:    "",
		Release:       "",
		Dist:          "",
		Environment:   "",
		HTTPClient:    nil,
		HTTPTransport: nil,
		HTTPProxy:     "",
		HTTPSProxy:    "",
		CaCerts:       nil,
	})

	server, err := promsentry.NewServer()
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Println(err)
		}
	}()

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, os.Interrupt)

	<-exitSignal
	server.Shutdown(context.Background()) // TODO: handle error
}
