package promsentry

import (
	"bytes"
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/aldy505/promsentry/sentry"
	"github.com/aldy505/promsentry/statsd"
	"github.com/prometheus/prometheus/storage/remote"
)

func NewServer(listenAddress string, tlsConfig *tls.Config) (*http.Server, error) {
	if listenAddress == "" {
		listenAddress = "127.0.0.1:3000"
	}

	router := http.NewServeMux()
	router.HandleFunc("/api/v1/write", func(w http.ResponseWriter, r *http.Request) {
		req, err := remote.DecodeWriteRequest(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b := &bytes.Buffer{}
		client := statsd.NewClient(b)
		hub := sentry.CurrentHub()

		for _, timeseries := range req.GetTimeseries() {
			var name string
			var tags = make(map[string]string)
			for _, l := range timeseries.GetLabels() {
				if l.GetName() == "__name__" {
					name = l.GetValue()
					continue
				}

				tags[l.GetName()] = l.GetValue()
			}

			for _, s := range timeseries.GetSamples() {
				err := client.Gauge(name, int64(s.GetValue()), tags)
				if err != nil {
					log.Println(err)
				}
			}

			for _, e := range timeseries.GetExemplars() {
				err := client.Duration(name, time.Duration(e.GetValue()), tags)
				if err != nil {
					log.Println(err)
				}
				for _, l := range e.GetLabels() {
					tags[l.GetName()] = l.GetValue()
				}
			}

			for _, hp := range timeseries.GetHistograms() {
				h := remote.HistogramProtoToHistogram(hp)
				err := client.Histogram(name, h.Count, tags)
				if err != nil {
					log.Println(err)
				}
			}
		}

		if err := client.Flush(); err != nil {
			log.Println(err)
		}

		metric := b.Bytes()
		hub.CaptureMetric(metric)

		w.WriteHeader(200)
	})

	server := &http.Server{
		Addr:              listenAddress,
		Handler:           router,
		TLSConfig:         tlsConfig,
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      time.Minute,
		IdleTimeout:       time.Minute,
	}

	return server, nil
}
