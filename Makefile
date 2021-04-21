build: cp.build sv.build

up: cp.build
	@docker-compose up

down:
	@docker-compose down

# control-plane
cp.build:
	@cd control-plane && go build

cp.install-air:
	@GO111MODULE=off go get -u github.com/cosmtrek/air

cp.serve:
	@cd control-plane && air -c air.toml

cp.dev: cp.install-air cp.serve

# server
sv.build:
	@cd server && go build

sv.launch:
ifeq ($(shell [ -z $(node) ] || [ -z $(listen) ] && echo true),true)
	@echo "Error: 'node' & 'listen' must be set: make node=node0 listen=80 sv.launch"
	@exit 1
endif
	@docker run --rm -it \
		-v $(shell pwd)/server/server:/go/bin/server \
		--network=service-discovery-exp_sd-cluster \
		--network-alias=$(node) \
		-e NODE=$(node) \
		-e LISTEN=$(listen) \
		-e ETCD_ENDPOINT=etcd:2379 \
		golang:1.16-stretch server

sv.run: sv.build sv.launch