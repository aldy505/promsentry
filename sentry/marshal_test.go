package sentry

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

var (
	goReleaseDate = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	utcMinusTwo   = time.FixedZone("UTC-2", -2*60*60)
)

func TestMarshalJSON(t *testing.T) {
	tests := []struct {
		in  interface{}
		out string
	}{
		// TODO: eliminate empty struct fields from serialization of empty event.
		// Only *Event implements json.Marshaler.
		// {Event{}, `{"sdk":{},"user":{}}`},
		{&Event{}, `{"timestamp":"0001-01-01T00:00:00Z","sdk":{},"start_timestamp":"0001-01-01T00:00:00Z"}`},
		// Only *Breadcrumb implements json.Marshaler.
		// {Breadcrumb{}, `{}`},
		{&Breadcrumb{}, `{}`},
	}
	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			want := tt.out
			b, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			got := string(b)
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("JSON serialization mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func WriteGoldenFile(t *testing.T, path string, bytes []byte) {
	t.Helper()
	err := os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(path, bytes, 0666)
	if err != nil {
		t.Fatal(err)
	}
}

func ReadOrGenerateGoldenFile(t *testing.T, path string, bytes []byte) string {
	t.Helper()
	b, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		if *generate {
			WriteGoldenFile(t, path, bytes)
			return string(bytes)
		}
		t.Fatalf("Missing golden file. Run `go test -args -gen` to generate it.")
	case err != nil:
		t.Fatal(err)
	}
	return string(b)
}
