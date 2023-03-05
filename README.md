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

## Todo

- [ ] Other Go stuff !?
- [ ] Lint n format

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
docker run --rm -p 9100:9100 sl-exporter --bind=0.0.0.0:9100
```

## Release

Simply push any tag to this repository
