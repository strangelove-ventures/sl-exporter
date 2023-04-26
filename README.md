# sl-exporter

Strangelove Prometheus Exporter

Exposes `/metrics` endpoint with Prometheus formatted metrics. Currently, we are only interested in
metadata metrics like `cosmos_asset_exponent`.

## Usage

```shell
# Build
go build -o exporter main.go
# Run
./exporter --bind=0.0.0.0:9100 --config=config.yaml
```

## Metrics

- `cosmos_asset_exponent{chain, denom}`

  Enriches Relayer metrics (e.g. `cosmos_relayer_wallet_balance`) to dynamically create Grafana panels and alerts.
- 
## Local Development

Run this once after you've cloned the repo:
```shell
make setup
```

To continually live reload changes while developing:
```shell
make watch
```

## More use cases

- Chain current block height

  This is a blocker to creating an alert on reference height for fullnodes. Such metrics are normally exposed in JSON
  endpoints and can not be used in alerting.

- USD value of a cosmos asset to attach it to metrics

  This allows creating financial panels

---

## Docker

```shell
# Build
docker buildx build --platform linux/amd64 -t sl-exporter .
# Run
docker run --rm -p 9100:9100 -v $(pwd)/config.yaml:/config.yaml sl-exporter --bind=0.0.0.0:9100
```

> The docker image does not contain a config file. Mount one and if necessary point to it using `--config` flag.
> This is an [example config](./config.example.yaml)

## Release

Simply push any tag to this repository

# Todos

- Add testing
- Enable linting on CI
- Use https://github.com/spf13/viper to simplify loading config and maybe flags as well
- Use slog for structured logging: https://pkg.go.dev/golang.org/x/exp/slog
- Load config once and don't recreate the registry