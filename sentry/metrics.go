package sentry

// Metric provide a.. type alias for []byte.
// You should provide a serialized statsd format using the statsd package.
// For multiple metric entries, please respect the new lines (you should provide the new lines).
type Metric []byte
