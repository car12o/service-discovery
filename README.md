# service-discovery-experiment

This project is a get hands dirty experiment around service discovery in order to understand exactly how everything is pulled, cached, and served in a service mesh.

## Stack
- **data-plane** - [envoy](https://www.envoyproxy.io/)
- **control-plane** - [go-control-plane](https://github.com/envoyproxy/go-control-plane)
- **service-registry** - [etcd](https://etcd.io/)
- **service** - simple [go](https://golang.org/) server

## Architecture
![image](/architecture.svg)

## Getting started


### Requirements
  - [Docker](https://docs.docker.com/engine/install/)
  - [Docker-compose](https://docs.docker.com/compose/install/)

### Start cluster
```bash
# This command bootstraps a cluster ready to provide service discover.
# Starting The following components:
# data-plane, control-plane & service-registry.
make up
```

### Run service
```bash
# This command starts a service that registers himself on the cluster.
make sv.run

# or specify node ID and exposed port
make node=node0 listen=80 sv.run
```

### Request service
```bash
# Service will log the received request.
curl localhost:80
# 2021-04-27 22:22:44.644659 I | Server received request
```
