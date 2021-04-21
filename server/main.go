package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var (
	cfg  *Config
	port uint
)

func init() {
	cfg = configFromEnv()
	rand.Seed(time.Now().UTC().UnixNano())
	port = uint(rand.Intn(9999))
}

func main() {
	timeout := 2 * time.Second
	etcd, err := NewEtcd(cfg.EtcdEndpoint, timeout)
	if err != nil {
		log.Panic(err)
	}
	defer etcd.Close()

	if err := etcd.Register(
		cfg.Node,
	).WithValue(EtcdRegisterValue{
		ListenerPort: uint32(cfg.Listen),
		UpstreamHost: cfg.Node,
		UpstreamPort: uint32(port),
	}).WithTimeout(
		timeout,
	).Exec(); err != nil {
		log.Panic(err)
	}

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		log.Println("Server received request")
	})
	log.Printf("Server running on port: %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil))
}
