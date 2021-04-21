package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"go.etcd.io/etcd/clientv3"
)

type Etcd struct {
	*clientv3.Client
}

type EtcdRegister struct {
	etcd   *Etcd
	ctx    context.Context
	cancel context.CancelFunc
	key    string
	value  EtcdRegisterValue
}

type EtcdRegisterValue struct {
	ListenerPort uint32
	UpstreamHost string
	UpstreamPort uint32
}

func NewEtcd(endpoint string, timeout time.Duration) (*Etcd, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{endpoint},
		DialTimeout: timeout,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	_, err = cli.Status(ctx, endpoint)
	cancel()
	if err != nil {
		return nil, err
	}

	return &Etcd{cli}, nil
}

func (e *Etcd) Register(key string) *EtcdRegister {
	return &EtcdRegister{
		etcd: e,
		ctx:  context.Background(),
		key:  key,
	}
}

func (er *EtcdRegister) WithValue(value EtcdRegisterValue) *EtcdRegister {
	er.value = value
	return er
}

func (er *EtcdRegister) WithTimeout(timeout time.Duration) *EtcdRegister {
	ctx, cancel := context.WithTimeout(er.ctx, timeout)
	er.ctx = ctx
	er.cancel = cancel
	return er
}

func (er *EtcdRegister) Exec() error {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(er.value); err != nil {
		return err
	}

	_, err := er.etcd.Put(er.ctx, fmt.Sprintf("service-%s", er.key), buf.String())
	if er.cancel != nil {
		er.cancel()
	}
	if err != nil {
		return err
	}

	return nil
}
