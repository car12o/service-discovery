package main

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Node         string
	Listen       uint
	EtcdEndpoint string
}

func configFromEnv() *Config {
	return &Config{
		Node:         getenvOr("NODE", "node0"),
		Listen:       getenvUintOr("LISTEN", 80),
		EtcdEndpoint: getenvOr("ETCD_ENDPOINT", "0.0.0.0:2379"),
	}
}

func getenvOr(env, defaultValue string) string {
	value := os.Getenv(env)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func getenvUintOr(env string, defaultValue uint) uint {
	value := getenvOr(env, strconv.FormatUint(uint64(defaultValue), 10))
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalln(err)
	}
	return uint(intValue)
}
