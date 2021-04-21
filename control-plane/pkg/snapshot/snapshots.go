package snapshot

import "github.com/envoyproxy/go-control-plane/pkg/cache/v3"

type UpstreamHost string

type Node struct {
	ListenerPort uint32
	UpstreamHost string
	UpstreamPort uint32
}

type Config struct {
	Version uint32
	Nodes   map[UpstreamHost]*Node
}

func Generate(config Config) (*cache.Snapshot, error) {
	snapshot := generateSnapshot(config)
	if err := snapshot.Consistent(); err != nil {
		return nil, err
	}
	return &snapshot, nil
}
