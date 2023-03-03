# sl-exporter

Strangelove Prometheus Exporter

Exposes `/metrics` endpoint with Prometheus formatted metrics. Currently, we are only interested in
metadata metrics like `cosmos_asset_exponent`.

## Usage

```shell
./sl_exporter --bind=0.0.0.0:9100 --config=config.yaml
```

## Metrics

- `cosmos_asset_exponent{chain, denom}`

  Enriches Relayer metrics (e.g. `cosmos_relayer_wallet_balance`) to dynamically create Grafana panels and alerts.

## Todo

- [ ] Other Go stuff !?
- [ ] Dockerize
- [ ] Lint n format
- [ ] CI to crete release?

## More use cases

- Chain current block height

  This is a blocker to creating an alert on reference height for fullnodes. Such metrics are normally exposed in JSON
  endpoints and can not be used in alerting.

- USD value of a cosmos asset to attach it to metrics

  This allows creating financial panels
