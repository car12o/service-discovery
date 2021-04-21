package main

import (
	"log"
	"time"

	"github.com/car12o/service-discovery-exp/control-plane/pkg/server"
	"github.com/car12o/service-discovery-exp/control-plane/pkg/services/etcd"
)

var (
	cfg *Config
)

func init() {
	cfg = configFromEnv()
}

func main() {
	etcd, err := etcd.NewClient(cfg.EtcdEndpoint, 2*time.Second)
	if err != nil {
		log.Panic(err)
	}
	defer etcd.Close()

	if err := server.New(
		cfg.NodeID,
		cfg.Port,
		etcd,
		cfg.Debug,
	).Run(); err != nil {
		log.Panic(err)
	}
}
