# Promsentry

Bring your Prometheus data to Sentry via [DDM](https://develop.sentry.dev/delightful-developer-metrics/).

Promsentry acts as a [Prometheus remote writer](https://prometheus.io/docs/concepts/remote_write_spec/) receiver (or a
remote write server if you say), that ingest the data and push it into Sentry's DDM, basically just converting
Prometheus' protobuf format into Sentry's accepted StatsD format.

## Installation

As of now, I don't provide any other way than to build the binary yourself.

1. Make sure you installed Go (minimum version of 1.21)
2. Run `go build -ldflags="-s -w" -o promsentry ./cmd`
3. Copy the `promsentry` binary into somewhere, for `/usr/local/bin/` you can do `sudo install promsentry /usr/local/bin/promsentry`.

## Running the application

Simply just invoke `promsentry` binary with some environment variables or configuration file. On your Prometheus configuration,
points the remote write target to the `{LISTEN_ADDRESS}/api/v1/write`

Here are some instructionS if you want to try around with local Prometheus Docker image.

### Without Sentry instance

1. Start local Prometheus server, simply by doing `docker compose up -d`. It will spawn a local Prometheus server on
   127.0.0.1:9090.
2. Then run the promsentry binary using `go run ./cmd`. It will start the server at `127.0.0.1:3000`.
3. See the debug log on your terminal.

### With real Sentry instance

1. Start local Prometheus server like the instruction above.
2. Set `SENTRY_DSN` on your environment variable to a valid Sentry DSN.
3. Then run the promsentry binary using `go run ./cmd`. It will start the server at `127.0.0.1:3000`.
4. Make sure the debug log on your terminal correctly sends some data.
5. Observe everything on your Sentry dashboard.

## Configuration

The program accepts 2 kinds of configuration:
1. Configuration file (JSON or YAML format)
2. Environment variables

### Configuration File

Start `promsentry` using `--config-file=./path/to/config.json` or `CONFIG_FILE=./path/to/config.yml promsentry`.

Refer to the schema and example values below. Every field is optional.

```jsonc
{
    "listen_address": "127.0.0.1:3000",
    "sentry_dsn": "https://xxxxxx@o123456.ingest.sentry.io/123456",
    "tls": {
        "certificate_authority_path": "./path/to/ca.pem",
        "server_certificate_path": "./path/to/cert.pem",
        "server_key_path": "./path/to/key.pem",
        "client_authentication_type": "VerifyClientCertIfGiven"
    },
    "debug": false
}

```yaml
listen_address: "127.0.0.1:3000"
sentry_dsn: "https://xxxxxx@o123456.ingest.sentry.io/123456"
tls:
    certificate_authority_path: "./path/to/ca.pem",
    server_certificate_path: "./path/to/cert.pem",
    server_key_path: "./path/to/key.pem",
    client_authentication_type: "VerifyClientCertIfGiven"
debug: false
```

### Environment variables

* `LISTEN_ADDRESS`
* `TLS_CERTIFICATE_AUTHORITY_PATH`
* `TLS_SERVER_CERTIFICATE_PATH`
* `TLS_SERVER_KEY_PATH`
* `TLS_CLIENT_AUTHENTICATION_TYPE`
* `SENTRY_DSN`
* `DEBUG`
