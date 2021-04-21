package etcd

import (
	"context"
	"time"

	"go.etcd.io/etcd/clientv3"
)

func NewClient(endpoint string, timeout time.Duration) (*clientv3.Client, error) {
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

	return cli, nil
}
