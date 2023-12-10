# Promsentry

Bring your Prometheus data to Sentry via DDM.

If you see this, then it's a work in progress. But, if you want to try stuff around:

#### Without Sentry instance

1. Start local Prometheus server, simply by doing `docker compose up -d`. It will spawn a local Prometheus server on
   127.0.0.1:9090.
2. Then run the promsentry binary using `go run ./cmd`. It will start the server at `127.0.0.1:3000`.
3. See the debug log on your terminal.

#### With real Sentry instance

I haven't tried this out, but if you wanted to:

1. Start local Prometheus server like the instruction above.
2. Set `SENTRY_DSN` on your environment variable to a valid Sentry DSN.
3. Then run the promsentry binary using `go run ./cmd`. It will start the server at `127.0.0.1:3000`.
4. Make sure the debug log on your terminal correctly sends some data.
5. Observe everything on your Sentry dashboard.
