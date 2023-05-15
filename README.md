# sl-exporter

Strangelove Prometheus Exporter

Exposes `/metrics` endpoint with Prometheus formatted metrics. 

## Motivation

This exporter allows you to monitor blockchains and validators which interest you. Although the current focus is
cosmos chains, the exporter is designed to be chain agnostic. The primary focus is to expose actionable metrics 
for SREs. 

It can be run independently and does not need to run alongside an RPC or validator node.

This exporter does not export metrics for entire chains. For that, consider [the cosmos-exporter](https://github.com/solarlabsteam/cosmos-exporter).

## Usage

Almost all configuration is via the config file. See [example config](./config.example.yaml).

```shell
# Build
go build -o exporter main.go
# See usage
./exporter -h
# Example run
./exporter --bind=0.0.0.0:9100 --config=config.yaml
```

## Local Development

Quickstart:
```shell
make setup
go run .
# In another shell, see /metrics endpoint
curl localhost:9100/metrics
```

To continually live reload changes while developing:
```shell
make watch
```

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

Push tag to the remote. 

## Roadmap

- [ ] Financial exchange data to aid with validator economics
- [ ] Non-cosmos chains such as Ethereum balance monitoring