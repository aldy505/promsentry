package promsentry

import (
	"bytes"
	"github.com/aldy505/promsentry/sentry"
	"github.com/aldy505/promsentry/statsd"
	"github.com/prometheus/prometheus/storage/remote"
	"net/http"
	"time"
)

func NewServer() (*http.Server, error) {
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
				client.Gauge(name, int64(s.GetValue()), tags) // TODO: Properly handle error
			}

			for _, e := range timeseries.GetExemplars() {
				client.Duration(name, time.Duration(e.GetValue()), tags) // TODO: Properly handle error
				for _, l := range e.GetLabels() {
					tags[l.GetName()] = l.GetValue()
				}
			}

			for _, hp := range timeseries.GetHistograms() {
				h := remote.HistogramProtoToHistogram(hp)
				client.Histogram(name, h.Count, tags) // TODO: Properly handle error
			}
		}

		client.Flush() // TODO: Properly handle error

		metric := b.Bytes()
		hub.CaptureMetric(metric)

		w.WriteHeader(200)
	})

	server := &http.Server{
		Addr:              "127.0.0.1:3000", // TODO: Change me!
		Handler:           router,
		TLSConfig:         nil,
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		IdleTimeout:       0,
	}

	return server, nil
}
