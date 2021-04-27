package main

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	NodeID       string
	Port         uint
	Debug        bool
	EtcdEndpoint string
}

func configFromEnv() *Config {
	return &Config{
		NodeID:       getenvOr("NODE_ID", "xds-node"),
		Port:         getenvUintOr("PORT", 5678),
		Debug:        getenvBoolOr("DEBUG", false),
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

func getenvBoolOr(env string, defaultValue bool) bool {
	value := getenvOr(env, strconv.FormatBool(defaultValue))
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		log.Fatalln(err)
	}
	return boolValue
}
