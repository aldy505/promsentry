package sentry

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
)

type customComplexError struct {
	Message string
}

func (e customComplexError) Error() string {
	return "customComplexError: " + e.Message
}

func (e customComplexError) AnswerToLife() string {
	return "42"
}

func setupClientTest() (*Client, *ScopeMock, *TransportMock) {
	scope := &ScopeMock{}
	transport := &TransportMock{}
	client, _ := NewClient(ClientOptions{
		Dsn:       "http://whatever@example.com/1337",
		Transport: transport,
	})

	return client, scope, transport
}

func TestCaptureEvent(t *testing.T) {
	client, _, transport := setupClientTest()

	eventID := EventID("0123456789abcdef")
	timestamp := time.Now().UTC()
	serverName := "testServer"

	client.CaptureEvent(&Event{
		EventID:    eventID,
		Timestamp:  timestamp,
		ServerName: serverName,
	}, nil, nil)

	if transport.lastEvent == nil {
		t.Fatal("missing event")
	}
	want := &Event{
		EventID:    eventID,
		Timestamp:  timestamp,
		ServerName: serverName,
		Level:      LevelInfo,
		Platform:   "go",
		Sdk: SdkInfo{
			Name:         "sentry.go",
			Version:      SDKVersion,
			Integrations: []string{},
			Packages: []SdkPackage{
				{
					// FIXME: name format doesn't follow spec in
					// https://docs.sentry.io/development/sdk-dev/event-payloads/sdk/
					Name:    "sentry-go",
					Version: SDKVersion,
				},
				// TODO: perhaps the list of packages is incomplete or there
				// should not be any package at all. We may include references
				// to used integrations like http, echo, gin, etc.
			},
		},
	}
	got := transport.lastEvent
	opts := cmp.Options{cmpopts.IgnoreFields(Event{}, "Release", "sdkMetaData", "attachments")}
	if diff := cmp.Diff(want, got, opts); diff != "" {
		t.Errorf("Event mismatch (-want +got):\n%s", diff)
	}
}

func TestCaptureEventShouldSendEventWithMessage(t *testing.T) {
	client, scope, transport := setupClientTest()
	event := NewEvent()
	event.Message = "event message"
	client.CaptureEvent(event, nil, scope)
	assertEqual(t, transport.lastEvent.Message, "event message")
}

func TestSampleRate(t *testing.T) {
	tests := []struct {
		SampleRate float64
		// tolerated range is [SampleRate-MaxDelta, SampleRate+MaxDelta]
		MaxDelta float64
	}{
		{0.00, 0.0},
		{0.25, 0.2},
		{0.50, 0.2},
		{0.75, 0.2},
		{1.00, 0.0},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprint(tt.SampleRate), func(t *testing.T) {
			var (
				total   uint64
				sampled uint64
			)
			// Call sample from multiple goroutines just like multiple hubs
			// sharing a client would. This should help uncover data races.
			var wg sync.WaitGroup
			for i := 0; i < 4; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < 10000; j++ {
						atomic.AddUint64(&total, 1)
						s := sample(tt.SampleRate)
						switch tt.SampleRate {
						case 0:
							if s {
								panic("sampled true when rate is 0")
							}
						case 1:
							if !s {
								panic("sampled false when rate is 1")
							}
						}
						if s {
							atomic.AddUint64(&sampled, 1)
						}
					}
				}()
			}
			wg.Wait()

			rate := float64(sampled) / float64(total)
			if !(tt.SampleRate-tt.MaxDelta <= rate && rate <= tt.SampleRate+tt.MaxDelta) {
				t.Errorf("effective sample rate was %f, want %f±%f", rate, tt.SampleRate, tt.MaxDelta)
			}
		})
	}
}

func BenchmarkProcessEvent(b *testing.B) {
	c, err := NewClient(ClientOptions{
		SampleRate: 0.25,
		Transport:  &TransportMock{},
	})
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		c.processEvent(&Event{}, nil, nil)
	}
}

func TestSDKIdentifier(t *testing.T) {
	client, _, _ := setupClientTest()
	assertEqual(t, client.GetSDKIdentifier(), "sentry.go")

	client.SetSDKIdentifier("sentry.go.test")
	assertEqual(t, client.GetSDKIdentifier(), "sentry.go.test")
}

func TestClientSetsUpTransport(t *testing.T) {
	client, _ := NewClient(ClientOptions{Dsn: "https://foobar@ingest.sentry.io/32"})
	require.IsType(t, &HTTPTransport{}, client.Transport)

	client, _ = NewClient(ClientOptions{})
	require.IsType(t, &noopTransport{}, client.Transport)
}
